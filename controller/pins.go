package main

import (
	"fmt"
	"log"

	rpio "github.com/stianeikeland/go-rpio"
)

type Pins []rpio.Pin

func NewPins(ns ...int) Pins {
	pins := make(Pins, 0)

	for _, n := range ns {
		pin := rpio.Pin(n)
		pin.Output()
		pins = append(pins, pin)
	}

	return pins
}

func (pins Pins) Set(states []bool) error {
	if numPins := len(pins); len(states) < numPins {
		return fmt.Errorf("Too few states provided, need atleast %v", numPins)
	}
	for i := 0; i < len(pins); i++ {
		state := states[i]

		log.Printf("Setting pin %v (%v) to %v", i, pins[i], state)
		if state {
			pins[i].High()
		} else {
			pins[i].Low()
		}
	}
	return nil
}

func (pins Pins) Low() {
	for _, pin := range pins {
		pin.Low()
	}
}
