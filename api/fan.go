package api

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
)

//

const FAN_PIN = "GPIO5" // PIN #29
// const TRIGGER_TEMP 	=	62.0		// Trigger the FAN once the CPU temperature goes above this (Celsius)

//

// This function is constantly (every 5 seconds) checking the CPU temperature
// and if it goes beyond the threshold defined in the dahsboard, it triggers the fan
func RunFanManager() error {

	fanPin := gpioreg.ByName(FAN_PIN)
	if fanPin == nil {
		return fmt.Errorf("failed to find fan_pin %v", PWR_BTN)
	}
	if err := fanPin.Out(gpio.Low); err != nil {
		return fmt.Errorf("can not access fan_pin: %v", err)
	}

	go func() {
		fanIsOn := false

		for {
			tempStr, _ := execOnHostWithLogs("vcgencmd measure_temp | egrep -o '[0-9]*\\.[0-9]*'", false)

			temp, err := strconv.ParseFloat(tempStr, 64)
			if err != nil {
				log.Printf("[ERR  ]: %s ", err.Error())
			}

			// if DEBUG_MODE {
			// 	// log.Printf( "[     ] CPU Temperature: [ %f ]", temp)
			// }

			if !fanIsOn && temp > Config.FanTriggerTemp {
				log.Printf("[     ] CPU Temperature: [ %f ]", temp)
				if err := fanPin.Out(gpio.High); err != nil {

					log.Printf("[ERR  ]: %s ", err.Error())

				} else {

					fanIsOn = true
				}

			}

			if fanIsOn && temp <= Config.FanTriggerTemp-3 {
				log.Printf("[     ] CPU Temperature: [ %f ]", temp)
				if err := fanPin.Out(gpio.Low); err != nil {

					log.Printf("[ERR  ]: %s ", err.Error())

				} else {

					fanIsOn = false
				}
			}

			time.Sleep(5 * time.Second)
		}

	}()

	log.Printf("[     ] Fan manager initialized.")
	return nil
}

//
