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
	turnPollDelay   = 50
)

type Car interface {
	Velocity(speed, angle int) error

	CurrentImage() []byte
	Heading() (heading float64, err error)
	DistanceInFront() (float64, error)

	Turn(swing int) error
	PointTo(angle int) error

	Close()
}

type nullCar struct {
}

func (*nullCar) Velocity(_, _ int) error {
	return nil
}

func (*nullCar) CurrentImage() []byte {
	return nil
}

func (*nullCar) Heading() (float64, error) {
	return 0, nil
}

func (*nullCar) DistanceInFront() (float64, error) {
	return 0, nil
}

func (*nullCar) Turn(_ int) error {
	return nil
}

func (*nullCar) PointTo(angle int) error {
	return nil
}

func (*nullCar) Close() {
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

	curSpeed, curAngle int

	disable chan *disableInstruction
	control chan *controlInstruction

	closing chan chan struct{}
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
		closing:    make(chan chan struct{}),
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
	ranging := false

	for {
		select {
		case waitc := <-c.closing:
			if ranging {
				<-rangingDone
			}
			waitc <- struct{}{}
			return
		case <-rangeTimer:
			rangeTimer = nil
			ranging = true
			go func() {
				dist, err := c.rf.Distance()
				if err != nil {
					rangingDone <- struct{}{}
					return
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
			ranging = false
		}
	}
}

func (c *car) stop() (err error) {
	if err = c.velocity(minSpeed, stopAngle); err != nil {
		return
	}
	time.Sleep(200 * time.Millisecond)
	if err = c.velocity(minSpeed, -stopAngle); err != nil {
		return
	}
	time.Sleep(500 * time.Millisecond)
	if err = c.velocity(minSpeed, straight); err != nil {
		return
	}
	return
}

func (c *car) velocity(speed, angle int) (err error) {
	if speed != c.curSpeed {
		log.Printf("car: setting speed to %v", speed)
		if err = c.engine.RunAt(speed); err != nil {
			return
		}
		c.curSpeed = speed
	}
	if angle != c.curAngle {
		log.Printf("car: setting angle to %v", angle)
		if err = c.frontWheel.Turn(angle); err != nil {
			return
		}
		c.curAngle = angle
	}
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
	time.Sleep(500 * time.Millisecond)
	c.gyro.Start()
	defer c.gyro.Stop()
	time.Sleep(500 * time.Millisecond)

	// Give a inertial boost.
	if err = c.Velocity(halfSpeed, straight); err != nil {
		return
	}

	orientations, err := c.gyro.Orientations()
	if err != nil {
		return
	}

	midPoint := float64(swing) * 0.4
	mult := float64(swing) / math.Abs(float64(swing))

	defer c.Velocity(minSpeed, straight)

	log.Print("car: starting to turn")
	defer log.Print("car: stopped turning")

	var min, max int
	if swing < 0 {
		min = swing
		max = 0
	} else {
		min = 0
		max = swing
	}

	clamp := func(v int) int {
		if v < min {
			return min
		}
		if v > max {
			return max
		}
		return v
	}

	for {
		timer := time.After(turnPollDelay * time.Millisecond)

		select {
		case <-timer:
			orientation := <-orientations
			currentZ := -int(orientation.Z)
			clampedZ := clamp(currentZ)

			left := math.Abs(float64(clampedZ - swing))
			if left < minTurn {
				return
			}
			var angle int64
			if math.Abs(float64(clampedZ)) < math.Abs(midPoint) {
				angle = util.Map(int64(clampedZ), 0, int64(midPoint), minTurn, int64(maxTurningAngle*mult))
			} else {
				angle = util.Map(int64(clampedZ), int64(midPoint), int64(swing), int64(maxTurningAngle*mult), minTurn)
			}
			if err = c.Velocity(quarterSpeed, int(angle)); err != nil {
				return
			}
		}
	}

	return
}

func (c *car) PointTo(angle int) (err error) {
	// Stop the car. Known state
	if err = c.Velocity(minSpeed, straight); err != nil {
		return
	}
	time.Sleep(1000 * time.Millisecond)

	heading, err := c.compass.Heading()
	if err != nil {
		return
	}

	swing := angle - int(heading)

	log.Printf("car: current heading %v, turning by %v", heading, swing)

	return c.Turn(swing)
}

func (c *car) Close() {
	waitc := make(chan struct{})
	c.closing <- waitc
	<-waitc
}
