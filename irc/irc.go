// Package irc is an IRC-over-websocket client implementation.
package irc

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// deadline is used to set a maximum read/write timeout. This shouldn't be
// violated on a working connection, due to in-protocol PING/PONG and normal
// activity.
//
// The PING interval appears to be around 4 to 5 minutes.
const deadline = 6 * 60 * time.Second

// IRC is an IRC connection.
type IRC struct {
	ctx context.Context

	address string
	con     *websocket.Conn
	lock    sync.Mutex
	err     error

	Dialer   *websocket.Dialer
	Messages chan *Message // recv
	Outgoing chan string   // send
}

// New creates a new IRC instance for a single IRC session.
func New(ctx context.Context) *IRC {
	return &IRC{
		ctx:      ctx,
		Outgoing: make(chan string, 4),
	}
}

// Connect to an IRC server.
func (irc *IRC) Connect(address string) (chan *Message, error) {
	irc.address = address

	dialer := irc.Dialer
	if dialer == nil {
		dialer = websocket.DefaultDialer
	}
	c, _, err := dialer.Dial(address, nil)
	if err != nil {
		return nil, err
	}

	irc.con = c
	go irc.sender()

	irc.Messages = irc.reader()
	return irc.Messages, nil
}

// Err returns the first error that occurred, if any.
func (irc *IRC) Err() error {
	irc.lock.Lock()
	err := irc.err
	irc.lock.Unlock()
	return err
}

func (irc *IRC) sender() {
	// TODO: while nice to have, Replacer has some interesting memory
	// allocation behavior.
	guard := strings.NewReplacer("\x00", " ", "\r", " ", "\n", " ")
	og := irc.Outgoing

	for {
		message, ok := <-og
		if !ok {
			// No more messages to send
			return
		}

		message = guard.Replace(message)
		irc.con.SetWriteDeadline(time.Now().Add(deadline))
		err := irc.con.WriteMessage(websocket.TextMessage, []byte(message+"\r\n"))
		if err != nil {
			irc.closeWithErr(err)
			break
		}
	}

	for range og {
		// We have to discard messages until Close() to ensure calls to Send()
		// don't hang
	}
}

func (irc *IRC) reader() chan *Message {
	incoming := make(chan *Message, 4)
	go func() {
		defer close(incoming)
		for {
			// TODO
			irc.con.SetReadDeadline(time.Now().Add(deadline))
			_, message, err := irc.con.ReadMessage()
			if err != nil {
				// XXX: could be io.EOF
				irc.closeWithErr(err)
				return
			}
			for _, line := range bytes.Split(message, []byte("\n")) {
				line = bytes.TrimRight(line, "\r")
				if len(line) == 0 {
					continue
				}
				m := ParseMessage(string(line))
				if m != nil {
					incoming <- m
				}
			}
		}
	}()
	return incoming
}

// Send queues a message to be sent to the IRC server, blocking if the queue is
// full.
func (irc *IRC) Send(message string) {
	// TODO: ensure this does not block forever on error-close
	irc.Outgoing <- message
}

func (irc *IRC) closeWithErr(err error) {
	irc.lock.Lock()
	if irc.err == nil {
		irc.err = err
	}
	con := irc.con
	irc.lock.Unlock()
	if con != nil {
		con.Close()
	}
}

// Close terminates the IRC connection and stops the sender goroutine.
func (irc *IRC) Close() error {
	irc.lock.Lock()
	defer irc.lock.Unlock()

	if irc.Outgoing != nil {
		close(irc.Outgoing)
		irc.Outgoing = nil
	}

	if irc.con != nil {
		err := irc.con.Close()
		irc.con = nil
		return err
	}

	return nil
}
