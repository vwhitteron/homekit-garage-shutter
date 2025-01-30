package main

import (
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
)

type GarageDoorOpenSensor struct {
	*accessory.Accessory
	*service.ContactSensor
}

func NewGarageDoorOpenSensor(info accessory.Info) *GarageDoorOpenSensor {
	acc := GarageDoorOpenSensor{}

	acc.Accessory = accessory.New(info, accessory.TypeSensor)
	acc.ContactSensor = service.NewContactSensor()

	acc.ContactSensor.ContactSensorState.SetValue(characteristic.ContactSensorStateContactNotDetected)

	acc.Accessory.AddService(acc.ContactSensor.Service)

	return &acc
}
