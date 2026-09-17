// gencert writes a self-signed certificate and key for 127.0.0.1 and
// localhost into the directory given as the first argument, for the -tls
// mode of the servers and the client. The client skips verification, so
// the certificate only has to exist.
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

func main() {
	dir := "./output"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "go-websocket-benchmark"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:     []string{"localhost"},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		log.Fatal(err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		log.Fatal(err)
	}
	write := func(name, typ string, b []byte) {
		f, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		if err := pem.Encode(f, &pem.Block{Type: typ, Bytes: b}); err != nil {
			log.Fatal(err)
		}
	}
	write("cert.pem", "CERTIFICATE", der)
	write("key.pem", "EC PRIVATE KEY", keyDER)
}
