package main

import (
	"log"
	"sync"
)

const (
	Servo = 0x53
	Motor = 0x4D
)

var bus *I2CBus

func init() {
	var err error

	bus, err = Bus(1)
	if err != nil {
		panic(err)
	}
}

type Car struct {
	addr byte

	curAngle, curSpeed int

	mu *sync.RWMutex
}

func NewCar(addr byte) *Car {
	return &Car{addr: addr, mu: &sync.RWMutex{}}
}

func (c *Car) Turn(angle int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.curAngle == angle {
		return nil
	}

	cmd := []byte{Servo, byte(angle)}
	if err := bus.WriteBytes(c.addr, cmd); err != nil {
		return err
	}

	log.Printf("Set the angle to %v", angle)

	c.curAngle = angle

	return nil
}

func (c *Car) Speed(speed int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.curSpeed == speed {
		return nil
	}

	cmd := []byte{Motor, byte(speed)}
	if err := bus.WriteBytes(c.addr, cmd); err != nil {
		return err
	}

	log.Printf("Set the speed to %v", speed)

	c.curSpeed = speed

	return nil
}

func (c *Car) Orientation() (speed, angle int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.curSpeed, c.curAngle
}
