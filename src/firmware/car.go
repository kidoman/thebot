package main

import (
	"log"
	"sync"
	"time"

	"github.com/kid0m4n/go-rpi/i2c"
)

const (
	rangeCheckDelay = 100
)

type Car interface {
	Velocity(speed, angle int) error

	CurrentImage() []byte
	Heading() (heading float64, err error)
	DistanceInFront() (float64, error)
}

type nullCar struct {
}

func (nullCar) Velocity(_, _ int) error {
	return nil
}

func (nullCar) CurrentImage() []byte {
	return nil
}

func (nullCar) Heading() (float64, error) {
	return 0, nil
}

func (nullCar) DistanceInFront() (float64, error) {
	return 0, nil
}

var NullCar = &nullCar{}

type controlInstruction struct {
	speed, angle int

	done chan error
}

type disableInstruction struct {
	disable  bool
	distance float64

	done chan error
}

type car struct {
	bus i2c.Bus

	mu sync.RWMutex

	camera     Camera
	compass    Compass
	rf         RangeFinder
	frontWheel *frontWheel
	engine     *engine

	disable chan *disableInstruction
	control chan *controlInstruction
}

func NewCar(bus i2c.Bus, camera Camera, compass Compass, rf RangeFinder, frontWheel *frontWheel, engine *engine) Car {
	c := &car{
		bus:        bus,
		camera:     camera,
		compass:    compass,
		rf:         rf,
		frontWheel: frontWheel,
		engine:     engine,
		disable:    make(chan *disableInstruction),
		control:    make(chan *controlInstruction),
	}
	go c.loop()
	return c
}

func (c *car) loop() {
	var rangeTimer <-chan time.Time
	resetRangeTimer := func() {
		rangeTimer = time.After(rangeCheckDelay * time.Millisecond)
	}
	resetRangeTimer()
	rangingDone := make(chan struct{})
	disabled := false

	for {
		select {
		case <-rangeTimer:
			rangeTimer = nil
			go func() {
				dist, err := c.rf.Distance()
				if err != nil {
					panic(err)
				}
				done := make(chan error)
				if dist < float64(*threshold) {
					c.disable <- &disableInstruction{true, dist, done}
				} else {
					c.disable <- &disableInstruction{false, dist, done}
				}

				rangingDone <- struct{}{}
			}()
		case inst := <-c.disable:
			if disabled == inst.disable {
				inst.done <- nil
				continue
			}
			var err error
			disabled = inst.disable
			if disabled {
				log.Printf("car: collision %.0f cm ahead, stopping car", inst.distance)
				err = c.stop()
			} else {
				log.Printf("car: obstruction cleared till %.0f cm, enabled car", inst.distance)
			}
			inst.done <- err
		case inst := <-c.control:
			var err error
			if !disabled {
				err = c.velocity(inst.speed, inst.angle)
			}
			inst.done <- err
		case <-rangingDone:
			resetRangeTimer()
		}
	}
}

func (c *car) stop() error {
	return c.velocity(0, 90)
}

func (c *car) velocity(speed, angle int) (err error) {
	log.Printf("car: setting speed to %v", speed)
	if err = c.engine.RunAt(speed); err != nil {
		return
	}
	log.Printf("car: setting angle to %v", angle)
	err = c.frontWheel.Turn(angle)
	return
}

func (c *car) Velocity(speed, angle int) error {
	done := make(chan error)
	c.control <- &controlInstruction{speed, angle, done}
	return <-done
}

func (c *car) Disable() error {
	done := make(chan error)
	c.disable <- &disableInstruction{disable: true, done: done}
	return <-done
}

func (c *car) CurrentImage() []byte {
	return c.camera.CurrentImage()
}

func (c *car) Heading() (float64, error) {
	return c.compass.Heading()
}

func (c *car) DistanceInFront() (float64, error) {
	return c.rf.Distance()
}
