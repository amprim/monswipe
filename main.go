package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"periph.io/x/host/v3"
)

func handleRequests() {
	r := mux.NewRouter()
	r.HandleFunc("/door/{door}", readDoor)
	r.HandleFunc("/door-reset", resetDoors)
	r.HandleFunc("/swipe/{door}/{fc}/{cn}", swipe)
	http.ListenAndServe("0.0.0.0:8888", r)
}

func main() {
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	doorSetup()
	swipeSetup()

	handleRequests()
}
