package main

import (
	"testing"
)

func TestCheckBit(t *testing.T) {
	if bit := checkBit(2, 1); bit != true {
		t.Errorf("checkBit(2, 1) = %v, want true", bit)
	}
	if bit := checkBit(2, 0); bit != false {
		t.Errorf("checkBit(2, 0) = %v, want false", bit)
	}
}
