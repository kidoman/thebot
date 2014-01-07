package main

import (
	"github.com/kid0m4n/go-rpi/i2c"
	"github.com/kid0m4n/go-rpi/sensor/lsm303"
)

type Compass interface {
	lsm303.LSM303
}

type compass struct {
	lsm303.LSM303
}

func NewCompass(bus i2c.Bus) Compass {
	return &compass{lsm303.New(bus)}
}
