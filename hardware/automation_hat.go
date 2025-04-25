package hardware

import (
	"fmt"

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

// AutomationHat represents an Automation HAT
type AutomationHat struct {
	opts AutomationHatOpts

	outputs []gpio.PinOut
	relays  []gpio.PinOut
	inputs  []gpio.PinIn
	leds    *sn3218.Dev
}

// NewAutomationHat returns a automationhat driver.
func NewAutomationHat(opts *AutomationHatOpts) (*AutomationHat, error) {
	i2cPort, err := i2creg.Open("/dev/i2c-1")
	if err != nil {
		return nil, err
	}

	leds, err := sn3218.New(i2cPort)
	if err != nil {
		// automationhat mini doesn't have leds
		leds = nil
	}

	dev := &AutomationHat{
		opts: *opts,

		outputs: []gpio.PinOut{
			rpi.P1_29, // GPIO 5
			rpi.P1_32, // GPIO 12
			rpi.P1_31, // GPIO 6
		},

		relays: []gpio.PinOut{
			rpi.P1_33, // GPIO 13
			rpi.P1_35, // GPIO 19
			rpi.P1_36, // GPIO 16
		},

		inputs: []gpio.PinIn{
			rpi.P1_37, // GPIO 26
			rpi.P1_38, // GPIO 20
			rpi.P1_40, // GPIO 21
		},

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

func (d *AutomationHat) GetOutput(output uint) (gpio.PinOut, error) {
	if output > uint(len(d.outputs))+1 {
		return nil, fmt.Errorf("invalid output %d", output)
	}

	return d.outputs[output-1], nil
}

func (d *AutomationHat) GetRelay(relay uint) (gpio.PinOut, error) {
	if relay > uint(len(d.relays))+1 {
		return nil, fmt.Errorf("invalid relay %d", relay)
	}

	return d.relays[relay-1], nil
}

func (d *AutomationHat) GetInput(input uint) (gpio.PinIn, error) {
	if input > uint(len(d.inputs))+1 {
		return nil, fmt.Errorf("invalid relay %d", input)
	}

	return d.inputs[input-1], nil
}

// Halt all internal devices.
func (d *AutomationHat) Halt() error {
	for _, output := range d.outputs {
		if err := output.Halt(); err != nil {
			return err
		}
	}

	for _, relay := range d.relays {
		if err := relay.Halt(); err != nil {
			return err
		}
	}

	for _, input := range d.inputs {
		if err := input.Halt(); err != nil {
			return err
		}
	}

	if d.leds != nil {
		if err := d.leds.Halt(); err != nil {
			return err
		}
	}

	return nil
}
