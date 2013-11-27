package main

import (
	"flag"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/stianeikeland/go-rpio"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
)

var (
	anglePinsStr = flag.String("angle", "17,18,10,22,23", "pins to use for angle control")
	speedPinsStr = flag.String("speed", "9,24,25", "pins to use for speed control")
)

func main() {
	flag.Parse()

	log.Print("Starting up...")

	err := rpio.Open()
	if err != nil {
		panic(err)
	}
	defer func() {
		log.Print("Cleaning up...")
		rpio.Close()
	}()

	anglePinsInts := extractPins(*anglePinsStr)
	speedPinsInts := extractPins(*speedPinsStr)

	log.Printf("Angle pins %v...", anglePinsInts)
	log.Printf("Speed pins %v...", speedPinsInts)

	anglePins := NewPins(anglePinsInts...)
	totalAnglePins := len(anglePins)
	speedPins := NewPins(speedPinsInts...)
	totalSpeedPins := len(anglePins)

	defer func() {
		log.Print("Setting pins low")
		anglePins.Low()
		speedPins.Low()
	}()

	r := mux.NewRouter()

	r.HandleFunc("/turn/{angle}", func(resp http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		angle, err := strconv.Atoi(vars["angle"])
		if err != nil {
			http.Error(resp, "angle not valid", http.StatusBadRequest)
			return
		}
		log.Printf("Received angle %v", angle)
		if angle == 180 {
			log.Print("Changing 180 -> 179")
			angle = 179
		}
		bits := angleToBits(angle, totalAnglePins)
		err = anglePins.Set(bits)
		if err != nil {
			panic(err)
		}
	}).
		Methods("POST")

	r.HandleFunc("/speed/{speed}", func(resp http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		speed, err := strconv.Atoi(vars["speed"])
		if err != nil {
			http.Error(resp, "speed not valid", http.StatusBadRequest)
			return
		}
		log.Printf("Received speed %v", speed)
		bits := nToBits(speed, totalSpeedPins)
		err = speedPins.Set(bits)
		if err != nil {
			panic(err)
		}
	}).
		Methods("POST")

	log.Print("Starting web server")
	go http.ListenAndServe(":8080", r)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit
}

func extractPins(pins string) []int {
	split := strings.Split(pins, ",")
	ps := make([]int, 0)
	for _, ns := range split {
		n, err := strconv.Atoi(ns)
		if err != nil {
			log.Printf("Could not parse %q", ns)
			log.Fatal(err)
		}
		ps = append(ps, n)
	}
	return ps
}

func angleToBits(angle, totalPins int) []bool {
	a := math.Pow(2, float64(totalPins))
	b := 180 / a

	scaledAngle := int(float64(angle) / b)

	return nToBits(scaledAngle, totalPins)
}

func nToBits(n, totalPins int) []bool {
	bits := make([]bool, totalPins)

	for i := range bits {
		bits[i] = checkBit(n, uint(i))
	}

	return bits
}

func checkBit(n int, bit uint) bool {
	if (n & (1 << bit)) != 0 {
		return true
	}
	return false
}
