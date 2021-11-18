// Package irc is an IRC-over-websocket client implementation.
package irc

import (
	"context"
	"errors"
	"net"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// deadline is used to set a maximum read/write timeout. This shouldn't be
// violated on a working connection, due to in-protocol PING/PONG and normal
// activity.
//
// The PING interval appears to be around 4 to 5 minutes.
const deadline = 6 * 60 * time.Second

// A Conn is an IRC connection.
type Conn interface {
	Read() (*Message, error)
	Send(string) error
	Close() error
}

// Dialer is a preconfigured IRC connector.
type Dialer interface {
	// Dial connects with a given context. The context does not (only) govern
	// the Dial itself, but also the established connection.
	Dial(context.Context) (Conn, error)
}

// New creates a new Dialer from a URL.
func New(addr string) (Dialer, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "ws":
		return websocketTransport{
			addr:   addr,
			dialer: websocket.DefaultDialer,
		}, nil
	case "wss":
		return websocketTransport{
			addr:   addr,
			dialer: websocket.DefaultDialer,
		}, nil
	case "ircs":
		return tlsTransport{
			addr: defaultPort(u.Host, "6697"),
		}, nil
	}

	return nil, errors.New("Unsupported schema")
}

func defaultPort(host, port string) string {
	_, p, err := net.SplitHostPort(host)
	if err != nil || p == "" {
		return host + ":" + port
	}
	return host
}
