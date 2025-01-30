package main

import (
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
)

type GarageDoorLock struct {
	*accessory.Accessory
	*service.LockMechanism
}

func NewGarageDoorLock(info accessory.Info) *GarageDoorLock {
	acc := GarageDoorLock{}

	acc.Accessory = accessory.New(info, accessory.TypeDoorLock)
	acc.LockMechanism = service.NewLockMechanism()

	acc.LockCurrentState.SetValue(characteristic.LockCurrentStateSecured)

	acc.LockTargetState.SetValue(characteristic.LockTargetStateSecured)

	acc.Accessory.AddService(acc.LockMechanism.Service)

	return &acc
}
