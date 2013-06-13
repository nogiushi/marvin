// +build !linux

package main

import "fmt"

func GPIOInterrupt(number int) (ch chan bool, err error) {
	return nil, fmt.Errorf("Not supported on current architecture")
}
