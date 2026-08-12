package api

import (
	"log"
	"strings"
	"time"

	"net/http"

	"io/ioutil"

	routing "github.com/julienschmidt/httprouter"

	"image"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/devices/ssd1306"
	"periph.io/x/periph/devices/ssd1306/image1bit"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

//

var OledBuffer string     // A Shared buffer for showing message on the OLED
var OledCurrentMsg string // The message which is showing on the OLED at the moment

// How often the screen is redrawn. The heartbeat indicator and the timeouts
// below are expressed as durations, so they stay correct if this is changed.
const oledRefreshInterval = 1 * time.Second

// For how long a message written through POST /oled stays on the screen if
// nobody clears it.
const oledMessageTimeout = 13 * time.Second

// The container states are read from the docker socket on the host, which costs
// a process spawn there, so they are not fetched on every redraw.
const bootStatusInterval = 10 * time.Second

//

var oledDev *ssd1306.Dev
var oledDoesNotExist bool // We check if the OLED does not exist, we just ignore it.

// var oledHaltTimeout		int			// Off timeout value in seconds
var oledHalted bool // If OLED is off (we turn it off after some time of not use and come back after a push button)

//

// This function initializes the OLED
func oledInit() error {

	oledDoesNotExist = false
	oledHalted = false

	// Use i2creg I²C bus registry to find the first available I²C bus.
	b, err := i2creg.Open("")
	if err != nil {
		oledDoesNotExist = true
		return err
	}

	oledDev, err = ssd1306.NewI2C(b, &ssd1306.DefaultOpts)
	if err != nil {
		oledDoesNotExist = true
		return err
	}
	return nil
}

//

// This function make the OLED black
func oledHalt() {

	if oledDev == nil {
		log.Printf("[ERR  ] OLED halt: No OLED Found!")
		return
	}
	oledShow("\n\n   Screen OFF", false)
	time.Sleep(1 * time.Second)
	oledHalted = true
	err := oledDev.Halt()
	if err != nil {
		log.Printf("[ERR  ] OLED halt: %s ", err.Error())
	}
}

//

// This function shows a given message on the OLED
func oledShow(msg string, withLogs bool) {

	if oledDoesNotExist {
		if withLogs && DEBUG_MODE {
			log.Printf("[OLED  ] Oled display does not exist!")
		}
		return
	}

	if OledCurrentMsg == msg {
		return // Do nothing
	}
	OledCurrentMsg = msg

	if oledHalted {
		oledHalted = false
	}

	if withLogs && DEBUG_MODE {
		log.Printf("[OLED  ] \"%s\"", msg)
	}

	// Draw on it.
	img := image1bit.NewVerticalLSB(oledDev.Bounds())
	// Note: this code is commented out so periph does not depend on:

	f := basicfont.Face7x13
	drawer := font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{image1bit.On},
		Face: f,
		// Dot:  fixed.P(0, img.Bounds().Dy()-1-f.Descent),
		Dot: fixed.P(0, f.Height-f.Descent),
	}

	lines := strings.Split(msg, "\n")
	for i, line := range lines {
		line = strings.TrimRight(string(line), " \n\t\r")

		drawer.Dot = fixed.P(0, (i+1)*f.Height-f.Descent)
		drawer.DrawString(line)
	}

	if err := oledDev.Draw(oledDev.Bounds(), img, image.Point{}); err != nil {
		log.Printf("[ERR  ] OLED [ %s ] command. \n\tError: [ %s ]", msg, err.Error())

		//Wait for a while and try again after failure
		time.Sleep(2 * time.Second)
		oledInit()
	}

}

//

// OLED Controller
// This function gets the gateway status periodically and updates the OLED
func RunOLEDManager() error {

	if err := oledInit(); err != nil {
		log.Printf("[WARN ] OLED init failed: %v", err)
	}

	if oledDoesNotExist {
		log.Println("[     ] The OLED could not be initialized, so the OLED functions will not be available.")
		return nil
	}

	go func() {

		OledBuffer = "" // Clear the buffer

		var msgShownFor time.Duration // For how long the current message is on the screen
		var screenOnFor time.Duration // For how long the screen has been on

		allBootedOK := false
		bootStatusMsg := ""
		var lastBootCheck time.Time // Zero value forces a check on the first iteration

		heartbeat := false // Just a toggle varianle to show heartbeat on the screen

		for {

			//

			// Automatically clear a message if it is not removed in time
			if msgShownFor > oledMessageTimeout {
				OledBuffer = ""
				oledShow("", false)
			}

			if len(OledBuffer) > 0 {
				oledShow(OledBuffer, true)
				msgShownFor += oledRefreshInterval
				time.Sleep(oledRefreshInterval)
				continue
			}
			msgShownFor = 0

			if oledHalted {
				time.Sleep(oledRefreshInterval)
				continue
			}

			if screenOnFor > time.Duration(Config.OLEDHaltTimeout)*time.Second {
				oledHalt()
				screenOnFor = 0
				continue
			}

			screenOnFor += oledRefreshInterval

			//

			if time.Since(lastBootCheck) >= bootStatusInterval {
				allBootedOK, bootStatusMsg = GetGWBootstatus(false)
				lastBootCheck = time.Now()
			}

			// While the gateway is still coming up, the container states are
			// more useful on the screen than the network status.
			if !allBootedOK {
				oledShow(bootStatusMsg, false)
				time.Sleep(oledRefreshInterval)
				continue
			}

			//

			heartTxt := "  "
			heartbeat = !heartbeat
			if heartbeat {
				heartTxt = "* "
			}

			// The cloud state comes from the cloud monitor, reading it does not
			// send anything over wlan0 or the modem.
			netTxt := "[ Internet NO ]"
			if CloudAccessible(false /*Without Logs*/) {
				netTxt = "[ Internet OK ]"
			}

			OledMsg := heartTxt + netTxt

			//

			// eip, wip, aip, ssid := GetAllIPs()

			// if len(eip) > 0 {
			// 	// msg.append( "Ethernet: "+ eip);
			// 	OledMsg += "\nEth: " + eip
			// }

			// if len(wip) > 0 {
			// 	OledMsg += "\n\nWiFi: (" + ssid + ")\n " + wip
			// }

			// if len(aip) > 0 {
			// 	OledMsg += "\n\nAP: (" + ssid + ")\n " + aip
			// }

			//

			oledShow(OledMsg, false)
			time.Sleep(oledRefreshInterval)

		} // End of `for`

	}()

	log.Printf("[     ] OLED manager initialized.")
	return nil

}

//

// Simply write the incoming message into the OLED buffer to be shown
func oledWrite(msg string) {
	OledBuffer = msg
}

//

// This function implements POST|PUT /oled API
func OledWriteMessage(resp http.ResponseWriter, req *http.Request, params routing.Params) {

	msg, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("[ERR  ] OLED [ %s ] command. \n\tError: [ %s ]", msg, err.Error())
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}

	oledWrite(string(msg))
}

//
