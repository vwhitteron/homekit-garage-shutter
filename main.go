package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"vwhitteron/homekit-garage-shutter/hardware"

	"github.com/spf13/viper"
)

var Version = "DEV"
var BuildTime string

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Println(sig)
		done <- true
	}()

	shutterOptions := &hardware.ShutterOptions{
		BaseDirectory:               "/opt/homekit-garage-shutter",
		EnableHomekitLockSwitch:     true,
		EnableHomekitLockMechanism:  true,
		EnableHomekitContactSensors: true,
		LockWhenClosed:              true,

		Name:         "Garage Shutter",
		Manufacturer: "generic",
		Model:        "",
		SerialNumber: "",

		HomekitPinCode: "00001234",

		OpenButtonRelay:   1,
		CloseButtonRelay:  3,
		OpenContactInput:  1,
		CloseContactInput: 2,
	}

	viper.SetEnvPrefix("HOMEBRIDGE_GARAGE_SHUTTER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))
	viper.AutomaticEnv()

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/boot/homekit-garage-shutter/")
	viper.AddConfigPath("/etc/homekit-garage-shutter/")
	viper.AddConfigPath("/opt/homekit-garage-shutter/")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("failed to read config file: ", err)
	} else {
		err = viper.Unmarshal(shutterOptions)
		if err != nil {
			log.Fatal("unmarshal config: ", err)
		}
	}

	if BuildTime == "" {
		BuildTime = time.Now().Format("2006-01-02_15:04:05")
	}
	fmt.Printf("homebridge-garage-shutter version %s (built %s)\n", Version, BuildTime)

	shutter := hardware.NewShutter(*shutterOptions)

	shutter.Run()

	<-done
}
