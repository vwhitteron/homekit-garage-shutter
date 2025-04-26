package hardware

import (
	"log"
	"time"
)

func (s *Shutter) signalCloseShutter() {
	log.Println("Homekit GarageDoorOpener request: target=close")

	if s.rejectSignalUntil.After(time.Now()) {
		s.hcOpener.RejectStateChange(int(s.shutterState), "close", "debounce")

		return
	}

	s.rejectSignalUntil = time.Now().Add(5 * time.Second)

	log.Println("Shutter remote: signal=close")
	s.closeButton.Out(true)
	time.Sleep(500 * time.Millisecond)
	s.closeButton.Out(false)

	s.shutterState = shutterStateClosing

	s.hcOpener.SetStateClosed(5 * time.Second)
}

func (s *Shutter) signalOpenShutter() {
	log.Println("Homekit GarageDoorOpener request: target=open")

	if s.rejectSignalUntil.After(time.Now()) {
		s.hcOpener.RejectStateChange(int(s.shutterState), "open", "debounce")

		return
	} else if s.hcLock.IsLocked() {
		s.hcOpener.RejectStateChange(int(s.shutterState), "open", "locked")

		return
	}

	s.rejectSignalUntil = time.Now().Add(5 * time.Second)

	log.Println("Shutter remote: signal=open")
	s.openButton.Out(true)
	time.Sleep(500 * time.Millisecond)
	s.openButton.Out(false)

	s.shutterState = shutterStateOpening

	s.hcOpener.SetStateOpen(5 * time.Second)
}

func (s *Shutter) signalLockShutter() {
	log.Println("Homekit LockMechanism request: signal=lock")

	s.hcLock.Secure()

	s.hcLockSwitch.TurnOn()

	if s.options.CloseWhenLocked {
		log.Println("Shutter remote: source=lock signal=close")
		s.closeButton.Out(true)
		time.Sleep(500 * time.Millisecond)
		s.closeButton.Out(false)
	}
}

func (s *Shutter) signalUnlockShutter() {
	log.Println("Homekit LockMechanism request: signal=unlock")

	s.hcLock.SetStateUnsecured()

	s.hcLockSwitch.On.UpdateValue(false)
}
