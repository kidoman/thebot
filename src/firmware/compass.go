package main

import (
	"github.com/kidoman/embd"
	"github.com/kidoman/embd/sensor/lsm303"
)

type Compass interface {
	Heading() (float64, error)
	Run() error
	Close() error
}

type nullCompass struct {
}

func (*nullCompass) Heading() (float64, error) {
	return 0, nil
}

func (*nullCompass) Run() error {
	return nil
}

func (*nullCompass) Close() error {
	return nil
}

var NullCompass = &nullCompass{}

type compass struct {
	*lsm303.LSM303
}

func NewCompass(bus embd.I2CBus) Compass {
	return &compass{lsm303.New(bus)}
}
