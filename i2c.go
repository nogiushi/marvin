package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	I2C_SLAVE = 0x0703
	I2C_SMBUS = 0x0720

	I2C_SMBUS_WRITE = 0
	I2C_SMBUS_READ  = 1

	I2C_SMBUS_QUICK            = 0
	I2C_SMBUS_BYTE             = 1
	I2C_SMBUS_BYTE_DATA        = 2
	I2C_SMBUS_WORD_DATA        = 3
	I2C_SMBUS_PROC_CALL        = 4
	I2C_SMBUS_BLOCK_DATA       = 5
	I2C_SMBUS_I2C_BLOCK_BROKEN = 6
	I2C_SMBUS_BLOCK_PROC_CALL  = 7
	I2C_SMBUS_I2C_BLOCK_DATA   = 8
)

type i2c struct {
	file *os.File
}

func NewI2C(bus byte) (i *i2c, err error) {
	i = &i2c{}
	i.file, err = os.OpenFile(fmt.Sprintf("/dev/i2c-%v", bus), os.O_RDWR, os.ModeExclusive)
	return
}

func (i *i2c) SetAddress(address byte) (err error) {
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, i.file.Fd(), I2C_SLAVE, uintptr(address)); errno != 0 {
		err = syscall.Errno(errno)
	}
	return
}

func (i *i2c) WriteByte(command, data byte) (err error) {
	d := struct {
		readWrite byte
		command   byte
		size      uint32
		data      uintptr
	}{
		I2C_SMBUS_WRITE,
		command,
		I2C_SMBUS_BYTE_DATA,
		uintptr(unsafe.Pointer(&data)),
	}
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, i.file.Fd(), I2C_SMBUS, uintptr(unsafe.Pointer(&d))); errno != 0 {
		err = syscall.Errno(errno)
	}

	return
}

func (i *i2c) ReadByte(command byte) (data byte, err error) {
	d := struct {
		readWrite byte
		command   byte
		size      uint32
		data      uintptr
	}{
		I2C_SMBUS_READ,
		command,
		I2C_SMBUS_BYTE_DATA,
		uintptr(unsafe.Pointer(&data)),
	}
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, i.file.Fd(), I2C_SMBUS, uintptr(unsafe.Pointer(&d))); errno != 0 {
		err = syscall.Errno(errno)
	}

	return
}
