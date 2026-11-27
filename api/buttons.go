package api

import (
	"fmt"
	"log"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
)

//

const WiFi_BTN = "GPIO6"     // PIN #31
const WiFi_BTN_COUNTDOWN = 3 // for n seconds the button needs to be held down to revert the wifi/web ui settings

const PWR_BTN = "GPIO26"     // PIN #37
const SHUTDOWN_COUNTDOWN = 3 // for n seconds the button needs to be held down to activate shutdown procedure

//

// This function handles the push buttons on WaziHAT
func RunButtonsManager() error {

	//
	//WiFi button

	btnPin := gpioreg.ByName(WiFi_BTN)
	if btnPin == nil {
		return fmt.Errorf("failed to find wifi_btn %v", WiFi_BTN)
	}

	// Set it as input with a pull down (it defaults to Low when unconnected).
	// NoEdge, not RisingEdge: see gpiopoll.go -- periph.io's edge detection
	// goes through the legacy sysfs GPIO interface, which the trixie kernel
	// renumbered, so requesting an edge fails. watchEdge polls instead.
	if err := btnPin.In(gpio.PullDown, gpio.NoEdge); err != nil {
		return fmt.Errorf("can not access wifi_btn: %v", err)
	}

	go watchEdge(btnPin, gpio.High, func() bool {

		if oledHalted {

			//Since power button and OLED shared a pin, we need to wait for the oled to be re-initialized
			go func() {
				time.Sleep(1 * time.Second)
				oledShow("\n\n   Screen ON", false)
			}()

		}

		if DEBUG_MODE {
			log.Printf("[     ] Button %s pushed", btnPin)
		}

		holdCounter := 1
		for btnPin.Read() == gpio.High {
			time.Sleep(1 * time.Second)
			holdCounter++
			if holdCounter > WiFi_BTN_COUNTDOWN {
				if DEBUG_MODE {
					log.Printf("[     ] Button %s held long enough. Triggering the action!", btnPin)
				}
				wifiOperation.Lock()
				ActivateAPMode()
				wifiOperation.Unlock()
			}
		}
		return true
	})

	//

	btnPwr := gpioreg.ByName(PWR_BTN)
	if btnPwr == nil {
		return fmt.Errorf("failed to find power_btn %v", PWR_BTN)
	}

	// Set it as input with a pull down (it defaults to Low when unconnected).
	// NoEdge for the same reason as the WiFi button above.
	if err := btnPwr.In(gpio.PullDown, gpio.NoEdge); err != nil {
		return fmt.Errorf("can not access power_btn: %v", err)
	}

	//Power button
	go watchEdge(btnPwr, gpio.High, func() bool {
		if oledHalted {
			oledShow("\n\n   Screen ON", false)
		}

		if DEBUG_MODE {
			log.Printf("[     ] Button %s pushed", btnPwr)
		}

		holdCounter := 1
		for btnPwr.Read() == gpio.High {
			time.Sleep(1 * time.Second)
			holdCounter++
			if holdCounter > SHUTDOWN_COUNTDOWN {
				if DEBUG_MODE {
					log.Printf("[     ] Button %s held long enough. Triggering the action!", btnPwr)
				}
				systemShutdown("shutdown")
				// false stops the watcher, matching the old goroutine's
				// return: the gateway is shutting down anyway.
				return false
			}
		}
		return true
	})

	//

	log.Printf("[     ] Button manager initialized.")
	return nil
}

//
