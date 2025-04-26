package hardware

import (
	"log"
	"sync"
	"time"
	"vwhitteron/homekit-garage-shutter/homekit"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/host/v3"
)

type shutterState int

const (
	shutterStateFault   shutterState = -100
	shutterStateUnset   shutterState = -1
	shutterStateStopped shutterState = 0
	shutterStateClosed  shutterState = 1
	shutterStateClosing shutterState = 2
	shutterStateOpening shutterState = 3
	shutterStateMoving  shutterState = 4
	shutterStateOpen    shutterState = 5
)

type Shutter struct {
	hat          *AutomationHat
	openButton   gpio.PinOut
	closeButton  gpio.PinOut
	openContact  gpio.PinIn
	closeContact gpio.PinIn

	hcLock       *homekit.GarageDoorLock
	hcLockSwitch *homekit.GarageDoorLockSwitch
	hcOpener     *homekit.GarageDoorOpener
	hcOpenSensor *homekit.GarageDoorOpenSensor

	accessories []*accessory.Accessory

	options      ShutterOptions
	shutterState shutterState

	mu                sync.Mutex
	rejectSignalUntil time.Time
}

type ShutterOptions struct {
	BaseDirectory string

	SwitchHoldMs               uint
	EnableHomekitLockSwitch    bool
	EnableHomekitLockMechanism bool
	EnableHomekitContactSensor bool
	LockWhenClosed             bool
	CloseWhenLocked            bool

	Name           string
	Manufacturer   string
	Model          string
	SerialNumber   string
	HomekitPinCode string

	CloseButtonRelay  uint
	OpenButtonRelay   uint
	CloseContactInput uint
	OpenContactInput  uint
}

func NewShutter(opts ShutterOptions) *Shutter {
	if _, err := host.Init(); err != nil {
		log.Fatalf("failed to initialize periph: %v", err)
	}

	hat, err := NewAutomationHat(&AutomationHatDefaultOpts)
	if err != nil {
		log.Fatalf("failed to initialize AutomationHAT: %v", err)
	}

	openButton, err := hat.GetRelay(opts.OpenButtonRelay)
	if err != nil {
		log.Fatalf("failed to setup open button (relay 1): %v", err)
	}
	closeButton, err := hat.GetRelay(opts.CloseButtonRelay)
	if err != nil {
		log.Fatalf("failed to setup close button (relay 3): %v", err)
	}

	openContact, err := hat.GetInput(opts.OpenContactInput)
	if err != nil {
		log.Fatalf("failed to setup open sensor (input 1): %v", err)
	}
	closeContact, err := hat.GetInput(opts.CloseContactInput)
	if err != nil {
		log.Fatalf("failed to setup close sensor (input 2): %v", err)
	}

	info := accessory.Info{
		Name:         opts.Name,
		Manufacturer: opts.Manufacturer,
		Model:        opts.Model,
		SerialNumber: opts.SerialNumber,
	}

	accessories := []*accessory.Accessory{}

	hcOpener := homekit.NewGarageDoorOpener(info)
	accessories = append(accessories, hcOpener.Accessory)

	var hcOpenSensor *homekit.GarageDoorOpenSensor
	if opts.EnableHomekitContactSensor {
		hcOpenSensor = homekit.NewGarageDoorOpenSensor(info)

		accessories = append(accessories, hcOpenSensor.Accessory)
	}

	var hcLock *homekit.GarageDoorLock
	if opts.EnableHomekitLockMechanism {
		hcLock = homekit.NewGarageDoorLock(info)

		accessories = append(accessories, hcLock.Accessory)
	}

	var hcLockSwitch *homekit.GarageDoorLockSwitch
	if opts.EnableHomekitLockSwitch {
		hcLockSwitch = homekit.NewGarageDoorLockSwitch(info)

		accessories = append(accessories, hcLockSwitch.Accessory)
	}

	if opts.SwitchHoldMs == 0 {
		opts.SwitchHoldMs = 500
	}

	return &Shutter{
		options:      opts,
		shutterState: shutterStateUnset,

		hat:          hat,
		openButton:   openButton,
		closeButton:  closeButton,
		openContact:  openContact,
		closeContact: closeContact,

		hcLock:       hcLock,
		hcLockSwitch: hcLockSwitch,
		hcOpener:     hcOpener,
		hcOpenSensor: hcOpenSensor,

		accessories: accessories,
	}
}

func (s *Shutter) Run() {
	baseDir := s.options.BaseDirectory
	if baseDir == "" {
		baseDir = "."
	}

	config := hc.Config{
		Port:        "40111",
		Pin:         s.options.HomekitPinCode,
		StoragePath: baseDir + "/data",
	}

	transport, err := hc.NewIPTransport(config, s.accessories[0], s.accessories...)
	if err != nil {
		log.Panic(err)
	}

	hc.OnTermination(func() {
		<-transport.Stop()
	})

	log.Printf("Accessories configued: %d\n", len(s.accessories))

	log.Println("Setting up Homekit garage door opener handler")
	s.hcOpener.GarageDoorOpener.TargetDoorState.OnValueRemoteUpdate(func(state int) {
		s.mu.Lock()

		switch state {
		case characteristic.TargetDoorStateOpen:
			s.signalOpenShutter()
		case characteristic.TargetDoorStateClosed:
			s.signalCloseShutter()
		default:
			log.Printf("Homekit GarageDoorOpener request: signal=nil [unexpected state %d]\n", state)
		}

		s.mu.Unlock()
	})

	if s.options.EnableHomekitLockMechanism {
		log.Println("Setting up Homekit lock mechanism handler")
		s.hcLock.LockMechanism.LockTargetState.OnValueRemoteUpdate(func(state int) {
			switch state {
			case characteristic.LockTargetStateUnsecured:
				s.signalUnlockShutter()
			case characteristic.LockTargetStateSecured:
				s.signalLockShutter()
			default:
				log.Printf("Homekit LockMechanism request: signal=nil [unexpected state %d]\n", state)
			}
		})
	}

	if s.options.EnableHomekitLockSwitch {
		log.Println("Setting up Homekit lock switch handler")
		s.hcLockSwitch.On.OnValueRemoteUpdate(func(state bool) {
			switch state {
			case false:
				s.signalUnlockShutter()
			case true:
				s.signalLockShutter()
			default:
				log.Printf("Homekit Switch request: signal=nil [unexpected state %t]\n", state)
			}
		})
	}

	go s.pollPhysicalState()

	log.Println("Starting Homekit server: pin=" + config.Pin)
	transport.Start()

	err = s.hat.Halt()
	if err != nil {
		log.Fatalf("Failed to halt AutomationHAT: %v", err)
	}
}
