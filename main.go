package main

import (
	"hub/utils"
	"log"
	"net/http"
	"os"
)

var (
	badToken   = "bad token"
	tokenKey   = "token"
	numberKey  = "number"
	noEnv      = "no env %s"
	newRequest = "new request %s: %s"
	urlNode    = "http://%s:6677"
)

func main() {
	log.Println("start server ...")

	token := os.Getenv(tokenKey)

	if token == "" {
		log.Fatalf(noEnv, token)
		return
	}

	hub := utils.NewHub(token)
	go hub.Serve()

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		log.Printf(newRequest, r.RequestURI, r.RemoteAddr)
		index(hub, w, r)
	})

	http.HandleFunc("/wd/hub/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf(newRequest, r.RequestURI, r.RemoteAddr)
		wdHub(hub, w, r)
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
