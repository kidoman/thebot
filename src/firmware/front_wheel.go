package main

import (
	"github.com/kid0m4n/go-rpi/motion/servo"
)

const (
	straight = 0
	left = -90
	right = 90
	
	maxTurn = 30
)

type frontWheel struct {
	servo *servo.Servo
}

func (fw *frontWheel) Turn(angle int) error {
	servoAngle := angle + 90
	return fw.servo.SetAngle(servoAngle)
}
