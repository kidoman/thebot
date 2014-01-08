package main

import (
	"github.com/kid0m4n/go-rpi/motion/servo"
)

const (
	straight = 0
	left     = -90
	right    = 90

	maxTurn = 30
)

type FrontWheel interface {
	Turn(angle int) error
}

type nullFrontWheel struct {
}

var NullFrontWheel = &nullFrontWheel{}

func (*nullFrontWheel) Turn(_ int) error {
	return nil
}

type frontWheel struct {
	servo *servo.Servo
}

func (fw *frontWheel) Turn(angle int) error {
	servoAngle := angle + 90
	return fw.servo.SetAngle(servoAngle)
}
