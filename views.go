package main

import (
	"fmt"
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

	token := r.Header.Get(tokenKey)

	if token != hub.Token {
		w.WriteHeader(400)
		w.Write([]byte(badToken))
		log.Println(badToken)
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
	token := r.Header.Get(tokenKey)

	if token != hub.Token {
		w.WriteHeader(400)
		w.Write([]byte(badToken))
		log.Println(badToken)
		return
	}

	number := r.Header.Get(numberKey)

	userIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("userip: %q is not IP:port", r.RemoteAddr)
		w.WriteHeader(500)
		return
	}

	user := number + "-" + userIP

	if node, ok := hub.Routes.Load(user); ok {
		u, _ := url.Parse(fmt.Sprintf(urlNode, node.(string)))
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

		u, _ := url.Parse(fmt.Sprintf(urlNode, freeNode))
		proxy := httputil.NewSingleHostReverseProxy(u)
		proxy.ServeHTTP(w, r)
	}
}
