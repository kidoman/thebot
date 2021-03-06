package main

import (
	"flag"
	"os"
	"os/signal"

	"github.com/golang/glog"
	"github.com/kidoman/embd"
	"github.com/kidoman/embd/controller/pca9685"
	"github.com/kidoman/embd/controller/servoblaster"
	"github.com/kidoman/embd/motion/servo"
	"github.com/kidoman/embd/sensor/bmp180"
	"github.com/kidoman/embd/sensor/l3gd20"
)

var (
	i2cBusNo         = flag.Int("bus", 1, "i2c bus to use")
	threshold        = flag.Int("threshold", 50, "safe distance to stop the car")
	camWidth         = flag.Int("camw", 640, "width of the captured camera image")
	camHeight        = flag.Int("camh", 480, "height of the captured camera image")
	camTurnImage     = flag.Int("camt", 270, "turn the image by these many degrees")
	camFps           = flag.Int("fps", 2, "fps for camera")
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
	glog.Info("main: starting up")

	flag.Parse()

	var car Car = NullCar
	if !*fakeCar {
		if err := embd.InitI2C(); err != nil {
			panic(err)
		}
		defer embd.CloseI2C()

		bus := embd.NewI2CBus(byte(*i2cBusNo))

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

			if err := embd.InitGPIO(); err != nil {
				panic(err)
			}
			defer embd.CloseGPIO()

			echoPin, err := embd.NewDigitalPin(*echoPinNumber)
			if err != nil {
				panic(err)
			}
			triggerPin, err := embd.NewDigitalPin(*triggerPinNumber)
			if err != nil {
				panic(err)
			}

			rf = NewRangeFinder(echoPin, triggerPin, thermometer)
		}
		defer rf.Close()

		var fw FrontWheel = NullFrontWheel
		if !*fakeFrontWheel {
			sb := servoblaster.New()
			defer sb.Close()

			pwm := sb.Channel(*sbChannel)

			servo := servo.New(pwm)
			fw = &frontWheel{servo}
		}
		defer fw.Turn(0)

		var engine Engine = NullEngine
		if !*fakeEngine {
			ctrl := pca9685.New(bus, 0x41)
			defer ctrl.Close()

			pwm := ctrl.AnalogChannel(15)

			engine = NewEngine(pwm)
		}
		defer engine.Stop()

		var gyro Gyroscope = NullGyroscope
		if !*fakeGyro {
			gyro = NewGyroscope(bus, l3gd20.R250DPS)
		}
		defer gyro.Close()

		car = NewCar(bus, cam, comp, rf, gyro, fw, engine)
	}
	defer car.Close()

	ws := NewWebServer(car)
	ws.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit

	glog.Info("main: all done")
}
