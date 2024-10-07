package main

import (
	"time"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
)

type GarageDoorOpener struct {
	*accessory.Accessory
	*service.GarageDoorOpener
	delayUpdate time.Time
}

func NewGarageDoorOpener(info accessory.Info) *GarageDoorOpener {
	acc := GarageDoorOpener{
		delayUpdate: time.Now(),
	}

	acc.Accessory = accessory.New(info, accessory.TypeGarageDoorOpener)
	acc.GarageDoorOpener = service.NewGarageDoorOpener()

	acc.ObstructionDetected.SetValue(false)

	acc.CurrentDoorState.SetMinValue(characteristic.CurrentDoorStateOpen)
	acc.CurrentDoorState.SetMaxValue(characteristic.CurrentDoorStateStopped)
	acc.CurrentDoorState.SetValue(characteristic.CurrentDoorStateClosed)

	acc.TargetDoorState.SetMinValue(characteristic.TargetDoorStateOpen)
	acc.TargetDoorState.SetMaxValue(characteristic.TargetDoorStateClosed)
	acc.TargetDoorState.SetValue(characteristic.TargetDoorStateClosed)

	acc.Accessory.AddService(acc.GarageDoorOpener.Service)

	return &acc
}
