package api

import (
	"fmt"
	"log"

	// "time"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"

	// "strconv"

	"encoding/json"
	"net/http"

	routing "github.com/julienschmidt/httprouter"
)

//

// The GPIO pin that receives a signal when the main power is cut
const POWER_IN_PIN = "GPIO23" // PIN #16

// If the hat has the power input, we monitor power input for blackout
var blackoutEnabled bool

//

// This function monitors the power input in the new boards
// and check if there is black out then shuts down the gateway
func RunBlackoutManager() error {

	inPin := gpioreg.ByName(POWER_IN_PIN)
	if inPin == nil {
		return fmt.Errorf("failed to find power_in_pin %v", POWER_IN_PIN)
	}

	// Set it as input, with a pull down (defaults to Low when unconnected) and
	// enable rising edge triggering.
	if err := inPin.In(gpio.PullDown, gpio.FallingEdge); err != nil {
		return fmt.Errorf("failed to access power_in_pin %v", err.Error())
	}

	blackoutEnabled = inPin.Read() == gpio.High // If we have power input that means we have that circuit enabled on the board

	if blackoutEnabled {

		log.Printf("[     ] Blackout monitor Enabled.")

	} else {

		log.Printf("[     ] Blackout monitor NOT enabled.")
		return nil
	}

	//

	go func() {

		for inPin.WaitForEdge(-1) {

			if DEBUG_MODE {
				log.Printf("[     ] %s Blackout signal received", inPin)
			}

			oledWrite("\n\n    BLACKOUT!")

			systemQuickShutdown()
			// time.Sleep( 1 * time.Second)
		}
	}()

	//

	log.Printf("[     ] Blackout monitor initialized.")

	return nil
}

//

// This function implements GET /blackout
// It returns a boolean value indicating whether the Wazigate is equipped with a blackout protection circuit
func BlackoutEnabled(resp http.ResponseWriter, req *http.Request, params routing.Params) {

	outJson, err := json.Marshal(blackoutEnabled)
	if err != nil {
		log.Printf("[ERR  ] %s", err.Error())
	} /**/

	resp.Write([]byte(outJson))
}

//
