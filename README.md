# doormon

Simple input pin monitor written in Go. Written to monitor pin inputs from a Uhppote access control board as part of hardware testing.

## Setup

As written, `doormon` expects to run on a Raspberry Pi (tested on a 3B), with input (in door order) on `GPIO12`, `GPIO16`, `GPIO20`, and `GPIO21`.

Door inputs from the Uhppote board should be routed through a 12V -> 3.3V optocoupler or other conversion circuit, such that triggering a door (connected by NO) will trigger a one-pin input on the Pi (to the appropriate GPIO).