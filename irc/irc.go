// Package irc is an IRC-over-websocket client implementation.
package irc

import (
	"context"
	"errors"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// deadline is used to set a maximum read/write timeout. This shouldn't be
// violated on a working connection, due to in-protocol PING/PONG and normal
// activity.
//
// The PING interval appears to be around 4 to 5 minutes.
const deadline = 6 * 60 * time.Second

// TODO: while nice to have, Replacer has some interesting memory
// allocation behavior.
var messageGuard = strings.NewReplacer("\x00", " ", "\r", " ", "\n", " ")

type Conn interface {
	Read() (*Message, error)
	Send(string) error
	Close() error
}

// IRC is an IRC connection.
type IRC struct {
	ctx  context.Context
	dial func(context.Context) (Conn, error)
}

// New creates a new IRC instance for a single IRC session.
func New(ctx context.Context, address string) (*IRC, error) {
	i := &IRC{ctx: ctx}
	return i, i.setupTransport(address)
}

func (irc *IRC) setupTransport(addr string) error {
	u, err := url.Parse(addr)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "ws":
		irc.dial = websocketTransport{
			addr:   addr,
			dialer: websocket.DefaultDialer,
		}.Dial
		return nil
	case "wss":
		irc.dial = websocketTransport{
			addr:   addr,
			dialer: websocket.DefaultDialer,
		}.Dial
		return nil
	case "irc":
		return errors.New("Insecure IRC not supported")
	case "ircs":
		irc.dial = tlsTransport{
			addr: defaultPort(u.Host, "6697"),
		}.Dial
		return nil
	}

	return errors.New("Unsupported schema")
}

func defaultPort(host, port string) string {
	_, p, err := net.SplitHostPort(host)
	if err != nil || p == "" {
		return host + ":" + port
	}
	return host
}

// Connect to the IRC server.
func (irc *IRC) Connect( /*context*/ ) (Conn, error) {
	ctx, cancel := context.WithTimeout(irc.ctx, time.Second*30)
	defer cancel()
	return irc.dial(ctx)
}
