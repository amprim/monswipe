package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
)

// Output pins for each door reader,
// in door order as D0, D1
var readers = [][]string{
	{"GPIO4", "GPIO27"},
	{"GPIO22", "GPIO5"},
	{"GPIO6", "GPIO13"},
	{"GPIO19", "GPIO26"},
}

var passthrough = []string{"GPIO23", "GPIO24"}

// Channels for sending cards, in door order.
var swipeChannels = []chan []int{
	make(chan []int, 1),
	make(chan []int, 1),
	make(chan []int, 1),
	make(chan []int, 1),
}

//var currentPassthrough = 1

func writePin(pin gpio.PinIO) {
	if err := pin.Out(gpio.Low); err != nil {
		log.Fatal(err)
	}
	time.Sleep(50 * time.Microsecond)
	if err := pin.Out(gpio.High); err != nil {
		log.Fatal(err)
	}
}

func parity(message string) string {
	left, right := message[:12], message[12:]
	if len(left) != 12 || len(right) != 12 {
		log.Fatalf("sides are not correct length (%d, %d)!", len(left), len(right))
	}

	l := "0"
	if (strings.Count(left, "1") % 2) == 1 {
		l = "1"
	}

	r := "1"
	if (strings.Count(right, "1") % 2) == 1 {
		r = "0"
	}

	return l + message + r
}

/*func readPins(d0 gpio.PinIO, d1 gpio.PinIO) int {
	for {
		if d0.WaitForEdge(-1) || d1.WaitForEdge(-1) {
			if d0.Read() == gpio.Low {
				println("read 0")
				return 0
			} else {
				println("read 1")
				return 1
			}
		}
	}
}*/

// pass through inputs from physical reader
/*func pass(pins []string) {
	// Lookup a pin by its number:
	d0 := gpioreg.ByName(pins[0])
	d1 := gpioreg.ByName(pins[1])
	if d0 == nil || d1 == nil {
		log.Fatalf("Error on passthrough: failed to find D0 and/or D1\n")
	}

	// Set d0 and d1 as input:
	if err := d0.In(gpio.PullUp, gpio.FallingEdge); err != nil {
		log.Fatal(err)
	}

	if err := d1.In(gpio.PullUp, gpio.FallingEdge); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Initialized passthrough reader (pins %s/%s)\n", pins[0], pins[1])

	// Wait for card data, then send:
	card := make([]int, 26)
	for {
		for {
			if len(card) == 26 {
				break
			}

			if d0.WaitForEdge(100*time.Microsecond) || d1.WaitForEdge(100*time.Microsecond) {
				if d0.Read() == gpio.Low {
					println("read 0")
					card = append(card, 0)
				} else {
					println("read 1")
					card = append(card, 1)
				}

				time.Sleep(2 * time.Millisecond)
			}
		}

		log.Printf("passthrough data: %d ", card)
		card = make([]int, 26)
	}
}*/

// send fake reader inputs to the controller
func send(reader int, pins []string, channel chan []int) {
	// Lookup a pin by its number:
	d0 := gpioreg.ByName(pins[0])
	d1 := gpioreg.ByName(pins[1])
	if d0 == nil || d1 == nil {
		log.Fatalf("Error on reader %d: failed to find D0 and/or D1\n", reader)
	}

	// Set d0 and d1 as output:
	if err := d0.Out(gpio.High); err != nil {
		log.Fatal(err)
	}

	if err := d1.Out(gpio.High); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Initialized sim swiper for reader %d (pins %s/%s)\n", reader, pins[0], pins[1])

	// Wait for card data, then send:
	for {
		card := <-channel
		fac := card[0]
		cdn := card[1]

		fcb := strconv.FormatInt(int64(fac), 2)
		fc := fmt.Sprintf("%08s", fcb)

		cnb := strconv.FormatInt(int64(cdn), 2)
		cn := fmt.Sprintf("%016s", cnb)

		//println(fc)
		//println(cn)

		message := fc + cn
		if len(message) != 24 {
			log.Fatalf("padded wiegand message is not correct length (%d)!", len(message))
		}
		message = parity(message)
		//println(message)

		for _, bit := range message {
			if bit == '0' {
				// write D0
				//print("0")
				writePin(d0)
			} else {
				// write D1
				//print("1")
				writePin(d1)
			}
			time.Sleep(2 * time.Millisecond)
		}
		fmt.Printf("\nReader %d: sent card data %d\n", reader, card)
	}
}

func swipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	door, ok := vars["door"]
	if !ok {
		http.Error(w, "Missing door number", http.StatusBadRequest)
		return
	}

	dr, err := strconv.Atoi(door)
	if err != nil {
		http.Error(w, "Door number is not an integer", http.StatusBadRequest)
		return
	}

	if (dr < 1) || (dr > 4) {
		http.Error(w, "Door number out of range", http.StatusBadRequest)
		return
	}

	fac, ok := vars["fc"]
	if !ok {
		http.Error(w, "Missing facility code", http.StatusBadRequest)
		return
	}

	fc, err := strconv.Atoi(fac)
	if err != nil {
		http.Error(w, "facility code is not an integer", http.StatusBadRequest)
		return
	}

	if (fc < 0) || (fc > 255) {
		http.Error(w, "facility code out of range", http.StatusBadRequest)
		return
	}

	card, ok := vars["cn"]
	if !ok {
		http.Error(w, "Missing card number", http.StatusBadRequest)
		return
	}

	cn, err := strconv.Atoi(card)
	if err != nil {
		http.Error(w, "card number is not an integer", http.StatusBadRequest)
		return
	}

	if (cn < 0) || (cn > 65535) {
		http.Error(w, "Card number out of range", http.StatusBadRequest)
		return
	}

	swipeChannels[dr-1] <- []int{fc, cn}

	http.Error(w, "", http.StatusOK)
	//log.Printf("Door %d: got card data %d.%d", dr, fc, cn)
}

func swipeSetup() {
	for i, pins := range readers {
		go send(i+1, pins, swipeChannels[i])
	}

	//go pass(passthrough)
}
