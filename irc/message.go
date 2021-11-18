package irc

import (
	"fmt"
	"strings"
)

type Message struct {
	Tags       map[string]string `json:"tags"`
	Source     string            `json:"source"`
	Command    string            `json:"command"`
	Args       []string          `json:"args"`
	HasTrailer bool              `json:"trailer,omitempty"` // Useful to know.. usually

	raw string
}

// Trailer returns the argument at position i, even if it is not strictly the
// final argument.
func (m *Message) Trailer(i int) string {
	if l := len(m.Args); l == (i + 1) {
		return m.Args[i]
	} else if l >= i {
		// Should never really happen
		return strings.Join(m.Args[i:], " ")
	}
	return ""
}

// Arg returns the given argument, even if it is not set.
func (m *Message) Arg(i int) string {
	if i < len(m.Args) {
		return m.Args[i]
	}
	return ""
}

func (m *Message) Raw() string {
	return m.raw
}

var unescapeTag = strings.NewReplacer("\\:", ";", "\\s", " ", "\\r", "\r", "\\n", "\n", "\\\\", "\\")

func parseTags(s string) map[string]string {
	tags := strings.Split(s, ";")
	t := make(map[string]string)

	for _, tag := range tags {
		split := strings.IndexByte(tag, '=')
		if split < 0 {
			t[tag] = ""
		} else {
			key := tag[:split]
			value := tag[split+1:]
			if strings.Contains(value, "\\") {
				t[key] = unescapeTag.Replace(value)
			} else {
				t[key] = value
			}
		}
	}

	return t
}

func (m *Message) String() string {
	if len(m.Tags) > 0 {
		return fmt.Sprintf("Message(tags=%v, from=%s, %s, args=%v)", m.Tags, m.Source, m.Command, m.Args)
	}
	return fmt.Sprintf("Message(from=%s, %s, args=%v)", m.Source, m.Command, m.Args)
}

func ParseMessage(line string) *Message {
	m := &Message{raw: line}

	if strings.HasPrefix(line, "@") {
		detag := strings.SplitN(line, " ", 2)
		m.Tags = parseTags(detag[0][1:])
		if len(detag) > 1 {
			line = detag[1]
		} else {
			line = ""
		}
	}

	parts := strings.SplitN(line, " :", 2)
	var payload string
	p := false

	if len(parts) > 1 {
		// XXX payload is just a "special" form of the last argument
		p = true
		m.HasTrailer = true
		payload = parts[1]
	}

	parts = strings.Split(parts[0], " ")

	if len(parts[0]) > 0 && parts[0][0] == ':' {
		m.Source = parts[0][1:]
		parts = parts[1:]
	}

	m.Command = parts[0]
	m.Args = parts[1:]
	if p {
		m.Args = append(m.Args, payload)
	}
	return m
}
