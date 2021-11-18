package irc

import (
	"context"
	"crypto/tls"
	"time"
)

type TLSOpt interface {
	apply(*tlsTransport)
}

// TLSConfig is a TLSOpt for specifying a tls.Config.
type TLSConfig struct{ *tls.Config }

func (cfg TLSConfig) apply(t *tlsTransport) {
	t.config = cfg.Config
}

// NewTLS is an extended initializer for TLS-based IRC.
func NewTLS(addr string, options ...TLSOpt) Dialer {
	transport := tlsTransport{
		addr: defaultPort(addr, "6697"),
	}
	for _, opt := range options {
		opt.apply(&transport)
	}
	return transport
}

type tlsTransport struct {
	addr   string
	dialer *tls.Dialer
	config *tls.Config
}

func (wt tlsTransport) Dial(ctx context.Context) (Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	config := wt.config
	nc, err := (&tls.Dialer{
		Config: config,
	}).DialContext(ctx, "tcp", wt.addr)
	if err != nil {
		return nil, err
	}
	c := nc.(*tls.Conn)
	if err := c.HandshakeContext(ctx); err != nil {
		nc.Close()
		return nil, err
	}
	return &netConn{conn: c}, nil
}
