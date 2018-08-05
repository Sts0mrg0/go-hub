package main

import (
	"net/http"
	"log"
	"os"
	)

func main() {
	log.Println("start server ...")

	token := os.Getenv("token")

	if token == "" {
		log.Fatalln("no env TOKEN")
		return
	}

	hub := newHub(token)
	go hub.serve()

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("new request: %s", r.RemoteAddr)
		index(hub, w, r)
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
