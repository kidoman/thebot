package main

import (
	"github.com/kid0m4n/go-rpi/controller/pca9685"
	"github.com/kid0m4n/go-rpi/util"
)

const (
	minSpeed  = 0
	halfSpeed = 50
	maxSpeed  = 100
)

type engine struct {
	channel int

	pwm *pca9685.PCA9685
}

func NewEngine(channel int, pwm *pca9685.PCA9685) *engine {
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

	onTime := 0
	offTime := util.Map(int64(speed), minSpeed, maxSpeed, 0, 4096)

	return e.pwm.SetPwm(e.channel, onTime, int(offTime))
}

func (e *engine) Stop() error {
	return e.RunAt(0)
}
