package homekit

import (
	"log"

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

	acc.ContactSensor.ContactSensorState.SetValue(characteristic.ContactSensorStateContactDetected)

	acc.Accessory.AddService(acc.ContactSensor.Service)

	return &acc
}

func (s *GarageDoorOpenSensor) SetStateOpen(source string) {
	if s.IsOpen() {
		return
	}

	log.Printf("Homekit ContactSensor update: source=%s contact=open", source)

	s.ContactSensorState.SetValue(characteristic.ContactSensorStateContactNotDetected)
}

func (s *GarageDoorOpenSensor) SetStateClosed(source string) {
	if s.IsClosed() {
		return
	}

	log.Printf("Homekit ContactSensor update: source=%s contact=closed", source)

	s.ContactSensorState.SetValue(characteristic.ContactSensorStateContactDetected)
}

func (s *GarageDoorOpenSensor) IsOpen() bool {
	return s.ContactSensorState.GetValue() == characteristic.ContactSensorStateContactNotDetected
}

func (s *GarageDoorOpenSensor) IsClosed() bool {
	return s.ContactSensorState.GetValue() == characteristic.ContactSensorStateContactDetected
}
