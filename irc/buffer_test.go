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
}
