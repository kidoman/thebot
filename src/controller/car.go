package main

import (
	"log"
	"sync"

	"github.com/kid0m4n/go-rpi/i2c"
)

const (
	Servo = 0x53
	Motor = 0x4D
	Reset = 0x52
)

type Car interface {
	Turn(angle int) error
	Speed(speed int) error
	Orientation() (speed, angle int)
	Reset() error
}

type nullCar struct {
}

func (nullCar) Turn(_ int) error {
	return nil
}

func (nullCar) Speed(_ int) error {
	return nil
}

func (nullCar) Orientation() (speed, angle int) {
	return 0, 0
}

func (nullCar) Reset() error {
	return nil
}

var NullCar = &nullCar{}

func NewCar(bus i2c.Bus, addr byte) Car {
	return &car{addr: addr, bus: bus}
}

type car struct {
	addr byte
	bus  i2c.Bus

	curAngle, curSpeed int

	mu sync.RWMutex
}

func (c *car) Turn(angle int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cmd := []byte{Servo, byte(angle)}
	if err := c.bus.WriteBytes(c.addr, cmd); err != nil {
		return err
	}

	log.Printf("Set the angle to %v", angle)

	c.curAngle = angle

	return nil
}

func (c *car) Speed(speed int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cmd := []byte{Motor, byte(speed)}
	if err := c.bus.WriteBytes(c.addr, cmd); err != nil {
		return err
	}

	log.Printf("Set the speed to %v", speed)

	c.curSpeed = speed

	return nil
}

func (c *car) Orientation() (speed, angle int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.curSpeed, c.curAngle
}

func (c *car) Reset() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.bus.WriteByte(c.addr, Reset); err != nil {
		return err
	}

	log.Print("Reset the device")

	c.curAngle, c.curSpeed = 0, 0

	return nil
}
