package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

// Input pins for the door monitors,
// in order of door (i.e. index 0 is Door 1, etc.).
var doors = []string{
	"GPIO12",
	"GPIO16",
	"GPIO20",
	"GPIO21",
}

// Channels for reporting state, in door order.
var channels = []chan struct{}{
	make(chan struct{}, 1),
	make(chan struct{}, 1),
	make(chan struct{}, 1),
	make(chan struct{}, 1),
}

func listen(door int, pin string, channel chan struct{}) {
	// Lookup a pin by its number:
	p := gpioreg.ByName(pin)
	if p == nil {
		log.Fatalf("Error on door %d: failed to find %s\n", door, pin)
	}

	// Set it as input, with an internal pull down resistor:
	if err := p.In(gpio.PullDown, gpio.FallingEdge); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Initialized monitor for door %d (pin %s)\n", door, pin)

	// Wait for edges as detected by the hardware, and print the value read:
	for {
		p.WaitForEdge(-1)
		if p.Read() == gpio.High {
			channel <- struct{}{}
			fmt.Printf("Door %d triggered\n", door)
		}
	}
}

func readDoor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	num, ok := vars["num"]
	if !ok {
		http.Error(w, "Missing door number", http.StatusBadRequest)
		return
	}

	dr, err := strconv.Atoi(num)
	if err != nil {
		http.Error(w, "Door number is not a number", http.StatusBadRequest)
		return
	}

	if (dr < 1) || (dr > 4) {
		http.Error(w, "Door number out of range", http.StatusBadRequest)
		return
	}

	select {
	case <-channels[dr-1]:
		http.Error(w, "", http.StatusOK)
		log.Printf("YES read door %d\n", dr)
		return
	default:
		http.Error(w, "", http.StatusNotFound)
		log.Printf("NAN read door %d\n", dr)
		return
	}
}

func handleRequests() {
	r := mux.NewRouter()
	r.HandleFunc("/door/{num}", readDoor)
	http.ListenAndServe("127.0.0.1:8888", r)
}

func main() {
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	for i, pin := range doors {
		go listen(i+1, pin, channels[i])
	}

	handleRequests()
}
