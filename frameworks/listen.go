package frameworks

import (
	"crypto/tls"
	"flag"
	"net"
	"sync"

	"github.com/libp2p/go-reuseport"
)

var (
	reuse    = flag.Bool("reuseport", false, `reuse port`)
	useTLS   = flag.Bool("tls", false, `serve WebSocket ports over TLS with -cert and -key`)
	certFile = flag.String("cert", "./output/cert.pem", `certificate for -tls`)
	keyFile  = flag.String("key", "./output/key.pem", `key for -tls`)

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
func ListenPlain(network, addr string) (net.Listener, error) {
	if *reuse {
		return reuseport.Listen(network, addr)
	}
	return net.Listen(network, addr)
}
