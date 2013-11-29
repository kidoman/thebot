package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/stianeikeland/go-rpio"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
)

const (
	arduinoAddr  = 0x50
	cameraWidth  = 640
	cameraHeight = 480
	cameraFps    = 4
)

func main() {
	log.Print("Starting up...")

	err := rpio.Open()
	if err != nil {
		panic(err)
	}
	defer func() {
		log.Print("Cleaning up...")
		rpio.Close()
	}()

	car := NewCar(arduinoAddr)
	camera := NewCamera(cameraWidth, cameraHeight, cameraFps)
	defer camera.Close()

	camera.Run()

	r := mux.NewRouter()

	r.HandleFunc("/turn/{angle}", func(resp http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		angle, err := strconv.Atoi(vars["angle"])
		if err != nil {
			http.Error(resp, "angle not valid", http.StatusBadRequest)
			return
		}
		log.Printf("Received angle %v", angle)
		if err = car.Turn(angle); err != nil {
			http.Error(resp, "could not send message to arduino", http.StatusInternalServerError)
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
		if err = car.Speed(speed); err != nil {
			http.Error(resp, "could not send message to arduino", http.StatusInternalServerError)
		}
	}).
		Methods("POST")

	r.HandleFunc("/orientation", func(resp http.ResponseWriter, req *http.Request) {
		speed, angle := car.Orientation()
		fmt.Fprintf(resp, "%v, %v", speed, angle)
	}).
		Methods("GET")

	r.HandleFunc("/reset", func(resp http.ResponseWriter, req *http.Request) {
		log.Print("Resetting...")
		if err = car.Reset(); err != nil {
			http.Error(resp, "could not reset", http.StatusInternalServerError)
		}
	}).
		Methods("POST")

	r.HandleFunc("/snapshot", func(resp http.ResponseWriter, req *http.Request) {
		log.Print("Sending current snapshot")

		image := camera.CurrentImage()
		resp.Write(image)
	}).
		Methods("GET")

	log.Print("Starting web server")
	go http.ListenAndServe(":8080", r)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit
}
