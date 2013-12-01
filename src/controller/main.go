package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/codegangsta/martini"
)

var (
	camWidth       = flag.Int("camw", 640, "width of the captured camera image")
	camHeight      = flag.Int("camh", 480, "height of the captured camera image")
	camFps         = flag.Int("fps", 4, "fps for camera")
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

	m := martini.Classic()

	m.Post("/speed/:speed/angle/:angle", func(w http.ResponseWriter, params martini.Params) {
		speed, err := strconv.Atoi(params["speed"])
		if err != nil {
			http.Error(w, "speed not valid", http.StatusBadRequest)
			return
		}
		angle, err := strconv.Atoi(params["angle"])
		if err != nil {
			http.Error(w, "angle not valid", http.StatusBadRequest)
			return
		}
		log.Printf("Received orientation %v, %v", angle, speed)
		if err = car.Turn(angle); err != nil {
			log.Print(err)
			http.Error(w, "could not send message to arduino", http.StatusInternalServerError)
			return
		}
		if err = car.Speed(speed); err != nil {
			log.Print(err)
			http.Error(w, "could not send message to arduino", http.StatusInternalServerError)
		}
	})

	m.Get("/orientation", func() string {
		speed, angle := car.Orientation()
		return fmt.Sprintf("%v, %v", speed, angle)
	})

	m.Post("/reset", func(w http.ResponseWriter, r *http.Request) {
		log.Print("Resetting...")
		if err := car.Reset(); err != nil {
			http.Error(w, "could not reset", http.StatusInternalServerError)
		}
	})

	m.Get("/snapshot", func(w http.ResponseWriter, r *http.Request) {
		log.Print("Sending current snapshot")

		image := camera.CurrentImage()
		w.Write(image)
	})

	log.Print("Starting web server")
	go m.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit

	log.Print("All done")
}
