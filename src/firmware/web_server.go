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
	m   *martini.ClassicMartini
	car Car
}

func NewWebServer(car Car) *WebServer {
	var ws WebServer

	ws.m = martini.Classic()
	ws.car = car

	ws.registerHandlers()

	return &ws
}

func (ws *WebServer) registerHandlers() {
	ws.m.Get("/ws", ws.wsHandler)
	ws.m.Post("/speed/:speed/angle/:angle", ws.setSpeedAndAngle)
	ws.m.Get("/distance", ws.distance)
	ws.m.Get("/snapshot", ws.snapshot)
	ws.m.Post("/swing/:swing", ws.swing)
}

func (ws *WebServer) Run() {
	log.Print("api: starting server")

	go ws.m.Run()
}

func (ws *WebServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Upgrade(w, r, nil, 1024*1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "api: not a websocket handshake", http.StatusBadRequest)
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

			_, err = ws.setVelocity(speedStr, angleStr)
			if err != nil {
				log.Print(err)
			}
		}
	}
}

func (ws *WebServer) setSpeedAndAngle(w http.ResponseWriter, params martini.Params) {
	code, err := ws.setVelocity(params["speed"], params["angle"])

	if err != nil {
		http.Error(w, err.Error(), code)
	}
}

func (ws *WebServer) distance(w http.ResponseWriter) string {
	distance, err := ws.car.DistanceInFront()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return fmt.Sprintf("%v", distance)
}

func (ws *WebServer) snapshot(w http.ResponseWriter) {
	log.Print("api: sending current snapshot")

	image := ws.car.CurrentImage()
	w.Write(image)
}

func (ws *WebServer) setVelocity(speedStr, angleStr string) (code int, err error) {
	speed, err := strconv.Atoi(speedStr)
	if err != nil {
		return http.StatusBadRequest, errors.New("speed not valid")
	}
	angle, err := strconv.Atoi(angleStr)
	if err != nil {
		return http.StatusBadRequest, errors.New("angle not valid")
	}
	log.Printf("api: received orientation %v, %v", angle, speed)
	if err = ws.car.Velocity(speed, angle); err != nil {
		return http.StatusInternalServerError, err
	}
	return 0, nil
}

func (ws *WebServer) swing(w http.ResponseWriter, params martini.Params) {
	swing, err := strconv.Atoi(params["swing"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err = ws.car.Turn(swing); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
