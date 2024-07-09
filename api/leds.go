package api

import (
	"log"
	"sync"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
)

//

const LED1_PIN = "GPIO27" // PIN #13

//

// This function controls the LED indicator
func RunLEDManager() error {

	go func() {
		blinkStart(LED1_PIN, 500, 500)

		// Wait for the host to come up before sending any command
		for {
			if hostReady() {
				break
			}
			time.Sleep(5 * time.Second)
		}

		for {
			if CloudAccessible(false /*Without Logs*/) {
				turnOnLED(LED1_PIN)
			} else {
				blinkStart(LED1_PIN, 100, 100)
			}
			time.Sleep(3 * time.Second)
		}
	}()

	log.Printf("[     ] LED manager initialized.")
	return nil
}

//

var ledLock1 chan struct{}
var wg1 sync.WaitGroup

// This function receives a GPIO attached to a LED, an ON time duration and an OFF time duration
// and blinks the LED accordingly
func blinkStart(ledPin string, onTime time.Duration, offTime time.Duration) {

	blinkStop(ledPin) // Clear blinking if it is already blinking...

	go func() {

		pin := gpioreg.ByName(ledPin) // LED pin

		ledLock1 = make(chan struct{}, 1)
		quitLock := &ledLock1
		wg1.Add(1)
		defer wg1.Done()

		for {

			select {
			case <-*quitLock:
				return
			default:
				if err := pin.Out(gpio.High); err != nil {
					log.Printf("[ERR  ]: LED %s ", err.Error())
				}

				time.Sleep(onTime * time.Millisecond)

				if err := pin.Out(gpio.Low); err != nil {
					log.Printf("[ERR  ]: LED %s ", err.Error())
				}

				time.Sleep(offTime * time.Millisecond)
			}

		}
	}()
}

//

// This function receives a GPIO attached to a LED and stops the blinking if it is blinking
func blinkStop(ledGpio string) {

	if ledGpio == LED1_PIN {
		select {
		case ledLock1 <- struct{}{}:
			wg1.Wait()
			close(ledLock1)
			ledLock1 = nil
		default:
		}
		return
	}
}

//

// This function receives a GPIO attached to a LED and turns on the LED
func turnOnLED(ledGpio string) {

	blinkStop(ledGpio)

	pin := gpioreg.ByName(ledGpio) // LED pin

	if err := pin.Out(gpio.High); err != nil {
		log.Printf("[ERR  ]: LED %s ", err.Error())
	}
}
