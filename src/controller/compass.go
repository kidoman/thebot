package main

import (
	"github.com/kid0m4n/go-rpi/i2c"
	"github.com/kid0m4n/go-rpi/sensor/lsm303"
)

type Compass interface {
	lsm303.LSM303
}

type compass struct {
	sensor lsm303.LSM303
}

func NewCompass(bus i2c.Bus) Compass {
	var comp compass

	comp.sensor = lsm303.New(bus)

	return &comp
}

func (comp *compass) Run() error {
	return comp.sensor.Run()
}

func (comp *compass) Heading() (float64, error) {
	return comp.sensor.Heading()
}

func (comp *compass) Close() error {
	return comp.sensor.Close()
}

func (comp *compass) SetPollDelay(delay int) {
	comp.sensor.SetPollDelay(delay)
}
