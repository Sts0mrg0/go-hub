package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type Hub struct {
	Nodes          sync.Map
	Token          string
	RegisterNode   chan string
	UnregisterNode chan string
	RegisterUser   chan [2]string
	UnregisterUser chan string
	FreeNode       chan string
	Routes         sync.Map
	RemoveUserChan sync.Map
}

func (hub *Hub) AddNode(node string) {
	hub.Nodes.Store(node, true)
	hub.FreeNode <- node
}

func (hub *Hub) RemoveNode(node string) {
	hub.Nodes.Delete(node)
}

func (hub *Hub) GetFreeNode(user string) string {
	for {
		node := <-hub.FreeNode

		_, nodeExist := hub.Nodes.Load(node)
		if nodeExist {
			return node
		}
	}
}

func (hub *Hub) addRoute(userObject [2]string) {
	hub.Routes.Store(userObject[0], userObject[1])
}

func (hub *Hub) removeRoute(userIP string) {
	if node, ok := hub.Routes.Load(userIP); ok {
		hub.Routes.Delete(userIP)
		hub.FreeNode <- node.(string)
	}
}

func (hub *Hub) Serve() {
	for {
		select {
		case addNode := <-hub.RegisterNode:
			hub.AddNode(addNode)
			go hub.pingNode(addNode)
		case removeNode := <-hub.UnregisterNode:
			hub.RemoveNode(removeNode)
		case addUser := <-hub.RegisterUser:
			hub.addRoute(addUser)
		case removeUser := <-hub.UnregisterUser:
			hub.removeRoute(removeUser)
		}
	}
}

func (hub *Hub) CheckLostConnect(c chan bool, userIP string) {
	for {
		timer := time.NewTimer(time.Minute)

		select {
		case <-c:
			timer.Stop()
			continue
		case <-timer.C:
			hub.UnregisterUser <- userIP
			timer.Stop()
			log.Printf("user %s lost connection", userIP)
			return
		}
	}
}

func (hub *Hub) pingNode(nodeString string) {
	countErrors := 0

	netClient := &http.Client{
		Timeout: time.Second * 2,
	}

	ticker := time.NewTicker(time.Second * 4)
	defer ticker.Stop()

	for range ticker.C {
		err := everyTick(netClient, countErrors, hub, nodeString)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func NewHub(token string) *Hub {
	return &Hub{
		Token:          token,
		RegisterNode:   make(chan string, 10),
		UnregisterNode: make(chan string, 10),
		RegisterUser:   make(chan [2]string),
		UnregisterUser: make(chan string),
		FreeNode:       make(chan string, 1000),
	}
}

func everyTick(netClient *http.Client, countErrors int, hub *Hub, nodeString string) error {
	if _, ok := hub.Nodes.Load(nodeString); !ok {
		return &ErrorString{fmt.Sprintf("node: %s removed from hub", nodeString)}
	}

	if countErrors >= 5 {
		hub.UnregisterNode <- nodeString
		return &ErrorString{fmt.Sprintf("node: %s lost connection, limit timeout", nodeString)}
	}

	req, err := http.NewRequest("GET", "http://"+nodeString+":6677/register", nil)

	if err != nil {
		log.Println(err)
		countErrors++
		return nil
	}

	req.Header.Set("token", hub.Token)
	resp, err := netClient.Do(req)

	if err != nil {
		log.Println(err)
		countErrors++
		return nil
	}

	bodyByte, e := ioutil.ReadAll(resp.Body)

	if e != nil {
		resp.Body.Close()
		return e
	}

	body := string(bodyByte)

	if body == "bad token" {
		resp.Body.Close()
		return &ErrorString{fmt.Sprintf("need check auth token on node: %s", nodeString)}
	}

	if body == "PONG" {
		log.Printf("node: %s is ACTIVE", nodeString)
		countErrors = 0
	} else {
		countErrors++
		log.Printf("wrong answer from node: %s, body: %s", nodeString, body)
	}
	resp.Body.Close()

	return nil
}
