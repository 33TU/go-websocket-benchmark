package frameworks

import (
	"context"
	"crypto/tls"
	"flag"
	"net"
	"sync"
	"syscall"

	"github.com/libp2p/go-reuseport"
)

var (
	reuse    = flag.Bool("reuseport", false, `reuse port`)
	useTLS   = flag.Bool("tls", false, `serve WebSocket ports over TLS with -cert and -key`)
	certFile = flag.String("cert", "./output/cert.pem", `certificate for -tls`)
	keyFile  = flag.String("key", "./output/key.pem", `key for -tls`)
	sndbuf   = flag.Int("sndbuf", 0, `SO_SNDBUF on every WebSocket listener, 0 to leave the kernel default`)
	rcvbuf   = flag.Int("rcvbuf", 0, `SO_RCVBUF on every WebSocket listener, 0 to leave the kernel default`)

	tlsOnce   sync.Once
	tlsConfig *tls.Config
)

// Listen opens a WebSocket port, wrapped in TLS when -tls is set.
func Listen(network, addr string) (net.Listener, error) {
	ln, err := ListenPlain(network, addr)
	if err != nil || !*useTLS {
		return ln, err
	}
	tlsOnce.Do(func() {
		cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
		if err != nil {
			panic(err)
		}
		tlsConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	})
	return tls.NewListener(ln, tlsConfig), nil
}

// ListenPlain opens a port without TLS regardless of -tls, for the pid
// endpoint the client reads over plain HTTP.
//
// With -sndbuf or -rcvbuf the listening socket carries an explicit
// SO_SNDBUF or SO_RCVBUF, which every accepted connection inherits, so the
// sizes apply to a whole server without touching its framework code. Every
// framework that listens through here gets them on the same terms;
// `-reuseport` takes another path and ignores them.
//
// The echo test carries the payload in both directions, so both sizes are
// in the path and are usually set together. Setting either one also turns
// off the kernel's autotuning for it, which is the point: autotuning grows
// the buffer into the megabytes, and a large message then goes to memory
// and comes back instead of staying in cache.
func ListenPlain(network, addr string) (net.Listener, error) {
	if *reuse {
		return reuseport.Listen(network, addr)
	}
	if *sndbuf <= 0 && *rcvbuf <= 0 {
		return net.Listen(network, addr)
	}
	lc := net.ListenConfig{
		Control: func(_, _ string, c syscall.RawConn) error {
			var serr error
			if err := c.Control(func(fd uintptr) {
				if *sndbuf > 0 {
					if serr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_SNDBUF, *sndbuf); serr != nil {
						return
					}
				}
				if *rcvbuf > 0 {
					serr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF, *rcvbuf)
				}
			}); err != nil {
				return err
			}
			return serr
		},
	}
	return lc.Listen(context.Background(), network, addr)
}
