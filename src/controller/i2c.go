package main

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
)

const (
	Delay     = 20
	I2C_SLAVE = 0x0703
)

type i2c_smbus_ioctl_data struct {
	readWrite byte
	command   byte
	size      uint32
	data      uintptr
}

var busMap map[byte]*I2CBus
var busMapLock sync.Mutex

type I2CBus struct {
	file *os.File
	addr byte
	lock sync.Mutex
}

func init() {
	busMap = make(map[byte]*I2CBus)
}

func Bus(bus byte) (i2cbus *I2CBus, err error) {
	busMapLock.Lock()
	defer busMapLock.Unlock()

	if i2cbus = busMap[bus]; i2cbus == nil {
		i2cbus = new(I2CBus)
		if i2cbus.file, err = os.OpenFile(fmt.Sprintf("/dev/i2c-%v", bus), os.O_RDWR, os.ModeExclusive); err != nil {
			busMap[bus] = i2cbus
		}
	}

	return
}

func (i2cbus *I2CBus) setAddress(addr byte) (err error) {
	if addr != i2cbus.addr {
		if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, i2cbus.file.Fd(), I2C_SLAVE, uintptr(addr)); errno != 0 {
			err = syscall.Errno(errno)
			return
		}

		i2cbus.addr = addr
	}

	return
}

func (i2cbus *I2CBus) WriteByte(addr, value byte) (err error) {
	i2cbus.lock.Lock()
	defer i2cbus.lock.Unlock()

	if err = i2cbus.setAddress(addr); err != nil {
		return
	}

	n, err := i2cbus.file.Write([]byte{value})

	if n != 1 {
		err = fmt.Errorf("i2c: Unexpected number (%v) of bytes written in I2CBus.WriteByte", n)
	}

	return
}

func (i2cbus *I2CBus) WriteBytes(addr byte, value []byte) error {
	i2cbus.lock.Lock()
	defer i2cbus.lock.Unlock()

	if err := i2cbus.setAddress(addr); err != nil {
		return err
	}

	for i := range value {
		n, err := i2cbus.file.Write([]byte{value[i]})

		if n != 1 {
			return fmt.Errorf("i2c: Unexpected number (%v) of bytes written in I2CBus.WriteBytes", n)
		}
		if err != nil {
			return err
		}

		time.Sleep(Delay * time.Millisecond)
	}

	return nil
}
