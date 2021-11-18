package irc

import (
	"context"
	"crypto/tls"
)

type tlsTransport struct {
	addr   string
	dialer *tls.Dialer
	certs  []tls.Certificate
}

func (wt tlsTransport) Dial(ctx context.Context) (Conn, error) {
	config := &tls.Config{
		//ServerName: snParts[0],
		Certificates: wt.certs,
	}

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
