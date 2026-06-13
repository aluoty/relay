package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
)

type ServerConfig struct {
	CertFile string
	KeyFile  string
}

type ClientConfig struct {
	Enabled  bool
	CAFile   string
	Insecure bool
}

func (c ServerConfig) Enabled() bool {
	return c.CertFile != "" || c.KeyFile != ""
}

func (c ServerConfig) Listen(addr string) (net.Listener, error) {
	if !c.Enabled() {
		return net.Listen("tcp", addr)
	}
	if c.CertFile == "" || c.KeyFile == "" {
		return nil, fmt.Errorf("both --tls-cert and --tls-key are required for TLS")
	}
	cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load tls cert: %w", err)
	}
	return tls.Listen("tcp", addr, &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
	})
}

func (c ClientConfig) Dial(addr string) (net.Conn, error) {
	if !c.Enabled {
		return net.Dial("tcp", addr)
	}

	cfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if c.Insecure {
		cfg.InsecureSkipVerify = true
	}
	if c.CAFile != "" {
		data, err := os.ReadFile(c.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read tls ca: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(data) {
			return nil, fmt.Errorf("parse tls ca: %s", c.CAFile)
		}
		cfg.RootCAs = pool
	}
	return tls.Dial("tcp", addr, cfg)
}
