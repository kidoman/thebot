package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/codegangsta/martini"
	"github.com/gorilla/websocket"
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

	setOrientation := func(speedStr, angleStr string) (code int, err error) {
		speed, err := strconv.Atoi(speedStr)
		if err != nil {
			return http.StatusBadRequest, errors.New("speed not valid")
		}
		angle, err := strconv.Atoi(angleStr)
		if err != nil {
			return http.StatusBadRequest, errors.New("angle not valid")
		}
		log.Printf("Received orientation %v, %v", angle, speed)
		if err = car.Turn(angle); err != nil {
			return http.StatusInternalServerError, err
		}
		if err = car.Speed(speed); err != nil {
			return http.StatusInternalServerError, err
		}

		return 0, nil
	}

	m := martini.Classic()

	m.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Upgrade(w, r, nil, 1024*1024, 1024)
		if _, ok := err.(websocket.HandshakeError); ok {
			http.Error(w, "Not a websocket handshake", http.StatusBadRequest)
			return
		} else if err != nil {
			log.Print(err)
			return
		}

		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if messageType == websocket.TextMessage {
				msg := string(p)
				parts := strings.Split(msg, ",")
				speedStr, angleStr := parts[0], parts[1]

				_, err = setOrientation(speedStr, angleStr)
				if err != nil {
					log.Print(err)
				}
			}
		}
	})

	m.Post("/speed/:speed/angle/:angle", func(w http.ResponseWriter, params martini.Params) {
		code, err := setOrientation(params["speed"], params["angle"])

		if err != nil {
			http.Error(w, err.Error(), code)
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
