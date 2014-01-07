package main

import (
	"github.com/kid0m4n/go-rpi/sensor/us020"
)

type RangeFinder interface {
	us020.US020
}

type rangeFinder struct {
	us020.US020
}

func NewRangeFinder(e, t int) RangeFinder {
	return &rangeFinder{us020.New(e, t)}
}
