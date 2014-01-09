package main

import (
	"math"

	"github.com/kid0m4n/go-rpi/motion/servo"
)

const (
	straight  = 0
	stopAngle = 40
	left      = -90
	right     = 90

	minTurn         = 5
	maxTurn         = 40
	maxTurningAngle = 20
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
	if math.Abs(float64(angle)) > maxTurn {
		angle = maxTurn * int(float64(angle)/math.Abs(float64(angle)))
	}
	servoAngle := angle + 90 + *fwCorrection
	return fw.servo.SetAngle(servoAngle)
}
