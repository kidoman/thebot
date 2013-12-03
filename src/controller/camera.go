package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"sync"
	"time"
)

const filename = "/tmp/image.jpg"

type Camera interface {
	Run()
	Close()
	CurrentImage() []byte
}

type nullCamera struct {
}

func (nullCamera) Run() {
}

func (nullCamera) Close() {
}

func (nullCamera) CurrentImage() []byte {
	wd, err := os.Getwd()
	if err != nil {
		log.Panic(err)
	}
	bytes, err := ioutil.ReadFile(path.Join(wd, "public/sample.jpeg"))
	if err != nil {
		log.Print(err)
	}
	return bytes
}

var NullCamera = &nullCamera{}

type camera struct {
	w, h, delay int

	currentImage []byte
	cimu         *sync.RWMutex

	quit chan bool
}

func NewCamera(w, h, fps int) Camera {
	var c camera

	c.currentImage = make([]byte, 0)
	c.cimu = &sync.RWMutex{}
	c.w, c.h = w, h
	c.delay = 1000 / fps
	c.quit = make(chan bool)

	return &c
}

func (c *camera) Run() {
	log.Print("Starting camera capture")

	go func() {
		conv := func(i int) string {
			return strconv.Itoa(i)
		}
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

func (c *camera) Close() {
	log.Print("Cleaning camera module")

	c.quit <- true
}

func (c *camera) CurrentImage() []byte {
	c.cimu.RLock()
	defer c.cimu.RUnlock()

	return c.currentImage
}
