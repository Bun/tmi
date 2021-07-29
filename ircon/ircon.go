// Package ircon maintains a connection to the Twitch IRC service.
package ircon

import (
	"context"
	"sync"
	"time"

	"bnsvc.net/tmi/irc"
)

// DefaultServer is the default Twitch "IRC" server used by e.g. web chat
const DefaultServer = "wss://irc-ws.chat.twitch.tv/"

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
	// Caps contains the set of capabilities that are requested on connect.
	// Advanced users can specify their own set.
	Caps string

	server       string
	nick, passwd string

	ctx context.Context
	con *irc.IRC
	mu  sync.Mutex

	Handler Handler
}

// New creates a new IRCon with the given credentials.
func New(nick, passwd string) *IRCon {
	if nick == "" {
		// Default anonymous login; cannot send messages(!)
		nick = "justinfan12345"
		passwd = "blah"
	}
	return &IRCon{
		Caps:   DefaultCaps,
		nick:   nick,
		passwd: passwd,
	}
}

// Nick returns the username or anonymous nickname used for this connection.
func (i *IRCon) Nick() string {
	return i.nick
}

// Background runs the connection in a background goroutine until ctx is done.
func (i *IRCon) Background(ctx context.Context) {
	i.ctx = ctx
	go i.loop()
}

func (i *IRCon) loop() {
	cd := newBackoff(15, 300)
	delay := time.After(0)
	for {
		i.mu.Lock()
		i.con = nil
		i.mu.Unlock()
		select {
		case <-i.ctx.Done():
			return
		case <-delay:
		}
		cd.Now()
		wait := i.establish()
		// Wait until connection is lost
		if wait != nil {
			select {
			case <-i.ctx.Done():
				i.con.Close()
				return
			case <-wait:
			}
		}
		delay = time.After(cd.Delay())
	}
}

func (i *IRCon) establish() chan struct{} {
	server := i.server
	if server == "" {
		server = DefaultServer
	}
	con := irc.New(i.ctx)
	msgs, err := con.Connect(server)
	if err != nil {
		i.Handler.Disconnected(err)
		return nil
	}
	passwd := i.passwd
	nick := i.nick
	if i.Caps != "" {
		con.Send("CAP REQ :" + i.Caps)
	}
	con.Send("PASS " + passwd)
	con.Send("NICK " + nick)
	con.Send("USER " + nick + " 8 * :" + nick)
	wait := make(chan struct{})
	i.mu.Lock()
	i.con = con
	i.mu.Unlock()
	i.Handler.Connected()
	go func() {
		defer func() {
			i.Handler.Disconnected(con.Err()) // TODO: validate order of events
		}()
		defer close(wait)
		defer con.Close()
		i.dispatch(con, msgs)
	}()
	return wait
}

func (i *IRCon) dispatch(con *irc.IRC, msgs chan *irc.Message) {
	for msg := range msgs {
		switch msg.Command {
		case "PING":
			con.Send("PONG :" + msg.Trailer(0))
		}
		// Call should not block
		// Call should implement error handling
		i.Handler.Message(msg)
	}
}

// Send sends a message to the currently active IRC connection. If there is no
// active connection, the message is lost.
func (i *IRCon) Send(s string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.con != nil {
		i.con.Send(s)
	}
}
