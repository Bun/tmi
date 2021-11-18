package irc

import (
	"net"
	"time"
)

type netConn struct {
	conn net.Conn
	buffer
	readbuf [4096]byte
}

func (wc *netConn) Read() (*Message, error) {
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
		n, err := wc.conn.Read(wc.readbuf[:])
		if n > 0 {
			wc.buffer.Next(wc.readbuf[:n])
		}
		if err != nil {
			return nil, err
		}
	}
}

func (wc *netConn) Close() error {
	return wc.conn.Close()
}

func (wc *netConn) Send(message string) error {
	buf := safeMessage(message)
	wc.conn.SetWriteDeadline(time.Now().Add(deadline))
	_, err := wc.conn.Write(buf)
	return err
}
