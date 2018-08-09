package main

import (
	"hub/utils"
	"log"
	"net/http"
	"os"
)

func main() {
	log.Println("start server ...")

	token := os.Getenv("token")

	if token == "" {
		log.Fatalln("no env TOKEN")
		return
	}

	hub := utils.NewHub(token)
	go hub.Serve()

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("new request %s: %s", r.RequestURI, r.RemoteAddr)
		index(hub, w, r)
	})

	http.HandleFunc("/wd/hub/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("new request %s: %s", r.RequestURI, r.RemoteAddr)
		wdHub(hub, w, r)
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
