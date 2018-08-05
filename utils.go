package main

import (
	"time"
	"net/http"
	"log"
	"io/ioutil"
	)

type Hub struct {
	nodes map[string]bool
	token string
	register chan string
	unregister chan string
}

func (hub *Hub) addNode(node string) {
	hub.nodes[node] = true
}

func (hub *Hub) removeNode(node string) {
	delete(hub.nodes, node)
}

func (hub *Hub) serve() {
	for {
		select {
		case add := <- hub.register:
			hub.addNode(add)
			go hub.pingNode(add)
		case remove := <- hub.unregister:
			hub.removeNode(remove)
		}
	}
}


func (hub *Hub) pingNode(nodeString string) {
	countErrors := 0

	ticker := time.NewTicker(time.Second * 4)
	defer ticker.Stop()

	for range ticker.C {
		if _, ok := hub.nodes[nodeString]; !ok {
			log.Fatalf("node: %s removed from hub", nodeString)
			return
		}

		if countErrors >= 5 {
			log.Fatalf("node: %s lost connection, limit timeout", nodeString)
			hub.unregister <- nodeString
			return
		}
		var netClient = &http.Client{
			Timeout: time.Second * 2,
		}
		resp, e := netClient.Get("http://" + nodeString + ":6677?token=" + hub.token)

		if e != nil {
			log.Println(e)
			countErrors++
			continue
		}

		defer resp.Body.Close()

		bodyByte, e := ioutil.ReadAll(resp.Body)

		if e != nil {
			log.Println(e)
			return
		}

		body := string(bodyByte)

		if body == "bad token" {
			log.Fatalf("need check auth token on node: %s", nodeString)
			return
		}

		if body == "PONG" {
			log.Printf("node: %s is ACTIVE", nodeString)
			countErrors = 0
		} else {
			countErrors++
			log.Printf("wrong answer from node: %s, body: %s", nodeString, body)
		}
	}
}


func newHub(token string) *Hub {
	return &Hub {
		nodes: make(map[string]bool),
		token: token,
		register: make(chan string, 10),
		unregister: make(chan string, 10),
	}
}