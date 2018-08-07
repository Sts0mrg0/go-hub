package main

import (
	"log"
	"net"
	"net/http"
)

func index(hub *Hub, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	token := r.Header.Get("token")

	if token != hub.token {
		w.WriteHeader(400)
		w.Write([]byte("bad token"))
		log.Println("bad token")
		return
	}

	nodeIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("userip: %q is not IP:port", r.RemoteAddr)
		w.WriteHeader(500)
		return
	}

	if _, ok := hub.nodes[nodeIP]; ok {
		log.Printf("node: %s already in hub", nodeIP)
	} else {
		hub.register <- nodeIP
	}

	w.WriteHeader(200)
	w.Write([]byte("DONE"))
}
