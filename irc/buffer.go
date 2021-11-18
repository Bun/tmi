package irc

import (
	"bytes"
)

type buffer struct {
	bytes.Buffer
}

func (b *buffer) Next(buf []byte) {
	b.Buffer.Write(buf)
}

func (b *buffer) Read() (*Message, error) {
	buf := b.Buffer.Bytes()
	n := bytes.IndexByte(buf, '\n')
	if n < 0 {
		return nil, nil
	}

	line := b.Buffer.Next(n + 1)
	line = bytes.TrimRight(line[:n], "\r")
	return ParseMessage(string(line)), nil
}
