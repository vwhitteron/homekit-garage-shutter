package hardware

import (
	"log"
	"time"

	"github.com/brutella/hc/characteristic"
	"periph.io/x/conn/v3/gpio"
)

func (s *Shutter) setShutterOpen() {
	if s.shutterState != shutterStateOpen {
		log.Println("Door Contact Sensor: source=hardware contact=open")
	}

	s.shutterState = shutterStateOpen

	s.hcOpenSensor.SetStateOpen("hardware")

	if !s.hcOpener.IsOpen() {
		s.updateState(characteristic.CurrentDoorStateOpen)
	}
}

func (s *Shutter) setShutterClosed() {
	if s.shutterState != shutterStateClosed {
		log.Println("Door Contact Sensor: source=hardware contact=closed")
	}

	s.shutterState = shutterStateClosed

	s.hcOpenSensor.SetStateClosed("hardware")

	if s.hcOpener.IsClosed() {
		return
	}

	s.updateState(characteristic.CurrentDoorStateClosed)

	if s.options.LockWhenClosed {
		s.hcLock.Secure()
		s.hcLockSwitch.TurnOn()
	}
}

func (s *Shutter) setShutterMoving() {
	if s.shutterState != shutterStateMoving {
		log.Println("Door Contact Sensor: source=hardware contact=moving")
	}

	s.shutterState = shutterStateMoving

	s.hcOpenSensor.SetStateOpen("hardware")

	if !s.hcOpener.IsMoving() {
		if s.hcOpener.TargetDoorState.GetValue() == characteristic.TargetDoorStateClosed {
			log.Println("Homekit GarageDoorOpener update: source=hardware target=open current=opening")

			s.hcOpener.SetStateOpen(0 * time.Second)
		} else {
			log.Println("Homekit GarageDoorOpener update: source=hardware target=closed current=closing")

			s.hcOpener.SetStateClosed(0 * time.Second)
		}
	}
}

func (s *Shutter) setShutterFault() {
	s.shutterState = shutterStateFault

	log.Println("Door Contact Sensor: source=hardware contact=FAULT")

	s.hcOpenSensor.SetStateClosed("hardware")
}

func (s *Shutter) pollPhysicalState() {
	for {
		time.Sleep(time.Second)

		if s.hcOpener.IsUpdateBlocked() {
			log.Println("Poll door state: delayed")
			continue
		}

		switch s.getShutterPosition() {
		case shutterStateOpen:
			s.setShutterOpen()
		case shutterStateClosed:
			s.setShutterClosed()
		case shutterStateMoving:
			s.setShutterMoving()
		default:
			s.setShutterFault()
		}
	}
}

func (s *Shutter) getShutterPosition() shutterState {
	position := shutterStateFault

	if s.openContact == nil || s.closeContact == nil {
		return position
	}

	closedContactState := s.closeContact.Read()
	openContactState := s.openContact.Read()

	btoi := map[gpio.Level]uint8{false: 0, true: 1}

	bitField := (btoi[closedContactState] << 1) | btoi[openContactState]

	switch bitField {
	case 0b00:
		position = shutterStateMoving
	case 0b01:
		position = shutterStateOpen
	case 0b10:
		position = shutterStateClosed
	default:
		position = shutterStateFault
	}

	return position
}
func (s *Shutter) updateState(newState int) {
	currentState := s.hcOpener.CurrentDoorState.GetValue()

	if currentState == newState {
		return
	}

	switch newState {
	case characteristic.CurrentDoorStateClosing:
		log.Println("Door state: source=hardware current=closing")
	case characteristic.CurrentDoorStateOpening:
		log.Println("Door state: source=hardware current=opening")
	case characteristic.CurrentDoorStateStopped:
		log.Println("Door state: source=hardware current=stopped")
	case characteristic.CurrentDoorStateOpen:
		log.Println("Door state: source=hardware current=open")

		if currentState != characteristic.CurrentDoorStateStopped {
			log.Println("Door state: source=hardware target=open")
			s.hcOpener.TargetDoorState.UpdateValue(newState)
		}
	case characteristic.CurrentDoorStateClosed:
		log.Println("Door state: source=hardware current=closed")

		if currentState != characteristic.CurrentDoorStateStopped {
			log.Println("Door state: source=hardware target=closed")
			s.hcOpener.TargetDoorState.UpdateValue(newState)
		}
	}

	s.hcOpener.CurrentDoorState.UpdateValue(newState)
}
