# monswipe

Simple input pin monitor and Wiegand outputter written in Go. Written to monitor door pins from, and to send Wiegand to, an Uhppote access control board as part of hardware testing.

## Setup

### Door monitor
As written, `monswipe` expects to run on a Raspberry Pi (tested on a 3B), with door inputs (in door order) on `GPIO12`, `GPIO16`, `GPIO20`, and `GPIO21`.

Door inputs from the Uhppote board should be routed through a 12V -> 3.3V optocoupler or other conversion circuit, such that triggering a door (connected by NO) will trigger a one-pin input on the Pi (to the appropriate GPIO).

### Swipe simulator
As written, `monswipe` expects to run on a Raspberry Pi (tested on a 3B), with wiegand outputs (in door and D0,D1 order) on `{"GPIO17", "GPIO27"}`, `{"GPIO22", "GPIO5"}`, `{"GPIO6", "GPIO13"}`, and `{"GPIO19", "GPIO26"}`.

All wiegand lines should be routed through a 3.3v <-> 5V logic converter, to prevent frying the Pi. Power for the 5V side of the equation can be provided by the Pi's 5V lines - GND must be provided from the Uhppote board (I used the reader GND lines).

## Usage

`monswipe` exposes three HTTP GET endpoints on `127.0.0.1:8888`.

* `/door/{door}` - Check if `{door}` has been triggered
    * `{door}` must be a number in the range 1-4.
    * If door has been triggered: `200 OK`
    * If door has not been triggered: `404 File Not Found`
    * You are expected to check the door trigger as soon as you send a (valid) swipe and/or unlock command. Additional triggers while a trigger is pending read will be ignored.

* `/door-reset` - Resets all door trigger channels.

* `/swipe/{fc}/{cn}` - Simulate a Wiegand swipe
    * currently supports only 26-bit/H10301 card data
        * `fc` must be in the range 0-255
        * `cn` must be in the range 0-65535