package main

import (
	"github.com/kid0m4n/go-rpi/i2c"
	"github.com/kid0m4n/go-rpi/sensor/l3gd20"
)

type gyroscope struct {
	*l3gd20.L3GD20
}

func NewGyroscope(bus i2c.Bus, rng *l3gd20.Range) *gyroscope {
	return &gyroscope{
		l3gd20.New(bus, rng),
	}
}
