package main

import (
	"hub/utils"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func index(hub *utils.Hub, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	token := r.Header.Get("token")

	if token != hub.Token {
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

	if _, ok := hub.Nodes.Load(nodeIP); ok {
		log.Printf("node: %s already in hub", nodeIP)
	} else {
		hub.RegisterNode <- nodeIP
	}

	w.WriteHeader(200)
	w.Write([]byte("DONE"))
}

func wdHub(hub *utils.Hub, w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")

	if token != hub.Token {
		w.WriteHeader(400)
		w.Write([]byte("bad token"))
		log.Println("bad token")
		return
	}

	number := r.Header.Get("number")

	userIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("userip: %q is not IP:port", r.RemoteAddr)
		w.WriteHeader(500)
		return
	}

	user := number + "-" + userIP

	if node, ok := hub.Routes.Load(user); ok {
		u, _ := url.Parse("http://" + node.(string) + ":6677")
		if c, _ok := hub.RemoveUserChan.Load(user); _ok {
			c.(chan bool) <- true
		}
		proxy := httputil.NewSingleHostReverseProxy(u)
		proxy.ServeHTTP(w, r)
	} else {
		freeNode := hub.GetFreeNode(user)
		hub.RegisterUser <- [2]string{user, freeNode}
		hub.RemoveUserChan.Store(user, make(chan bool, 1))

		c, _ := hub.RemoveUserChan.Load(user)

		go hub.CheckLostConnect(c.(chan bool), user)

		u, _ := url.Parse("http://" + freeNode + ":6677")
		proxy := httputil.NewSingleHostReverseProxy(u)
		proxy.ServeHTTP(w, r)
	}
}
