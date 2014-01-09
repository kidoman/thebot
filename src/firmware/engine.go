package main

import (
	"github.com/kid0m4n/go-rpi/util"
)

const (
	minSpeed     = 0
	quarterSpeed = 25
	halfSpeed    = 50
	maxSpeed     = 100
)

type Engine interface {
	RunAt(speed int) error
	Stop() error
}

type nullEngine struct {
}

func (*nullEngine) RunAt(_ int) error {
	return nil
}

func (*nullEngine) Stop() error {
	return nil
}

var NullEngine = &nullEngine{}

const (
	minAnalogValue = 0
	maxAnalogValue = 255
)

type pwm interface {
	SetAnalog(channel int, value byte) error
}

type engine struct {
	channel int

	pwm pwm
}

func NewEngine(channel int, pwm pwm) Engine {
	return &engine{
		channel: channel,
		pwm:     pwm,
	}
}

// RunAt sets the engine speed. Valid values at [0-100]
func (e *engine) RunAt(speed int) error {
	if speed < minSpeed {
		speed = minSpeed
	}
	if speed > maxSpeed {
		speed = maxSpeed
	}

	value := util.Map(int64(speed), minSpeed, maxSpeed, minAnalogValue, maxAnalogValue)

	return e.pwm.SetAnalog(e.channel, byte(value))
}

func (e *engine) Stop() error {
	return e.RunAt(0)
}
