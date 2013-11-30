package main

import (
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

const filename = "/tmp/image.jpg"

type Camera struct {
	w, h, delay int

	currentImage []byte
	cimu         *sync.RWMutex

	quit chan bool
}

func NewCamera(w, h, fps int) *Camera {
	camera := &Camera{}
	camera.currentImage = make([]byte, 0)
	camera.cimu = &sync.RWMutex{}
	camera.w, camera.h = w, h
	camera.delay = 1000 / fps
	camera.quit = make(chan bool)

	return camera
}

func (c *Camera) Run() {
	log.Print("Starting camera capture")

	conv := func(i int) string {
		return strconv.Itoa(i)
	}

	go func() {
		timer := time.Tick(time.Duration(c.delay) * time.Millisecond)

		for {
			select {
			case <-timer:
				log.Print("Taking snapshot")

				cmd := exec.Command("raspistill", "-n", "-w", conv(c.w), "-h", conv(c.h), "-t", "30", "-o", filename)
				err := cmd.Run()
				if err != nil {
					log.Print("Could not take a snapshot")
					continue
				}
				newImage, err := ioutil.ReadFile(filename)
				if err != nil {
					panic(err)
				}

				c.cimu.Lock()
				c.currentImage = newImage
				c.cimu.Unlock()
			case <-c.quit:
				return
			}
		}
	}()
}

func (c *Camera) Close() {
	log.Print("Cleaning camera module")

	c.quit <- true
}

func (c *Camera) CurrentImage() []byte {
	c.cimu.RLock()
	defer c.cimu.RUnlock()

	return c.currentImage
}
