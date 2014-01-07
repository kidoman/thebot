package main

import (
	"github.com/kid0m4n/go-rpi/motion/servo"
)

type frontWheel struct {
	servo *servo.Servo
}

func (fw *frontWheel) Turn(angle int) error {
	servoAngle := angle + 90
	return fw.servo.SetAngle(servoAngle)
}
