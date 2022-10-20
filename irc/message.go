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
	t := make(map[string]string)
	iter := strIter{s: s, b: ';'}
	for iter.Next() {
		key, value := splitb(iter.v, '=')
		if strings.IndexByte(value, '\\') > -1 {
			t[key] = unescapeTag.Replace(value)
		} else {
			t[key] = value
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
		var tags string
		tags, line = splitb(line, ' ')
		m.Tags = parseTags(tags[1:])
	}

	if strings.HasPrefix(line, ":") {
		var source string
		source, line = splitb(line, ' ')
		m.Source = source[1:]
	}

	// Payload is just a "special" form of the last argument
	var payload string
	line, payload, m.HasTrailer = split(line, " :")

	args := strings.Split(line, " ")
	m.Command = args[0]
	m.Args = args[1:]
	if m.HasTrailer {
		m.Args = append(m.Args, payload)
	}
	return m
}

func splitb(s string, b byte) (string, string) {
	c := strings.IndexByte(s, b)
	if c == -1 {
		return s, ""
	}
	return s[:c], s[c+1:]
}

func split(s, f string) (string, string, bool) {
	c := strings.Index(s, f)
	if c == -1 {
		return s, "", false
	}
	return s[:c], s[c+len(f):], true
}

type strIter struct {
	s string
	b byte
	v string
}

func (i *strIter) Next() bool {
	if i.s == "" {
		return false
	}
	i.v, i.s = splitb(i.s, i.b)
	return true
}
