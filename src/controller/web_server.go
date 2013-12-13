package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/codegangsta/martini"
	"github.com/gorilla/websocket"
)

type WebServer struct {
	m    *martini.ClassicMartini
	car  Car
	cam  Camera
	comp Compass
	rf   RangeFinder
}

func NewWebServer(car Car, cam Camera, comp Compass, rf RangeFinder) *WebServer {
	var ws WebServer

	ws.m = martini.Classic()
	ws.car = car
	ws.cam = cam
	ws.comp = comp
	ws.rf = rf

	ws.registerHandlers()

	return &ws
}

func (ws *WebServer) registerHandlers() {
	ws.m.Get("/ws", ws.wsHandler)
	ws.m.Post("/speed/:speed/angle/:angle", ws.setSpeedAndAngle)
	ws.m.Get("/orientation", ws.orientation)
	ws.m.Get("/distance", ws.distance)
	ws.m.Get("/snapshot", ws.snapshot)
	ws.m.Post("/reset", ws.reset)
}

func (ws *WebServer) Run() {
	log.Print("Starting web server")

	go ws.m.Run()
}

func (ws *WebServer) wsHandler(w http.ResponseWriter, r *http.Request) {
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

			_, err = ws.setOrientation(speedStr, angleStr)
			if err != nil {
				log.Print(err)
			}
		}
	}
}

func (ws *WebServer) setSpeedAndAngle(w http.ResponseWriter, params martini.Params) {
	code, err := ws.setOrientation(params["speed"], params["angle"])

	if err != nil {
		http.Error(w, err.Error(), code)
	}
}

func (ws *WebServer) orientation(w http.ResponseWriter) string {
	heading, err := ws.comp.Heading()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return fmt.Sprintf("%v", heading)
}

func (ws *WebServer) distance(w http.ResponseWriter) string {
	distance, err := ws.rf.Distance()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return fmt.Sprintf("%v", distance)
}

func (ws *WebServer) snapshot(w http.ResponseWriter) {
	log.Print("Sending current snapshot")

	image := ws.cam.CurrentImage()
	w.Write(image)
}

func (ws *WebServer) reset(w http.ResponseWriter) {
	log.Print("Resetting...")
	if err := ws.car.Reset(); err != nil {
		http.Error(w, "could not reset", http.StatusInternalServerError)
	}
}

func (ws *WebServer) setOrientation(speedStr, angleStr string) (code int, err error) {
	speed, err := strconv.Atoi(speedStr)
	if err != nil {
		return http.StatusBadRequest, errors.New("speed not valid")
	}
	angle, err := strconv.Atoi(angleStr)
	if err != nil {
		return http.StatusBadRequest, errors.New("angle not valid")
	}
	log.Printf("Received orientation %v, %v", angle, speed)
	if err = ws.car.Speed(speed); err != nil {
		return http.StatusInternalServerError, err
	}
	if err = ws.car.Turn(angle); err != nil {
		return http.StatusInternalServerError, err
	}
	return 0, nil
}
