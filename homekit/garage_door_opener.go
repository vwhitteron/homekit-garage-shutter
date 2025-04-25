package homekit

import (
	"log"
	"time"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
)

var shutterStateName = map[int]string{
	-100: "fault",
	-1:   "unset",
	0:    "stopped",
	1:    "closed",
	2:    "closing",
	3:    "opening",
	4:    "moving",
	5:    "open",
}

type GarageDoorOpener struct {
	*accessory.Accessory
	*service.GarageDoorOpener
	updatesBlockedUntil time.Time
}

func NewGarageDoorOpener(info accessory.Info) *GarageDoorOpener {
	acc := GarageDoorOpener{
		updatesBlockedUntil: time.Now(),
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

func (o *GarageDoorOpener) BlockUpdateUntil(t time.Time) {
	if t.Before(o.updatesBlockedUntil) {
		return
	}

	o.updatesBlockedUntil = t
}

func (o *GarageDoorOpener) IsUpdateBlocked() bool {
	return o.updatesBlockedUntil.After(time.Now())
}

func (o *GarageDoorOpener) SetStateClosed(seconds time.Duration) {
	log.Println("Homekit GarageDoorOpener update: target=closed current=closing")

	o.TargetDoorState.UpdateValue(characteristic.TargetDoorStateClosed)
	o.CurrentDoorState.UpdateValue(characteristic.CurrentDoorStateClosing)

	o.BlockUpdateUntil(time.Now().Add(seconds))
}

func (o *GarageDoorOpener) SetStateOpen(seconds time.Duration) {
	log.Println("Homekit GarageDoorOpener update: target=open current=opening")

	o.TargetDoorState.UpdateValue(characteristic.TargetDoorStateOpen)
	o.CurrentDoorState.UpdateValue(characteristic.CurrentDoorStateOpening)

	o.BlockUpdateUntil(time.Now().Add(seconds))
}

func (o *GarageDoorOpener) RejectStateChange(initial int, target string, reason string) {
	log.Printf("Homekit GarageDoorOpener request: target=%s current=%s  status=rejected reason=%s\n", target, shutterStateName[initial], reason)

	time.Sleep(1 * time.Second)

	o.TargetDoorState.SetValue(initial)
	o.CurrentDoorState.SetValue(initial)
}

func (o *GarageDoorOpener) IsOpen() bool {
	return o.CurrentDoorState.GetValue() == characteristic.CurrentDoorStateOpen
}

func (o *GarageDoorOpener) IsClosed() bool {
	return o.CurrentDoorState.GetValue() == characteristic.CurrentDoorStateClosed
}

func (o *GarageDoorOpener) IsMoving() bool {
	currentDoorState := o.CurrentDoorState.GetValue()

	return currentDoorState == characteristic.CurrentDoorStateOpening ||
		currentDoorState == characteristic.CurrentDoorStateClosing
}
