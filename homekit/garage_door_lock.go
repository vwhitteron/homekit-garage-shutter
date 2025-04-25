package homekit

import (
	"log"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
)

type GarageDoorLock struct {
	*accessory.Accessory
	*service.LockMechanism
}

var lockState = map[int]string{
	0: "Unsecured",
	1: "Secured",
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

func (l *GarageDoorLock) Secure() {
	current := l.LockTargetState.GetValue()
	log.Printf("Homekit LockMechanism update: target=%s current=secured", lockState[current])

	l.LockTargetState.UpdateValue(characteristic.LockTargetStateSecured)
	l.LockCurrentState.UpdateValue(characteristic.LockCurrentStateSecured)
}

func (l *GarageDoorLock) SetStateUnsecured() {
	current := l.LockTargetState.GetValue()
	log.Printf("Homekit LockMechanism update: target=%s current=unsecured", lockState[current])

	l.LockTargetState.UpdateValue(characteristic.LockTargetStateUnsecured)
	l.LockCurrentState.UpdateValue(characteristic.LockCurrentStateUnsecured)
}

func (l *GarageDoorLock) IsLocked() bool {
	lockState := l.LockCurrentState.GetValue()

	if lockState == characteristic.LockCurrentStateUnsecured {
		log.Println("Is locked: current=unlocked")

		return false
	}

	log.Println("Is locked: current=locked")

	return true
}
