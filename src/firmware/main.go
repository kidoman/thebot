package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/kid0m4n/go-rpi/controller/pca9685"
	"github.com/kid0m4n/go-rpi/controller/servoblaster"
	"github.com/kid0m4n/go-rpi/i2c"
	"github.com/kid0m4n/go-rpi/motion/servo"
	"github.com/kid0m4n/go-rpi/sensor/l3gd20"
)

var (
	threshold        = flag.Int("threshold", 30, "safe distance to stop the car")
	camWidth         = flag.Int("camw", 640, "width of the captured camera image")
	camHeight        = flag.Int("camh", 480, "height of the captured camera image")
	camFps           = flag.Int("fps", 4, "fps for camera")
	fakeCar          = flag.Bool("fcr", false, "fake the car")
	fakeCam          = flag.Bool("fcm", false, "fake the camera")
	fakeCompass      = flag.Bool("fcp", false, "fake the compass")
	fakeEngine       = flag.Bool("fe", false, "fake the engine")
	fakeRangeFinder  = flag.Bool("frf", false, "fake the range finder")
	fakeFrontWheel   = flag.Bool("ffw", false, "fake the front wheel")
	fakeGyro         = flag.Bool("fg", false, "fake the gyro")
	echoPinNumber    = flag.Int("epn", 10, "GPIO pin connected to the echo pad")
	triggerPinNumber = flag.Int("tpn", 9, "GPIO pin connected to the trigger pad")
)

func main() {
	log.Print("main: starting up")

	flag.Parse()

	var cam Camera = NullCamera
	if !*fakeCam {
		cam = NewCamera(*camWidth, *camHeight, *camFps)
	}
	defer cam.Close()
	cam.Run()

	var comp Compass = NullCompass
	if !*fakeCompass {
		comp = NewCompass(i2c.Default)
	}
	defer comp.Close()
	comp.Run()

	var rf RangeFinder = NullRangeFinder
	if !*fakeRangeFinder {
		rf = NewRangeFinder(*echoPinNumber, *triggerPinNumber)
	}
	defer rf.Close()

	var fw FrontWheel = NullFrontWheel
	if !*fakeFrontWheel {
		sb := servoblaster.New()
		defer sb.Close()

		servo := servo.New(sb, 0)
		fw = &frontWheel{servo}
	}
	defer fw.Turn(0)

	var engine Engine = NullEngine
	if !*fakeEngine {
		pwmMotor := pca9685.New(i2c.Default, 0x41, 1000)
		defer pwmMotor.Close()
		engine = NewEngine(15, pwmMotor)
	}
	defer engine.Stop()

	var gyro Gyroscope = NullGyroscope
	if !*fakeGyro {
		gyro = NewGyroscope(i2c.Default, l3gd20.R250DPS)
	}
	defer gyro.Close()

	var car Car = NullCar
	if !*fakeCar {
		car = NewCar(i2c.Default, cam, comp, rf, gyro, fw, engine)
	}

	ws := NewWebServer(car)
	ws.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit

	log.Print("main: all done")
}
