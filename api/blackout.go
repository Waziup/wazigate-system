package api

/*
* This file monitors the power input in the new boards 
* and check if there is black out then shuts down the gateway
*/

import (
	"log"
	// "time"
    "periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host"
	"periph.io/x/periph/conn/gpio/gpioreg"
	// "strconv"
	
	"encoding/json"
	"net/http"
	routing "github.com/julienschmidt/httprouter"
)

/*-------------------------*/

const	POWER_IN_PIN		=	"GPIO23"		// PIN #16
var		blackoutEnabled		bool				// If the hat has the power input, we monitor power input for blackout

/*-------------------------*/

func BlackoutLoop(){
	
	log.Printf( "[Info  ] Blackout monitor initialized.")
	
	if _, err := host.Init(); err != nil {
		log.Printf( "[Err   ]: %s ", err.Error())
	}

	inPin	:= gpioreg.ByName( POWER_IN_PIN)
	if inPin == nil {
		log.Printf( "[Err   ] Failed to find %v", POWER_IN_PIN)
	}

	// Set it as input, with a pull down (defaults to Low when unconnected) and
	// enable rising edge triggering.
	if err := inPin.In( gpio.PullDown, gpio.FallingEdge); err != nil {
		log.Printf( "[Err   ]: %s ", err.Error())
	}

	blackoutEnabled = inPin.Read() == gpio.High // If we have power input that means we have that circuit enabled on the board

	if( blackoutEnabled) {

		log.Printf( "[Info  ] Blackout monitor Enabled.")

	}else{

		log.Printf( "[Info  ] Blackout monitor NOT enabled.")
		return;
	}


	/*---------*/

	go func(){

		for inPin.WaitForEdge(-1) {

			if( DEBUG_MODE){
				log.Printf( "[Info   ] %s Blackout signal received", inPin)
			}

			oledWrite( "\n\n    BLACKOUT!")

			systemQuickShutdown();
			// time.Sleep( 1 * time.Second)
		}
	}()

	/*---------*/
	
}

/*-------------------------*/

func BlackoutEnabled( resp http.ResponseWriter, req *http.Request, params routing.Params){

	outJson, err := json.Marshal( blackoutEnabled)
	if( err != nil) {
		log.Printf( "[Err   ] %s", err.Error())
	}/**/

	resp.Write( []byte( outJson))
}

/*-------------------------*/