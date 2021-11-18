package irc

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
)

type websocketTransport struct {
	addr   string
	dialer *websocket.Dialer
}

func (wt websocketTransport) Dial(ctx context.Context) (Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	c, _, err := wt.dialer.DialContext(ctx, wt.addr, nil)
	if err != nil {
		return nil, err
	}
	return &websocketConn{conn: c}, nil
}

type websocketConn struct {
	conn *websocket.Conn
	buffer
}

func (wc *websocketConn) Read() (*Message, error) {
	for first := true; ; first = false {
		msg, err := wc.buffer.Read()
		if err != nil {
			return nil, err
		} else if msg != nil {
			return msg, nil
		}

		if first {
			// TODO
			wc.conn.SetReadDeadline(time.Now().Add(deadline))
		}
		_, message, err := wc.conn.ReadMessage()
		if err != nil {
			return nil, err
		}

		wc.buffer.Next(message)
	}
}

func (wc *websocketConn) Close() error {
	return wc.conn.Close()
}

func (wc *websocketConn) Send(message string) error {
	buf := safeMessage(message)
	wc.conn.SetWriteDeadline(time.Now().Add(deadline))
	return wc.conn.WriteMessage(websocket.TextMessage, buf)
}
