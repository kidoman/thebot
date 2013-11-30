package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/gorilla/mux"
)

var (
	camWidth       = flag.Int("camw", 320, "width of the captured camera image")
	camHeight      = flag.Int("camh", 240, "height of the captured camera image")
	camFps         = flag.Int("fps", 2, "fps for camera")
	arduinoAddrStr = flag.String("addr", "0x50", "arduino i2c address")
)

func main() {
	log.Print("Starting up...")

	flag.Parse()

	camera := NewCamera(*camWidth, *camHeight, *camFps)
	defer camera.Close()
	camera.Run()

	arduinoAddr, err := strconv.ParseInt(*arduinoAddrStr, 0, 0)
	if err != nil {
		log.Fatalf("Could not parse %q for arduino i2c address", *arduinoAddrStr)
	}
	car := NewCar(byte(arduinoAddr))

	r := mux.NewRouter()

	r.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		http.ServeFile(resp, req, "/home/pi/index.html")
	}).
		Methods("GET")

	r.HandleFunc("/turn/{angle}/speed/{speed}", func(resp http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		angle, err := strconv.Atoi(vars["angle"])
		if err != nil {
			http.Error(resp, "angle not valid", http.StatusBadRequest)
			return
		}
		speed, err := strconv.Atoi(vars["speed"])
		if err != nil {
			http.Error(resp, "speed not valid", http.StatusBadRequest)
			return
		}
		log.Printf("Received orientation %v, %v", angle, speed)
		if err = car.Turn(angle); err != nil {
			http.Error(resp, "could not send message to arduino", http.StatusInternalServerError)
			return
		}
		if err = car.Speed(speed); err != nil {
			http.Error(resp, "could not send message to arduino", http.StatusInternalServerError)
		}
	}).
		Methods("POST")

	r.HandleFunc("/orientation", func(resp http.ResponseWriter, req *http.Request) {
		speed, angle := car.Orientation()
		fmt.Fprintf(resp, "%v, %v", speed, angle)
	})

	r.HandleFunc("/reset", func(resp http.ResponseWriter, req *http.Request) {
		log.Print("Resetting...")
		if err := car.Reset(); err != nil {
			http.Error(resp, "could not reset", http.StatusInternalServerError)
		}
	}).
		Methods("POST")

	r.HandleFunc("/snapshot", func(resp http.ResponseWriter, req *http.Request) {
		log.Print("Sending current snapshot")

		image := camera.CurrentImage()
		resp.Write(image)
	})

	log.Print("Starting web server")
	go http.ListenAndServe(":8080", r)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit

	log.Print("All done")
}
