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

// ews_sync is the shape of gws_std: net/http on every port, upgrade in a
// handler, and each echo written synchronously before the next read.
func main() {
	flag.Parse()
	addrs, err := config.GetFrameworkServerAddrs(config.EwsSync)
	if err != nil {
		logging.Fatalf("GetFrameworkBenchmarkAddrs(%v) failed: %v", config.EwsSync, err)
	}
	lns := startServers(addrs)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	for _, ln := range lns {
		ln.Close()
	}
}

func startServers(addrs []string) []net.Listener {
	lns := make([]net.Listener, 0, len(addrs))
	for _, addr := range addrs {
		mux := &http.ServeMux{}
		mux.HandleFunc("/ws", onWebsocket)
		mux.HandleFunc("/pid", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "%d", os.Getpid()) })
		ln, err := frameworks.Listen("tcp", addr)
		if err != nil {
			logging.Fatalf("Listen failed: %v", err)
		}
		lns = append(lns, ln)
		go func() { logging.Printf("server exit: %v", (&http.Server{Handler: mux}).Serve(ln)) }()
	}
	return lns
}

func onWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, res, err := transport.Upgrade(w, r, handshake.Options{})
	if err != nil {
		return
	}
	frameworks.SetNoDelay(conn, *nodelay)
	// Return from the handler and read on a fresh goroutine, as gws_std does:
	// net/http then frees the request state it would otherwise keep alive for
	// the life of the connection.
	go echo(conn, res)
}

func echo(conn net.Conn, res handshake.Result) {
	defer conn.Close()
	c, err := ws.NewConn(conn, ws.Config{Role: ws.Server, Compression: res.Compression})
	if err != nil {
		return
	}
	for {
		op, payload, err := c.ReadMessage()
		if err != nil {
			return
		}
		if err := c.Write(op, payload); err != nil {
			return
		}
	}
}
