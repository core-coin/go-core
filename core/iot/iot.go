// Package iot +build arm64
package iot

import "github.com/stianeikeland/go-rpio/v4"

const (
	pin = 1
)

func EnablePin() error {
	err := rpio.Open()
	if err != nil {
		return err
	}
	pin := rpio.Pin(pin)
	pin.Input() // Input mode

	// Read status
	res := pin.Read()

	// Switch on
	pin.Output()
	if res == rpio.High {
		pin.Low()
	} else {
		pin.High()
	}

	return nil
}

func DisablePin() error {
	pin := rpio.Pin(pin)
	pin.Input() // Input mode

	// Read status
	res := pin.Read()

	// Switch on
	pin.Output()
	if res == rpio.High {
		pin.Low()
	} else {
		pin.High()
	}

	return rpio.Close()
}
