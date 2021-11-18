package irc

import "testing"

func TestBuffer(t *testing.T) {
	b := buffer{}
	b.Next([]byte("NOTICE * :Hi\n"))

	msg, err := b.Read()
	if err != nil {
		t.Fatal(err)
	} else if msg == nil {
		t.Error("Expected message")
	}

	msg, err = b.Read()
	if err != nil {
		t.Fatal(err)
	} else if msg != nil {
		t.Error("Did not expect message")
	}

	b.Next([]byte("PRIVMSG #a :1\r\n"))
	b.Next([]byte("PRIVMSG #b :2\r\n"))
	b.Next([]byte("PRIVMSG #b"))

	i := 0
	for {
		msg, err = b.Read()
		if err != nil {
			t.Fatal(err)
		} else if msg == nil {
			break
		}
		i++
	}
	if i != 2 {
		t.Error(i)
	}

	b.Next([]byte(" :3\r\n"))
	msg, err = b.Read()
	if err != nil {
		t.Fatal(err)
	} else if msg == nil {
		t.Error("Expected message")
	}
}
