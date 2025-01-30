package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"periph.io/x/host/v3"
)

const LockOnClose = true

func main() {
	// signal handler
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigs // Wait for signal
		log.Println(sig)
		done <- true
	}()

	// init periph
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	opts := &AutomationHatDefaultOpts

	hat, err := NewAutomationHat(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer hat.Halt()

	// init homekit accessory
	info := accessory.Info{
		Name:         "Garage Shutter",
		Manufacturer: "文化シャッター",
		Model:        "CY-2201(A)",
		SerialNumber: "0002197",
	}

	lock := NewGarageDoorLock(info)
	lockSwitch := NewGarageDoorLockSwitch(info)
	shutter := NewGarageDoorOpener(info)
	openSensor := NewGarageDoorOpenSensor(info)

	// configure the ip transport
	config := hc.Config{
		Port:        "40111",
		Pin:         "00002197",
		StoragePath: "/opt/homekit-garage-shutter/data",
	}
	t, err := hc.NewIPTransport(config, shutter.Accessory, lock.Accessory, lockSwitch.Accessory, openSensor.Accessory)
	if err != nil {
		log.Panic(err)
	}

	hc.OnTermination(func() {
		<-t.Stop()
	})

	lock.LockMechanism.LockTargetState.OnValueRemoteUpdate(func(state int) {
		switch state {
		case characteristic.LockTargetStateUnsecured:
			log.Println("Homekit LockMechanism request: signal=unlock")

			log.Println("Homekit LockMechanism update: target=unsecured current=unsecured")
			lock.LockTargetState.UpdateValue(characteristic.LockTargetStateUnsecured)
			lock.LockCurrentState.UpdateValue(characteristic.LockCurrentStateUnsecured)

			log.Println("Homekit Switch update: value=off")
			lockSwitch.On.UpdateValue(false)
		case characteristic.LockTargetStateSecured:
			log.Println("Homekit LockMechanism request: signal=lock")

			log.Println("Homekit LockMechanism update: target=secured current=secured")
			lock.LockTargetState.UpdateValue(characteristic.LockTargetStateSecured)
			lock.LockCurrentState.UpdateValue(characteristic.LockCurrentStateSecured)

			log.Println("Homekit Switch update: value=on")
			lockSwitch.On.UpdateValue(true)

			log.Println("Shutter remote: signal=close")
			relay := hat.GetRelay3()
			relay.Out(true)
			time.Sleep(500 * time.Millisecond)
			relay.Out(false)
		}
	})

	lockSwitch.On.OnValueRemoteUpdate(func(state bool) {
		switch state {
		case false:
			log.Println("Homekit LockMechanism request: signal=unlock")

			log.Println("Homekit LockMechanism update: target=unsecured current=unsecured")
			lock.LockTargetState.UpdateValue(characteristic.LockTargetStateUnsecured)
			lock.LockCurrentState.UpdateValue(characteristic.LockCurrentStateUnsecured)

			log.Println("Homekit Switch update: value=off")
			lockSwitch.On.UpdateValue(false)
		case true:
			log.Println("Homekit LockMechanism request: signal=lock")

			log.Println("Homekit LockMechanism update: target=secured current=secured")
			lock.LockTargetState.UpdateValue(characteristic.LockTargetStateSecured)
			lock.LockCurrentState.UpdateValue(characteristic.LockCurrentStateSecured)

			log.Println("Homekit Switch update: value=on")
			lockSwitch.On.UpdateValue(true)
		}
	})

	shutter.GarageDoorOpener.TargetDoorState.OnValueRemoteUpdate(func(state int) {
		switch state {
		case characteristic.TargetDoorStateOpen:
			log.Println("Homekit GarageDoorOpener request: target=open")

			if isUnlocked(lock) {
				log.Println("Homekit GarageDoorOpener update: target=open current=opening")

				log.Println("Shutter remote: signal=open")
				relay := hat.GetRelay1()
				relay.Out(true)
				time.Sleep(500 * time.Millisecond)
				relay.Out(false)

				shutter.TargetDoorState.UpdateValue(characteristic.TargetDoorStateOpen)
				shutter.CurrentDoorState.UpdateValue(characteristic.CurrentDoorStateOpening)
				shutter.delayUpdate = time.Now().Add(5 * time.Second)
			} else {
				log.Println("Homekit GarageDoorOpener request: status=rejected reason=locked")

				shutter.delayUpdate = time.Now().Add(2 * time.Second)
				time.Sleep(1 * time.Second)

				log.Println("Homekit GarageDoorOpener update: target=closed")
				shutter.TargetDoorState.UpdateValue(characteristic.TargetDoorStateClosed)
			}
		case characteristic.TargetDoorStateClosed:
			log.Println("Homekit GarageDoorOpener request: target=close")

			log.Println("Shutter remote: signal=close")
			relay := hat.GetRelay3()
			relay.Out(true)
			time.Sleep(500 * time.Millisecond)
			relay.Out(false)

			log.Println("Homekit GarageDoorOpener update: target=closed current=closing")
			shutter.TargetDoorState.UpdateValue(characteristic.TargetDoorStateClosed)
			shutter.CurrentDoorState.UpdateValue(characteristic.CurrentDoorStateClosing)
			shutter.delayUpdate = time.Now().Add(5 * time.Second)
		default:
			log.Printf("Homekit GarageDoorOpener request: signal=nil [unexpected state %d]\n", state)
		}
	})

	go pollDoorState(shutter, lock, lockSwitch, openSensor, hat)

	t.Start()

	<-done // Wait
}

func pollDoorState(acc *GarageDoorOpener,
	lock *GarageDoorLock,
	lockSwitch *GarageDoorLockSwitch,
	sensor *GarageDoorOpenSensor,
	hat *AutomationHatDev) {

	for {
		doorClosedInput := hat.GetInput1()
		doorOpenInput := hat.GetInput2()

		updateDoorState := true
		if acc.delayUpdate.After(time.Now()) {
			log.Printf("Poll door state: delay=%d ms", time.Until(acc.delayUpdate).Milliseconds())
			// time.Sleep(time.Until(acc.delayUpdate))
			updateDoorState = false
		}

		if doorClosedInput != nil && doorOpenInput != nil {
			isClosed := doorClosedInput.Read()
			isOpen := doorOpenInput.Read()

			if !isOpen && isClosed { // shutter closed
				if sensor.ContactSensorState.GetValue() != characteristic.ContactSensorStateContactDetected {
					log.Println("Homekit ContactSensor update: source=hardware contact=closed")
					sensor.ContactSensorState.SetValue(characteristic.ContactSensorStateContactDetected)
				}

				if updateDoorState {
					if acc.CurrentDoorState.GetValue() != characteristic.CurrentDoorStateClosed {
						updateState(acc, characteristic.CurrentDoorStateClosed)

						if LockOnClose {
							lock.LockTargetState.UpdateValue(characteristic.LockTargetStateSecured)
							lock.LockCurrentState.UpdateValue(characteristic.LockTargetStateSecured)
							lockSwitch.On.UpdateValue(true)
						}
					}
				}
			} else if !isClosed && isOpen { // shutter open
				if sensor.ContactSensorState.GetValue() != characteristic.ContactSensorStateContactNotDetected {
					log.Println("Homekit ContactSensor update: source=hardware contact=open")
					sensor.ContactSensorState.SetValue(characteristic.ContactSensorStateContactNotDetected)
				}

				if updateDoorState {
					if acc.CurrentDoorState.GetValue() != characteristic.CurrentDoorStateOpen {
						updateState(acc, characteristic.CurrentDoorStateOpen)
					}
				}
			} else if !isClosed && !isOpen { // shutter moving
				if sensor.ContactSensorState.GetValue() != characteristic.ContactSensorStateContactNotDetected {
					log.Println("Homekit ContactSensor update: source=harware contact=open")
					sensor.ContactSensorState.SetValue(characteristic.ContactSensorStateContactNotDetected)
				}

				if updateDoorState {
					currentDoorState := acc.CurrentDoorState.GetValue()
					targetDoorState := acc.TargetDoorState.GetValue()

					if currentDoorState != characteristic.CurrentDoorStateOpening && currentDoorState != characteristic.CurrentDoorStateClosing {
						if targetDoorState == characteristic.TargetDoorStateClosed {
							log.Println("Homekit GarageDoorOpener update: source=hardware target=open current=opening")
							acc.TargetDoorState.UpdateValue(characteristic.TargetDoorStateOpen)
							acc.CurrentDoorState.UpdateValue(characteristic.CurrentDoorStateOpening)
						} else {
							log.Println("Homekit GarageDoorOpener update: source=hardware target=closed current=closing")
							acc.TargetDoorState.UpdateValue(characteristic.TargetDoorStateClosed)
							acc.CurrentDoorState.UpdateValue(characteristic.CurrentDoorStateClosing)
						}
					}
				}
			} else { // shutter in invalid state
				log.Println("ERROR: Unexpected door state reported as both opened and closed")

				if sensor.ContactSensorState.GetValue() != characteristic.ContactSensorStateContactDetected {
					log.Println("Homekit ContactSensor update: source=hardware contact=closed")
					sensor.ContactSensorState.SetValue(characteristic.ContactSensorStateContactDetected)
				}
			}
		}

		time.Sleep(time.Second)
	}
}

func updateState(acc *GarageDoorOpener, newState int) {
	oldState := acc.CurrentDoorState.GetValue()

	if oldState == newState {
		return
	}

	log.Printf("Door state: old=%d new=%d\n", oldState, newState)

	switch newState {
	case characteristic.CurrentDoorStateClosing:
		log.Println("Door state: current=closing")
	case characteristic.CurrentDoorStateOpening:
		log.Println("Door state: opening")
	case characteristic.CurrentDoorStateStopped:
		log.Println("Door state: current=stopped")
	case characteristic.CurrentDoorStateOpen:
		log.Println("Door state: current=open")

		if oldState != characteristic.CurrentDoorStateStopped {
			log.Println("Door state: target=open")
			acc.TargetDoorState.UpdateValue(newState)
		}
	case characteristic.CurrentDoorStateClosed:
		log.Println("Door state: current=closed")

		if oldState != characteristic.CurrentDoorStateStopped {
			log.Println("Door state: target=closed")
			acc.TargetDoorState.UpdateValue(newState)
		}
	}

	acc.CurrentDoorState.UpdateValue(newState)
}

func isUnlocked(acc *GarageDoorLock) bool {
	lockState := acc.LockCurrentState.GetValue()

	if lockState == characteristic.LockCurrentStateUnsecured {
		log.Println("Is locked: current=unlocked")

		return true
	}

	log.Println("Is locked: current=locked")

	return false
}
