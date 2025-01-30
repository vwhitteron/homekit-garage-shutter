package main

import (
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/service"
)

type GarageDoorLockSwitch struct {
	*accessory.Accessory
	*service.Switch
}

func NewGarageDoorLockSwitch(info accessory.Info) *GarageDoorLockSwitch {
	acc := GarageDoorLockSwitch{}

	acc.Accessory = accessory.New(info, accessory.TypeSwitch)
	acc.Switch = service.NewSwitch()

	acc.Switch.On.SetValue(true)

	acc.Accessory.AddService(acc.Switch.Service)

	return &acc
}
