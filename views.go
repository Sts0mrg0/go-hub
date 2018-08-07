package main

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
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
		log.Printf("nodeIp: %q is not IP:port", r.RemoteAddr)
		w.WriteHeader(500)
		return
	}

	if _, ok := hub.nodes[nodeIP]; ok {
		log.Printf("node: %s already in hub", nodeIP)
	} else {
		hub.registerNode <- nodeIP
	}

	w.WriteHeader(200)
	w.Write([]byte("DONE"))
}

func wdHub(hub *Hub, w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")

	if token != hub.token {
		w.WriteHeader(400)
		w.Write([]byte("bad token"))
		log.Println("bad token")
		return
	}

	userIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("userip: %q is not IP:port", r.RemoteAddr)
		w.WriteHeader(500)
		return
	}

	if node, ok := hub.routes[userIP]; ok {
		u, _ := url.Parse("http://" + node + ":6677")
		if c, _ok := hub.removeUserChan[userIP]; _ok {
			c <- true
		}
		proxy := httputil.NewSingleHostReverseProxy(u)
		proxy.ServeHTTP(w, r)
	} else {
		freeNode := hub.getFreeNode(userIP)
		hub.registerUser <- [2]string{userIP, freeNode}
		hub.removeUserChan[userIP] = make(chan bool, 1)

		go hub.checkLostConnect(hub.removeUserChan[userIP], userIP)

		u, _ := url.Parse("http://" + freeNode + ":6677")
		proxy := httputil.NewSingleHostReverseProxy(u)
		proxy.ServeHTTP(w, r)
	}
}
