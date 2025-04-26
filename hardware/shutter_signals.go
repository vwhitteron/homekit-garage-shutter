package hardware

import (
	"log"
	"time"

	"periph.io/x/conn/v3/gpio"
)

func (s *Shutter) signalCloseShutter() {
	log.Println("Homekit GarageDoorOpener request: target=close")

	if s.rejectSignalUntil.After(time.Now()) {
		s.hcOpener.RejectStateChange(int(s.shutterState), "close", "debounce")

		return
	}

	s.rejectSignalUntil = time.Now().Add(5 * time.Second)

	log.Println("Shutter remote: signal=close")
	s.pressButton(s.closeButton)

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
	s.pressButton(s.openButton)

	s.shutterState = shutterStateOpening

	s.hcOpener.SetStateOpen(5 * time.Second)
}

func (s *Shutter) signalLockShutter() {
	log.Println("Homekit LockMechanism request: signal=lock")

	s.hcLock.Secure()

	s.hcLockSwitch.TurnOn()

	if s.options.CloseWhenLocked {
		log.Println("Shutter remote: source=lock signal=close")
		s.pressButton(s.closeButton)
	}
}

func (s *Shutter) signalUnlockShutter() {
	log.Println("Homekit LockMechanism request: signal=unlock")

	s.hcLock.SetStateUnsecured()

	s.hcLockSwitch.On.UpdateValue(false)
}

func (s *Shutter) pressButton(button gpio.PinOut) {
	duration := time.Duration(s.options.SwitchHoldMs) * time.Millisecond

	err := button.Out(true)
	if err != nil {
		log.Printf("Error engaging relay %q: %v", button.Name(), err)

		return
	}

	time.Sleep(duration)

	err = button.Out(false)
	if err != nil {
		log.Printf("Error releasing relay %q: %v", button.Name(), err)

		return
	}
}
