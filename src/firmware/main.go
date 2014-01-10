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
	"github.com/kid0m4n/go-rpi/sensor/bmp180"
	"github.com/kid0m4n/go-rpi/sensor/l3gd20"
)

var (
	i2cBusNo         = flag.Int("bus", 1, "i2c bus to use")
	threshold        = flag.Int("threshold", 50, "safe distance to stop the car")
	camWidth         = flag.Int("camw", 640, "width of the captured camera image")
	camHeight        = flag.Int("camh", 480, "height of the captured camera image")
	camTurnImage     = flag.Int("camt", 270, "turn the image by these many degrees")
	camFps           = flag.Int("fps", 4, "fps for camera")
	echoPinNumber    = flag.Int("epn", 10, "GPIO pin connected to the echo pad")
	triggerPinNumber = flag.Int("tpn", 9, "GPIO pin connected to the trigger pad")
	sbChannel        = flag.Int("sbc", 0, "servo blaster channel to use for controlling front wheel")
	fwCorrection     = flag.Int("fwc", 0, "correction to be applied to the front wheel angle")

	fakeCar         = flag.Bool("fcr", false, "fake the car")
	fakeCam         = flag.Bool("fcm", false, "fake the camera")
	fakeCompass     = flag.Bool("fcp", false, "fake the compass")
	fakeEngine      = flag.Bool("fe", false, "fake the engine")
	fakeRangeFinder = flag.Bool("frf", false, "fake the range finder")
	fakeFrontWheel  = flag.Bool("ffw", false, "fake the front wheel")
	fakeGyro        = flag.Bool("fg", false, "fake the gyro")
)

func main() {
	log.Print("main: starting up")

	flag.Parse()

	bus, err := i2c.NewBus(byte(*i2cBusNo))
	if err != nil {
		log.Panic(err)
	}

	var cam Camera = NullCamera
	if !*fakeCam {
		cam = NewCamera(*camWidth, *camHeight, *camTurnImage, *camFps)
	}
	defer cam.Close()
	cam.Run()

	var comp Compass = NullCompass
	if !*fakeCompass {
		comp = NewCompass(bus)
	}
	defer comp.Close()

	var rf RangeFinder = NullRangeFinder
	if !*fakeRangeFinder {
		thermometer := bmp180.New(bus)
		defer thermometer.Close()

		rf = NewRangeFinder(*echoPinNumber, *triggerPinNumber, thermometer)
	}
	defer rf.Close()

	var fw FrontWheel = NullFrontWheel
	if !*fakeFrontWheel {
		sb := servoblaster.New()
		defer sb.Close()

		servo := servo.New(sb, *sbChannel)
		fw = &frontWheel{servo}
	}
	defer fw.Turn(0)

	var engine Engine = NullEngine
	if !*fakeEngine {
		pwmMotor := pca9685.New(bus, 0x41)
		defer pwmMotor.Close()
		engine = NewEngine(15, pwmMotor)
	}
	defer engine.Stop()

	var gyro Gyroscope = NullGyroscope
	if !*fakeGyro {
		gyro = NewGyroscope(bus, l3gd20.R250DPS)
	}
	defer gyro.Close()

	var car Car = NullCar
	if !*fakeCar {
		car = NewCar(bus, cam, comp, rf, gyro, fw, engine)
	}
	defer car.Close()

	ws := NewWebServer(car)
	ws.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit

	log.Print("main: all done")
}
