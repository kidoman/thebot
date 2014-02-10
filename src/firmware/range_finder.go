package main

import (
	"github.com/kidoman/embd/sensor/us020"
)

const (
	maxDistance = 999
)

type RangeFinder interface {
	Distance() (float64, error)
	Close()
}

type nullRangeFinder struct {
}

func (*nullRangeFinder) Distance() (float64, error) {
	return maxDistance, nil
}

func (*nullRangeFinder) Close() {
}

var NullRangeFinder = &nullRangeFinder{}

type rangeFinder struct {
	*us020.US020
}

func NewRangeFinder(e, t int, thermometer us020.Thermometer) RangeFinder {
	return &rangeFinder{us020.New(e, t, thermometer)}
}
