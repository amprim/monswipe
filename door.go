package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
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
var doorChannels = []chan struct{}{
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
	if err := p.In(gpio.PullDown, gpio.RisingEdge); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Initialized monitor for door %d (pin %s)\n", door, pin)

	// Wait for edges as detected by the hardware, and print the value read:
	for {
		p.WaitForEdge(-1)
		select {
		case channel <- struct{}{}:
			fmt.Printf("Door %d triggered\n", door)
		default:
			continue
		}
	}
}

func readDoor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	door, ok := vars["door"]
	if !ok {
		http.Error(w, "Missing door number", http.StatusBadRequest)
		return
	}

	dr, err := strconv.Atoi(door)
	if err != nil {
		http.Error(w, "Door number is not a number", http.StatusBadRequest)
		return
	}

	if (dr < 1) || (dr > 4) {
		http.Error(w, "Door number out of range", http.StatusBadRequest)
		return
	}

	select {
	case <-doorChannels[dr-1]:
		http.Error(w, "", http.StatusOK)
		log.Printf("YES read door %d\n", dr)
		return
	default:
		http.Error(w, "", http.StatusNotFound)
		log.Printf("NAN read door %d\n", dr)
		return
	}
}

// clear any residual channel contents
func resetDoors(w http.ResponseWriter, r *http.Request) {
	for {
		select {
		case <-doorChannels[0]:
			continue
		case <-doorChannels[1]:
			continue
		case <-doorChannels[2]:
			continue
		case <-doorChannels[3]:
			continue
		default:
			println("All channels reset")
			return
		}
	}

}

func doorSetup() {
	for i, pin := range doors {
		go listen(i+1, pin, doorChannels[i])
	}
}
