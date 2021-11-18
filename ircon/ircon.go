// Package ircon maintains a connection to the Twitch IRC service.
package ircon

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"raccatta.cc/tmi/irc"
)

const (
	// DefaultIRCServer is the default Twitch IRC server
	DefaultIRCServer = "ircs://irc.chat.twitch.tv:6697/"

	// DefaultWebChat is the default Twitch "IRC" server used by e.g. web chat
	DefaultWebChat = "wss://irc-ws.chat.twitch.tv/"

	// DefaultServer is the default server.
	DefaultServer = DefaultWebChat
)

type (
	// Sender abstraction for Handshaker
	Sender interface {
		Send(msg string) error
	}
	Handshaker interface {
		Handshake(s Sender) error
	}
)

// DefaultCaps is the default set of capabilities. The twitch.tv/membership
// capability is omitted for performance reasons and its general lack of
// usefulness in most scenarios.
const DefaultCaps = "twitch.tv/tags twitch.tv/commands"

// Message aliases irc.Message for import convenience.
type Message = irc.Message

// A Handler receives events from IRCon.
type Handler interface {
	// Connected is called when a connection is fully established and the
	// connection is ready to send messages.
	Connected()

	// Disconnected is called when an active connection ends, potentially due
	// to an error. It will also be called when establishing a connection
	// fails, before it was fully Connected.
	Disconnected(err error)

	// Message is called for every incoming message. Note that it can be called
	// before Connected or after Disconnected are called.
	Message(*irc.Message)
}

// An IRCon is an automatically reconnecting IRC connection.
type IRCon struct {
	// Server allows connecting to a different TMI server.
	Server string

	handshaker Handshaker
	con        *conn
	mu         sync.Mutex
}

// New creates a new IRCon with the given credentials.
func New(h Handshaker) *IRCon {
	return &IRCon{
		handshaker: h,
	}
}

// Run maintains a connection to the IRC server until the context is done.
func (i *IRCon) Run(ctx context.Context, h Handler) {
	i.loop(ctx, h)
}

func (i *IRCon) loop(ctx context.Context, h Handler) {
	cd := newBackoff(15, 300)
	delay := time.After(0)
	for {
		i.mu.Lock()
		i.con = nil
		i.mu.Unlock()
		select {
		case <-ctx.Done():
			return
		case <-delay:
		}
		cd.Now()
		i.session(ctx, h)
		delay = time.After(cd.Delay())
	}
}

func (i *IRCon) session(ctx context.Context, h Handler) {
	server := i.Server
	if server == "" {
		server = DefaultServer
	}
	factory, err := irc.New(ctx, server)
	if err != nil {
		h.Disconnected(err)
		return
	}

	con, err := factory.Connect()
	if err != nil {
		h.Disconnected(err)
		return
	}
	c := &conn{
		Conn: con,
		h:    h,
	}
	i.mu.Lock()
	i.con = c
	i.mu.Unlock()
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		defer con.Close()
		for {
			msg, err := con.Read()
			if err != nil {
				c.closeWithErr(fmt.Errorf("Read failed: %w", err))
				return
			}

			if msg.Command == "PING" {
				con.Send("PONG :" + msg.Trailer(0))
			}

			// Call should not block
			// Call should implement error handling
			h.Message(msg)
		}
	}()
	if err := i.handshaker.Handshake(con); err == nil {
		h.Connected()
	}
	<-wait
}

// Send sends a message to the currently active IRC connection. If there is no
// active connection, the message is lost.
//
// TODO: typically we want to associate message sending with a specific
// connection instance, and not confuse it during reconnects.
// TODO: make a variant that can wait for a succesful connection.
func (i *IRCon) Send(s string) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.con == nil {
		return ErrNotConnected
	}
	err := i.con.Send(s)
	if err != nil {
		i.con.closeWithErr(fmt.Errorf("Send failed: %w", err))
	}
	return err
}

var ErrNotConnected = errors.New("Not connected")

type conn struct {
	irc.Conn
	disconnected sync.Once
	h            Handler
}

func (c *conn) closeWithErr(err error) {
	c.disconnected.Do(func() {
		c.h.Disconnected(err)
	})
	c.Close()
}
