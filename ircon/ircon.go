// Package ircon maintains a connection to the Twitch IRC service.
package ircon

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"bnsvc.net/tmi/irc"
)

// DefaultServer is the default Twitch "IRC" server used by e.g. web chat
const DefaultServer = "wss://irc-ws.chat.twitch.tv/"

// Message aliases irc.Message for import convenience.
type Message = irc.Message

// A Handler receives events from IRCon.
type Handler interface {
	Connected()
	Disconnected(err error)
	Message(*irc.Message)
}

// An IRCon is an automatically reconnecting IRC connection.
type IRCon struct {
	server       string
	nick, passwd string

	ctx context.Context
	con *irc.IRC
	mu  sync.Mutex

	Handler Handler
}

// New creates a new IRCon with the given credentials.
func New(nick, passwd string) *IRCon {
	return &IRCon{
		nick:   nick,
		passwd: passwd,
	}
}

// Background runs the connection in a background goroutine until ctx is done.
func (i *IRCon) Background(ctx context.Context) {
	i.ctx = ctx
	go i.loop()
}

func (i *IRCon) loop() {
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
		delay = time.After(time.Second * 30)
		wait := i.establish()
		if wait != nil {
			select {
			case <-i.ctx.Done():
				i.con.Close()
				return
			case <-wait:
			}
		}
	}
}

func (i *IRCon) establish() chan struct{} {
	server := i.server
	if server == "" {
		server = DefaultServer
	}
	log.Println("Establishing connection to", server)
	con := irc.New(i.ctx)
	msgs, err := con.Connect(server)
	if err != nil {
		log.Println("Failed to connect to", server, "->", err)
		return nil
	}
	passwd := i.passwd
	nick := i.nick
	if nick == "" {
		// Anonymous login; cannot send messages(!)
		nick = "justinfan12345"
		passwd = "blah"
	}
	con.Send("CAP REQ :twitch.tv/tags twitch.tv/membership twitch.tv/commands")
	con.Send(fmt.Sprint("PASS ", passwd))
	con.Send(fmt.Sprint("NICK ", nick))
	con.Send(fmt.Sprint("USER ", nick, " 8 * :", nick))
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
			con.Send(fmt.Sprint("PONG :", msg.Trailer(0)))
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
