//+build !windows

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

	// Switch on
	pin.Output()
	pin.High()

	return nil
}

func DisablePin() error {
	pin := rpio.Pin(pin)

	// Switch off
	pin.Output()
	pin.Low()

	return rpio.Close()
}
