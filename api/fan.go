package api

import (
	"log"
	"strconv"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
)

/*-------------------------*/

const FAN_PIN = "GPIO5"   // PIN #29
const TRIGGER_TEMP = 62.0 // Trigger the FAN once the CPU temperature goes above this (Celsius)

/*-------------------------*/

func FanLoop() {

	go func() {
		if _, err := host.Init(); err != nil {
			log.Printf("[Err   ]: %s ", err.Error())
		}

		pin := gpioreg.ByName(FAN_PIN) // FAN pin
		pin.Out(gpio.Low)
		fanIsOn := false

		for {
			tempStr, _ := execOnHostWithLogs("vcgencmd measure_temp | egrep -o '[0-9]*\\.[0-9]*'", false, nil)

			temp, err := strconv.ParseFloat(tempStr, 64)
			if err != nil {
				log.Printf("[Err   ]: %s ", err.Error())
			}

			if DEBUG_MODE {
				// log.Printf( "[Info  ] CPU Temperature: [ %f ]", temp)
			}

			if !fanIsOn && temp > TRIGGER_TEMP {
				log.Printf("[Info  ] CPU Temperature: [ %f ]", temp)
				if err := pin.Out(gpio.High); err != nil {

					log.Printf("[Err   ]: %s ", err.Error())

				} else {

					fanIsOn = true
				}

			}

			if fanIsOn && temp <= TRIGGER_TEMP-3 {
				log.Printf("[Info  ] CPU Temperature: [ %f ]", temp)
				if err := pin.Out(gpio.Low); err != nil {

					log.Printf("[Err   ]: %s ", err.Error())

				} else {

					fanIsOn = false
				}
			}

			time.Sleep(5 * time.Second)
		}

	}()

	log.Printf("[Info  ] Fan manager initialized.")

}

/*-------------------------*/
