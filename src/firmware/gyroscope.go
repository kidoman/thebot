package main

import (
	"github.com/kidoman/embd"
	"github.com/kidoman/embd/sensor/l3gd20"
)

type Gyroscope interface {
	Orientations() (<-chan l3gd20.Orientation, error)

	Start() error
	Stop() error
	Close() error
}

type nullGyroscope struct {
}

func (*nullGyroscope) Orientations() (<-chan l3gd20.Orientation, error) {
	return nil, nil
}

func (*nullGyroscope) Start() error {
	return nil
}

func (*nullGyroscope) Stop() error {
	return nil
}

func (*nullGyroscope) Close() error {
	return nil
}

var NullGyroscope = &nullGyroscope{}

type gyroscope struct {
	*l3gd20.L3GD20
}

func NewGyroscope(bus embd.I2CBus, rng *l3gd20.Range) Gyroscope {
	return &gyroscope{
		l3gd20.New(bus, rng),
	}
}
