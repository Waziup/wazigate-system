package api

import (
	"log"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
)

/*-------------------------*/

const LED1_PIN = "GPIO27" // PIN #13
const LED2_PIN = "GPIO22" // PIN #15

/*-------------------------*/

func LEDsLoop() {

	go func() {
		if _, err := host.Init(); err != nil {
			log.Printf("[Err   ]: %s ", err.Error())
		}

		blinkStart(LED1_PIN, 500, 500)
		time.Sleep(500 * time.Millisecond)
		blinkStart(LED2_PIN, 500, 500)

		// Wait for the host to come up before sending any command
		for {
			if hostReady() {
				break
			}
			time.Sleep(5 * time.Second)
		}

		/*----------*/

		for {

			/*----------*/

			netInfo, err := getNetWiFi()
			if err != nil {
				log.Printf("[ERR  ] Get wifi info: %v", err)
				blinkStart(LED2_PIN, 50, 100)

			} else {

				if netInfo["ap_mode"].(bool) {
					blinkStart(LED2_PIN, 1000, 1000)

				} else if netInfo["ip"].(string) != "" {
					blinkStart(LED2_PIN, 100, 2000)

				} else {
					blinkStart(LED2_PIN, 100, 100)
				}

			}

			/*----------*/

			if CloudAccessible(false /*Without Logs*/) {
				blinkStart(LED1_PIN, 100, 2000)
			} else {
				blinkStart(LED1_PIN, 100, 100)
			}

			/*----------*/

			time.Sleep(3 * time.Second)
		}

	}()

	log.Printf("[Info  ] LED manager initialized.")
}

/*-------------------------*/

var ledLock1, ledLock2 chan struct{}

func blinkStart(ledGpio string, onTime time.Duration, offTime time.Duration) {

	blinkStop(ledGpio) // Clear blinking if it is already blinking...
	go func(ledGpio string) {

		var quit chan struct{}
		switch ledGpio {
		case LED1_PIN:
			{
				ledLock1 = make(chan struct{})
				quit = ledLock1
			}
		case LED2_PIN:
			{
				ledLock2 = make(chan struct{})
				quit = ledLock2
			}
		}

		pin := gpioreg.ByName(ledGpio) // LED pin
		pin.Out(gpio.Low)

		for {
			select {
			case <-quit:
				// done.Done()
				return
			default:
				{

					if err := pin.Out(gpio.High); err != nil {
						log.Printf("[Err   ]: LED %s ", err.Error())
					}

					time.Sleep(onTime * time.Millisecond)

					if err := pin.Out(gpio.Low); err != nil {
						log.Printf("[Err   ]: LED %s ", err.Error())
					}

					time.Sleep(offTime * time.Millisecond)
				}
			}

		}
	}(ledGpio)
}

/*-------------------------*/
func blinkStop(ledGpio string) {

	if ledGpio == LED1_PIN && ledLock1 != nil {
		ledLock1 <- struct{}{}
		close(ledLock1)
		return
	}

	if ledGpio == LED2_PIN && ledLock2 != nil {
		ledLock2 <- struct{}{}
		close(ledLock2)
		return
	}
}

/*-------------------------*/
