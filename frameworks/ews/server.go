package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"

	"go-websocket-benchmark/config"
	"go-websocket-benchmark/frameworks"
	"go-websocket-benchmark/logging"

	"github.com/33TU/ews/handshake"
	"github.com/33TU/ews/transport"
	"github.com/33TU/ews/ws"
)

var (
	nodelay = flag.Bool("nodelay", true, `tcp nodelay`)
	_       = flag.Int("b", 1024, `read buffer size`)
	_       = flag.Int("mrb", 4096, `max read buffer size`)
	_       = flag.Int64("m", 1024*1024*1024*2, `memory limit`)
	_       = flag.Int("mb", 10000, `max blocking online num, e.g. 10000`)
)

func main() {
	flag.Parse()
	addrs, err := config.GetFrameworkServerAddrs(config.Ews)
	if err != nil {
		logging.Fatalf("GetFrameworkBenchmarkAddrs(%v) failed: %v", config.Ews, err)
	}
	lns := startServers(addrs)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	for _, ln := range lns {
		ln.Close()
	}
}

// startServers runs a transport.Server, ews's accept loop without net/http,
// on every port but the last. The benchmark reads the pid over HTTP from the
// last port, so that one is served by net/http with the WebSocket path
// handed to transport.Upgrade.
func startServers(addrs []string) []net.Listener {
	server := &transport.Server{
		Handler: func(conn net.Conn, res handshake.Result, _ *transport.Request) {
			frameworks.SetNoDelay(conn, *nodelay)
			echo(conn, res)
		},
	}
	lns := make([]net.Listener, 0, len(addrs))
	for i, addr := range addrs {
		ln, err := frameworks.Listen("tcp", addr)
		if err != nil {
			logging.Fatalf("Listen failed: %v", err)
		}
		lns = append(lns, ln)
		if i == len(addrs)-1 {
			mux := &http.ServeMux{}
			mux.HandleFunc("/pid", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "%d", os.Getpid()) })
			mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
				conn, res, err := transport.Upgrade(w, r, handshake.Options{})
				if err != nil {
					return
				}
				defer conn.Close()
				frameworks.SetNoDelay(conn, *nodelay)
				echo(conn, res)
			})
			go func() { logging.Printf("server exit: %v", (&http.Server{Handler: mux}).Serve(ln)) }()
			continue
		}
		go func() { logging.Printf("server exit: %v", server.Serve(ln)) }()
	}
	return lns
}

// echo relays every message back through the connection's queue, the ews
// equivalent of gws's WriteAsync: Send returns before the write and a writer
// goroutine runs only while the queue is nonempty.
func echo(conn net.Conn, res handshake.Result) {
	c, err := ws.NewConn(conn, ws.Config{Role: ws.Server, Compression: res.Compression})
	if err != nil {
		return
	}
	q := c.NewQueue(0)
	for {
		op, payload, err := c.ReadMessage()
		if err != nil {
			return
		}
		if err := q.Send(op, payload); err != nil {
			return
		}
	}
}
