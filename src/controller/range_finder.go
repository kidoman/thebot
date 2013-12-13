package main

import (
	"github.com/kid0m4n/go-rpi/sensor/us020"
)

type RangeFinder interface {
	us020.US020
}

type rangeFinder struct {
	sensor us020.US020
}

func NewRangeFinder(e, t int) RangeFinder {
	var rf rangeFinder

	rf.sensor = us020.New(e, t)

	return &rf
}

func (rf *rangeFinder) Distance() (float64, error) {
	return rf.sensor.Distance()
}
