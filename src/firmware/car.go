package main

import (
	"log"
	"math"
	"sync"
	"time"

	"github.com/kid0m4n/go-rpi/i2c"
	"github.com/kid0m4n/go-rpi/util"
)

const (
	rangeCheckDelay = 100
)

type Car interface {
	Velocity(speed, angle int) error

	CurrentImage() []byte
	Heading() (heading float64, err error)
	DistanceInFront() (float64, error)

	Turn(swing int) error
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

func (nullCar) Turn(_ int) error {
	return nil
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
	gyro       Gyroscope
	frontWheel FrontWheel
	engine     Engine

	disable chan *disableInstruction
	control chan *controlInstruction
}

func NewCar(bus i2c.Bus, camera Camera, compass Compass, rf RangeFinder, gyro Gyroscope, frontWheel FrontWheel, engine Engine) Car {
	c := &car{
		bus:        bus,
		camera:     camera,
		compass:    compass,
		rf:         rf,
		gyro:       gyro,
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
				<-done

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
	return c.velocity(minSpeed, straight)
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

func (c *car) Turn(swing int) (err error) {
	// Stop the car. Known state
	if err = c.Velocity(minSpeed, straight); err != nil {
		return
	}
	time.Sleep(1 * time.Second)

	// Give a inertial boost.
	if err = c.Velocity(halfSpeed, straight); err != nil {
		return
	}
	time.Sleep(1 * time.Second)

	orientations, err := c.gyro.Orientations()
	if err != nil {
		return
	}

	c.gyro.Start()
	defer c.gyro.Stop()

	midPoint := float64(swing / 2)
	mult := float64(swing) / math.Abs(float64(swing))

	defer c.Velocity(minSpeed, straight)

	for {
		select {
		case orientation := <-orientations:
			currentZ := -orientation.Z
			left := math.Abs(currentZ - float64(swing))
			if left < 1 {
				return
			}
			var angle int64
			if math.Abs(currentZ) < math.Abs(midPoint) {
				angle = util.Map(int64(currentZ), 0, int64(midPoint), 0, int64(maxTurn*mult))
			} else {
				angle = util.Map(int64(currentZ), int64(midPoint), int64(swing), int64(maxTurn*mult), 0)
			}
			c.Velocity(halfSpeed, int(angle))
		}
	}

	return
}

func (c *car) PointTo(angle int) error {
	return nil
}
