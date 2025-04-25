package homekit

import (
	"log"

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

func (l *GarageDoorLockSwitch) TurnOn() {
	log.Println("Homekit Switch update: value=on")

	l.On.UpdateValue(true)
}

func (l *GarageDoorLockSwitch) SetStateOff() {
	log.Println("Homekit Switch update: value=off")

	l.On.UpdateValue(false)
}
