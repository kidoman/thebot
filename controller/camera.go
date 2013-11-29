package main

import (
	"bytes"
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

	cmd := exec.Command("raspistill", "-n", "-vf", "-w", conv(c.w), "-h", conv(c.h), "-tl", conv(c.delay), "-t", "9999999", "-o", filename)
	err := cmd.Start()
	if err != nil {
		panic(nil)
	}

	var timer <-chan time.Time
	resetTimer := func() {
		timer = time.After(time.Duration(c.delay) * time.Millisecond)
	}
	resetTimer()

	go func() {
		for {
			select {
			case <-timer:
				c.cimu.Lock()
				newImage, err := ioutil.ReadFile(filename)
				if err != nil {
					panic(err)
				}
				if !bytes.Equal(c.currentImage, newImage) {
					c.currentImage = newImage
				}
				c.cimu.Unlock()
				resetTimer()
			case <-c.quit:
				cmd.Wait()
				c.quit <- true
				return
			}
		}
	}()
}

func (c *Camera) Close() {
	log.Print("Cleaning camera module")

	c.quit <- true
	<-c.quit
}

func (c *Camera) CurrentImage() []byte {
	c.cimu.RLock()
	defer c.cimu.RUnlock()

	return c.currentImage
}
