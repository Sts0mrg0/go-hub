package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Hub struct {
	nodes          map[string]bool
	token          string
	registerNode   chan string
	unregisterNode chan string
	registerUser   chan [2]string
	unregisterUser chan string
	freeNode       chan string
	routes         map[string]string
	removeUserChan map[string]chan bool
}

func (hub *Hub) addNode(node string) {
	hub.nodes[node] = true
	hub.freeNode <- node
}

func (hub *Hub) removeNode(node string) {
	delete(hub.nodes, node)
}

func (hub *Hub) getFreeNode(user string) string {
	for {
		node := <-hub.freeNode
		_, nodeExist := hub.nodes[node]
		if nodeExist {
			return node
		}
	}
}

func (hub *Hub) addRoute(userObject [2]string) {
	hub.routes[userObject[0]] = userObject[1]
}

func (hub *Hub) removeRoute(userIP string) {
	if node, ok := hub.routes[userIP]; ok {
		delete(hub.routes, userIP)
		hub.freeNode <- node
	}
}

func (hub *Hub) serve() {
	for {
		select {
		case addNode := <-hub.registerNode:
			hub.addNode(addNode)
			go hub.pingNode(addNode)
		case removeNode := <-hub.unregisterNode:
			hub.removeNode(removeNode)
		case addUser := <-hub.registerUser:
			hub.addRoute(addUser)
		case removeUser := <-hub.unregisterUser:
			hub.removeRoute(removeUser)
		}
	}
}

func (hub *Hub) checkLostConnect(c chan bool, userIP string) {
	for {
		timer := time.NewTimer(time.Minute)

		select {
		case <-c:
			timer.Stop()
			continue
		case <-timer.C:
			hub.unregisterUser <- userIP
			timer.Stop()
			log.Printf("user %s lost connection", userIP)
			return
		}
	}
}

func (hub *Hub) pingNode(nodeString string) {
	countErrors := 0

	var netClient = &http.Client{
		Timeout: time.Second * 2,
	}

	ticker := time.NewTicker(time.Second * 4)
	defer ticker.Stop()

	for range ticker.C {
		if _, ok := hub.nodes[nodeString]; !ok {
			log.Printf("node: %s removed from hub", nodeString)
			return
		}

		if countErrors >= 5 {
			log.Printf("node: %s lost connection, limit timeout", nodeString)
			hub.unregisterNode <- nodeString
			return
		}

		req, err := http.NewRequest("GET", "http://"+nodeString+":6677/register", nil)

		if err != nil {
			log.Println(err)
			countErrors++
			continue
		}

		req.Header.Set("token", hub.token)
		resp, err := netClient.Do(req)

		if err != nil {
			log.Println(err)
			countErrors++
			continue
		}

		bodyByte, e := ioutil.ReadAll(resp.Body)

		if e != nil {
			log.Println(e)
			resp.Body.Close()
			return
		}

		body := string(bodyByte)

		if body == "bad token" {
			log.Printf("need check auth token on node: %s", nodeString)
			resp.Body.Close()
			return
		}

		if body == "PONG" {
			log.Printf("node: %s is ACTIVE", nodeString)
			countErrors = 0
		} else {
			countErrors++
			log.Printf("wrong answer from node: %s, body: %s", nodeString, body)
		}
		resp.Body.Close()
	}
}

func newHub(token string) *Hub {
	return &Hub{
		nodes:          make(map[string]bool),
		token:          token,
		registerNode:   make(chan string, 10),
		unregisterNode: make(chan string, 10),
		registerUser:   make(chan [2]string),
		unregisterUser: make(chan string),
		routes:         make(map[string]string),
		freeNode:       make(chan string, 1000),
		removeUserChan: make(map[string]chan bool),
	}
}
