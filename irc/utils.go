package irc

func safeMessage(m string) []byte {
	buf := append([]byte(m), '\r', '\n')
	for i, c := range buf[:len(buf)-2] {
		if c == '\r' || c == '\n' || c == 0 {
			buf[i] = ' '
		}
	}
	return buf
}
