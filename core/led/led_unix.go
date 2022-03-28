//+build !windows

package led

import "github.com/stianeikeland/go-rpio/v4"

func EnableLed(port int) error {
	pin := rpio.Pin(port)

	err := rpio.Open()
	if err != nil {
		return err
	}

	// Switch on
	pin.Output()
	pin.High()

	return nil
}

func DisableLed(port int) error {
	pin := rpio.Pin(port)

	// Switch off
	pin.Output()
	pin.Low()

	return rpio.Close()
}
