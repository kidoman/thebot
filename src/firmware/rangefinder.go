package main

import (
	"github.com/kidoman/embd"
	"github.com/kidoman/embd/sensor/us020"
)

const (
	maxDistance = 999
)

type RangeFinder interface {
	Distance() (float64, error)
	Close() error
}

type nullRangeFinder struct {
}

func (*nullRangeFinder) Distance() (float64, error) {
	return maxDistance, nil
}

func (*nullRangeFinder) Close() error {
	return nil
}

var NullRangeFinder = &nullRangeFinder{}

type rangeFinder struct {
	*us020.US020
}

func NewRangeFinder(echoPin, triggerPin embd.DigitalPin, thermometer us020.Thermometer) RangeFinder {
	return &rangeFinder{us020.New(echoPin, triggerPin, thermometer)}
}
