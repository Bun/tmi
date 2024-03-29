package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"raccatta.cc/tmi/irc"
	"raccatta.cc/tmi/ircon"
)

type handler struct {
	Con      *ircon.IRCon
	Channels []string
}

func (h *handler) Connected() {
	fmt.Println("#", time.Now(), "Connected")
	// 50/15
	go func() {
		N := 50
		for i := 0; i < len(h.Channels); i += N {
			end := i + N
			if end > len(h.Channels) {
				end = len(h.Channels)
			}
			clist := strings.Join(h.Channels[i:end], ",")
			fmt.Println("# Join:", clist)
			h.Con.Send("JOIN " + clist)
			time.Sleep(time.Second * 16)
		}
	}()
}

func (handler) Disconnected(err error) {
	fmt.Println("#", time.Now(), "Disconnected:", err)
}

func (handler) Message(msg *ircon.Message) {
	fmt.Println("MSG:", msg.Raw())
}

func main() {
	nick := os.Getenv("NICK")     // Your username
	passwd := os.Getenv("PASSWD") // OAuth token in the form oauth:...
	channel := os.Getenv("CHAN")
	server := os.Getenv("SERVER")

	if server == "irc" {
		server = ircon.DefaultIRCServer
	} else if server == "" {
		server = ircon.DefaultServer
	}

	dialer, err := irc.New(server)
	if err != nil {
		panic(err)
	}

	con := ircon.New(dialer, ircon.TwitchHandshaker(nick, passwd))
	con.Server = server
	h := &handler{
		Con: con,
	}
	if channel != "" {
		chans := strings.Split(channel, ",")
		for i, cname := range chans {
			chans[i] = addPrefix(cname, "#")
		}
		h.Channels = chans
	}

	ctx, cancel := context.WithCancel(context.Background())
	go con.Run(ctx, h)

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
		mux.HandleFunc("/cancel", func(w http.ResponseWriter, r *http.Request) {
			cancel()
		})
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
