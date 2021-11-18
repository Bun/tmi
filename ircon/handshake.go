package ircon

// IRCHandshake implements a standard pre-registered-state handshake for the
// IRC protocol.
type IRCHandshake struct {
	// Caps contains the set of capabilities that are requested on connect.
	// Advanced users can specify their own set.
	Caps string

	Nick   string
	GECOS  string
	Passwd string
}

// TwitchHandshaker creates an IRCHandshake with some Twitch-specific details.
func TwitchHandshaker(nick, passwd string) IRCHandshake {
	if nick == "" {
		// Default anonymous login; cannot send messages(!)
		nick = "justinfan12345"
		passwd = "blah"
	}
	return IRCHandshake{
		Caps:   DefaultCaps,
		Nick:   nick,
		Passwd: passwd,
		GECOS:  nick,
	}
}

func (irc IRCHandshake) Handshake(con Sender) error {
	if irc.Caps != "" {
		if err := con.Send("CAP REQ :" + irc.Caps); err != nil {
			return err
		}
	}
	if irc.Passwd != "" {
		if err := con.Send("PASS " + irc.Passwd); err != nil {
			return err
		}
	}
	if err := con.Send("NICK " + irc.Nick); err != nil {
		return err
	}
	if err := con.Send("USER " + irc.Nick + " 8 * :" + irc.GECOS); err != nil {
		return err
	}
	return nil
}
