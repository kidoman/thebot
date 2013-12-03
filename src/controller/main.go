package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
)

var (
	camWidth       = flag.Int("camw", 640, "width of the captured camera image")
	camHeight      = flag.Int("camh", 480, "height of the captured camera image")
	camFps         = flag.Int("fps", 4, "fps for camera")
	arduinoAddrStr = flag.String("addr", "0x50", "arduino i2c address")
	fakeCar        = flag.Bool("fcr", false, "fake the car")
	fakeCam        = flag.Bool("fcm", false, "fake the camera")
)

func main() {
	log.Print("Starting up...")

	flag.Parse()

	var cam Camera = NullCamera
	if !*fakeCam {
		cam = NewCamera(*camWidth, *camHeight, *camFps)
	}
	defer cam.Close()
	cam.Run()

	arduinoAddr, err := strconv.ParseInt(*arduinoAddrStr, 0, 0)
	if err != nil {
		log.Fatalf("Could not parse %q for arduino i2c address", *arduinoAddrStr)
	}
	var car Car = NullCar
	if !*fakeCar {
		bus, err := Bus(1)
		if err != nil {
			panic(err)
		}
		car = NewCar(bus, byte(arduinoAddr))
	}

	ws := NewWebServer(car, cam)
	ws.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit

	log.Print("All done")
}
