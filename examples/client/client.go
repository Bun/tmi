package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"bnsvc.net/tmi/ircon"
)

type handler struct {
	Con      *ircon.IRCon
	Channels []string
}

func (h *handler) Connected() {
	fmt.Println("#", time.Now(), "Connected")
	for _, c := range h.Channels {
		h.Con.Send("JOIN " + c)
	}
}

func (handler) Disconnected(err error) {
	fmt.Println("#", time.Now(), "Disconnected:", err)
}

func (handler) Message(msg *ircon.Message) {
	fmt.Println(msg.Raw())
}

func main() {
	nick := os.Getenv("NICK")     // Your username
	passwd := os.Getenv("PASSWD") // OAuth token in the form oauth:...
	channel := os.Getenv("CHAN")

	ctx := context.Background()
	con := ircon.New(nick, passwd)
	h := &handler{
		Con: con,
	}
	if channel != "" {
		h.Channels = []string{addPrefix(channel, "#")}
	}
	con.Handler = h
	con.Background(ctx)

	raw := func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Raw string `json:"raw"`
		}
		d := json.NewDecoder(r.Body)
		if err := d.Decode(&req); err != nil || req.Raw == "" {
			w.WriteHeader(400)
			return
		}
		con.Send(req.Raw)
	}

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/raw", raw)
		s := &http.Server{
			Addr:           "localhost:2048",
			Handler:        mux,
			ReadTimeout:    120 * time.Second,
			WriteTimeout:   120 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		s.ListenAndServe()
	}()

	<-ctx.Done()
}

func addPrefix(s, pfx string) string {
	if !strings.HasPrefix(s, pfx) {
		return pfx + s
	}
	return s
}
