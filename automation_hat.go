package main

import (
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/sn3218"
	"periph.io/x/host/v3/rpi"
)

var (
	LedADC1     = 0
	LedADC2     = 1
	LedADC3     = 2
	LedOutput1  = 3
	LedOutput2  = 4
	LedOutput3  = 5
	LedRelay1NO = 6
	LedRelay1NC = 7
	LedRelay2NO = 8
	LedRelay2NC = 9
	LedRelay3NO = 10
	LedRelay3NC = 11
	LedInput3   = 12
	LedInput2   = 13
	LedInput1   = 14
	LedWarn     = 15 // red
	LedComm     = 16 // blue
	LedPower    = 17 // green
)

type AutomationHatOpts struct {
	AutoLeds bool
}

var AutomationHatDefaultOpts = AutomationHatOpts{
	AutoLeds: true,
}

// AutomationHatDev represents an Automation HAT
type AutomationHatDev struct {
	opts AutomationHatOpts

	output1 gpio.PinOut
	output2 gpio.PinOut
	output3 gpio.PinOut

	relay1 gpio.PinOut
	relay2 gpio.PinOut
	relay3 gpio.PinOut

	input1 gpio.PinIn
	input2 gpio.PinIn
	input3 gpio.PinIn

	leds *sn3218.Dev
}

// NewAutomationHat returns a automationhat driver.
func NewAutomationHat(opts *AutomationHatOpts) (*AutomationHatDev, error) {
	i2cPort, err := i2creg.Open("/dev/i2c-1")
	if err != nil {
		return nil, err
	}

	leds, err := sn3218.New(i2cPort)
	if err != nil {
		// automationhat mini doesn't have leds
		leds = nil
	}

	dev := &AutomationHatDev{
		opts: *opts,

		output1: rpi.P1_29, // GPIO 5
		output2: rpi.P1_32, // GPIO 12
		output3: rpi.P1_31, // GPIO 6

		relay1: rpi.P1_33, // GPIO 13
		relay2: rpi.P1_35, // GPIO 19
		relay3: rpi.P1_36, // GPIO 16

		input1: rpi.P1_37, // GPIO 26
		input2: rpi.P1_38, // GPIO 20
		input3: rpi.P1_40, // GPIO 21

		leds: leds,
	}

	if dev.leds != nil && dev.opts.AutoLeds {
		if err := dev.leds.WakeUp(); err != nil {
			return nil, err
		}

		if err := dev.leds.SwitchAll(true); err != nil {
			return nil, err
		}

		if err := dev.leds.Brightness(LedPower, 0x01); err != nil {
			return nil, err
		}
	}

	return dev, nil
}

// GetOutput1 returns gpio.PinOut corresponding to output 1
func (d *AutomationHatDev) GetOutput1() gpio.PinOut {
	return d.output1
}

// GetOutput2 returns gpio.PinOut corresponding to output 2
func (d *AutomationHatDev) GetOutput2() gpio.PinOut {
	return d.output2
}

// GetOutput3 returns gpio.PinOut corresponding to output 3
func (d *AutomationHatDev) GetOutput3() gpio.PinOut {
	return d.output3
}

// GetRelay1 returns gpio.PinOut corresponding to relay 1
func (d *AutomationHatDev) GetRelay1() gpio.PinOut {
	return d.relay1
}

// GetRelay2 returns gpio.PinOut corresponding to relay 2
func (d *AutomationHatDev) GetRelay2() gpio.PinOut {
	return d.relay2
}

// GetRelay3 returns gpio.PinOut corresponding to relay 3
func (d *AutomationHatDev) GetRelay3() gpio.PinOut {
	return d.relay3
}

// GetInput1 returns gpio.PinIn corresponding to input 1
func (d *AutomationHatDev) GetInput1() gpio.PinIn {
	return d.input1
}

// GetInput2 returns gpio.PinIn corresponding to input 2
func (d *AutomationHatDev) GetInput2() gpio.PinIn {
	return d.input2
}

// GetInput3 returns gpio.PinIn corresponding to input 3
func (d *AutomationHatDev) GetInput3() gpio.PinIn {
	return d.input3
}

func (d *AutomationHatDev) GetLeds() *sn3218.Dev {
	return d.leds
}

// Halt all internal devices.
func (d *AutomationHatDev) Halt() error {
	if err := d.output1.Halt(); err != nil {
		return err
	}

	if err := d.output2.Halt(); err != nil {
		return err
	}

	if err := d.output3.Halt(); err != nil {
		return err
	}

	if err := d.relay1.Halt(); err != nil {
		return err
	}

	if err := d.relay2.Halt(); err != nil {
		return err
	}

	if err := d.relay3.Halt(); err != nil {
		return err
	}

	if err := d.input1.Halt(); err != nil {
		return err
	}

	if err := d.input2.Halt(); err != nil {
		return err
	}

	if err := d.input3.Halt(); err != nil {
		return err
	}

	if d.leds != nil {
		if err := d.leds.Halt(); err != nil {
			return err
		}
	}

	return nil
}
