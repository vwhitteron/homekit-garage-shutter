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

	acc := NewGarageDoorOpener(info)

	// configure the ip transport
	config := hc.Config{
		Port:        "40111",
		Pin:         "00002197",
		StoragePath: "/opt/homekit-garage-shutter/data",
	}
	t, err := hc.NewIPTransport(config, acc.Accessory)
	if err != nil {
		log.Panic(err)
	}

	hc.OnTermination(func() {
		<-t.Stop()
	})

	hat.leds.WakeUp()
	hat.leds.Switch((LedComm), true)

	acc.GarageDoorOpener.TargetDoorState.OnValueRemoteUpdate(func(state int) {
		switch state {
		case characteristic.TargetDoorStateOpen:
			relay := hat.GetRelay1()

			log.Println("Door controller: signal=open")

			relay.Out(true)
			hat.leds.Switch(LedRelay1NC, true)
			time.Sleep(500 * time.Millisecond)
			relay.Out(false)
			hat.leds.Switch(LedRelay1NC, false)

			acc.TargetDoorState.UpdateValue(characteristic.TargetDoorStateOpen)
			acc.CurrentDoorState.UpdateValue(characteristic.CurrentDoorStateOpening)
			acc.delayUpdate = time.Now().Add(5 * time.Second)
		case characteristic.TargetDoorStateClosed:
			relay := hat.GetRelay3()

			log.Println("Door controller: signal=close")

			relay.Out(true)
			hat.leds.Switch(LedRelay3NC, true)
			time.Sleep(500 * time.Millisecond)
			relay.Out(false)
			hat.leds.Switch(LedRelay3NC, false)

			acc.TargetDoorState.UpdateValue(characteristic.TargetDoorStateClosed)
			acc.CurrentDoorState.UpdateValue(characteristic.CurrentDoorStateClosing)
			acc.delayUpdate = time.Now().Add(5 * time.Second)
		default:
			log.Printf("Door controller: signal=nil [unexpected state %d]\n", state)
		}
	})

	go pollDoorState(acc, hat)

	t.Start()

	<-done // Wait

	hat.leds.Switch((LedComm), false)
}

func pollDoorState(acc *GarageDoorOpener, hat *AutomationHatDev) {
	for {
		doorClosedInput := hat.GetInput1()
		doorOpenInput := hat.GetInput2()

		if acc.delayUpdate.After(time.Now()) {
			continue
		}

		if doorClosedInput != nil && doorOpenInput != nil {
			isClosed := doorClosedInput.Read()
			isOpen := doorOpenInput.Read()

			hat.leds.Switch(LedInput1, bool(isClosed))
			hat.leds.Switch(LedInput2, bool(isOpen))

			if !isOpen && isClosed {
				updateState(acc, hat, characteristic.CurrentDoorStateClosed)

			} else if !isClosed && isOpen {
				updateState(acc, hat, characteristic.CurrentDoorStateOpen)

			} else if !isClosed && !isOpen {
				currentDoorState := acc.CurrentDoorState.GetValue()
				targetDoorState := acc.TargetDoorState.GetValue()

				if currentDoorState != characteristic.CurrentDoorStateOpening && currentDoorState != characteristic.CurrentDoorStateClosing {

					if targetDoorState == characteristic.TargetDoorStateClosed {
						log.Println("Door state: target=open current=opening")
						acc.TargetDoorState.UpdateValue(characteristic.TargetDoorStateOpen)
						acc.CurrentDoorState.UpdateValue(characteristic.CurrentDoorStateOpening)
					} else {
						log.Println("Door state: target=closed current=closing")
						acc.TargetDoorState.UpdateValue(characteristic.TargetDoorStateClosed)
						acc.CurrentDoorState.UpdateValue(characteristic.CurrentDoorStateClosing)
					}
				}

			} else {
				log.Println("ERROR: Unexpected door state reported as both opened and closed")

			}
		}

		time.Sleep(time.Second)
	}
}

func updateState(acc *GarageDoorOpener, hat *AutomationHatDev, newState int) {
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
