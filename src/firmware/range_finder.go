package main

import (
	"github.com/kid0m4n/go-rpi/sensor/us020"
)

const (
	maxDistance = 999
)

type RangeFinder interface {
	us020.US020
}

type rangeFinder struct {
	us020.US020
}

type nullRangeFinder struct {
}

func (*nullRangeFinder) Distance() (float64, error) {
	return maxDistance, nil
}

func (*nullRangeFinder) Close() {
}

var NullRangeFinder = &nullRangeFinder{}

func NewRangeFinder(e, t int) RangeFinder {
	return &rangeFinder{us020.New(e, t)}
}
