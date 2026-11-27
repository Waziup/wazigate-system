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

	// Set it as input with a pull down (it defaults to Low when unconnected).
	//
	// NoEdge rather than FallingEdge: periph.io implements edge detection via
	// the kernel's legacy sysfs GPIO interface, which the trixie kernel no
	// longer numbers the way v3.6.7 expects, so asking for an edge here fails
	// outright. watchEdge polls the memory-mapped registers instead; see
	// gpiopoll.go for the full story.
	if err := inPin.In(gpio.PullDown, gpio.NoEdge); err != nil {
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

	// The circuit pulls the pin low when mains power is cut, so what used to be
	// FallingEdge is a transition to Low.
	go watchEdge(inPin, gpio.Low, func() bool {

		if DEBUG_MODE {
			log.Printf("[     ] %s Blackout signal received", inPin)
		}

		oledWrite("\n\n    BLACKOUT!")

		systemQuickShutdown()
		// time.Sleep( 1 * time.Second)
		return true
	})

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
