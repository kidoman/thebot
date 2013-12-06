package main

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockCamera struct {
	image []byte
}

func (m *mockCamera) Run() {
}

func (m *mockCamera) CurrentImage() []byte {
	return m.image
}

func (m *mockCamera) Close() {
}

type mockCar struct {
	speed, angle int

	resetErr error
}

func (m *mockCar) Turn(angle int) error {
	return nil
}

func (m *mockCar) Speed(speed int) error {
	return nil
}

func (m *mockCar) Orientation() (speed, angle int) {
	return m.speed, m.angle
}

func (m *mockCar) Reset() error {
	return m.resetErr
}

func TestOrietation(t *testing.T) {
	car := &mockCar{speed: 10, angle: 20}
	ws := &WebServer{car: car}
	res := ws.orientation()
	if res != "10, 20" {
		t.Fatalf("Expected orientation to be '10, 20', got '%v'", res)
	}
}

func TestSnapshot(t *testing.T) {
	image := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	cam := &mockCamera{image: image}
	ws := &WebServer{cam: cam}
	rec := httptest.NewRecorder()
	ws.snapshot(rec)
	if !bytes.Equal(rec.Body.Bytes(), image) {
		t.Fatal("Could not retrieve image from camera")
	}
}

func TestReset(t *testing.T) {
	car := &mockCar{}
	ws := &WebServer{car: car}
	rec := httptest.NewRecorder()
	ws.reset(rec)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status code %v, got %v", http.StatusOK, rec.Code)
	}
}

func TestResetError(t *testing.T) {
	car := &mockCar{resetErr: errors.New("error")}
	ws := &WebServer{car: car}
	rec := httptest.NewRecorder()
	ws.reset(rec)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %v, got %v", http.StatusInternalServerError, rec.Code)
	}
	if rec.Body.String() != "could not reset\n" {
		t.Errorf("Expected body to be %q, got %q", "could not reset\n", rec.Body.String())
	}
}

func TestSetOrientation(t *testing.T) {
	tests := []struct {
		speedStr, angleStr string
		code               int
		err                error
		errMessage         string
	}{
		{speedStr: "a", angleStr: "", code: http.StatusBadRequest, err: errors.New("speed not valid")},
	}

	car := &mockCar{}
	ws := &WebServer{car: car}

	for _, test := range tests {
		code, err := ws.setOrientation(test.speedStr, test.angleStr)
		if code != test.code {
			t.Errorf("Expected code %v, got %v", test.code, code)
		}
		if err != nil && test.err != nil && err.Error() != test.err.Error() {
			t.Errorf("Expected error %q, got %q", test.err.Error(), err.Error())
		}
	}
}
