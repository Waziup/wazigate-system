package api

import (
	// "fmt"
	"net/http"
	"os"

	// "io/ioutil"
	// "path/filepath"
	"log"
	// "encoding/json"

	routing "github.com/julienschmidt/httprouter"
)

/*-------------------------*/

var DEBUG_MODE bool    //DEBUG mode sends the errors via the HTTP responds
var WIFI_DEVICE string //Wifi Interface which can be set via env
var ETH_DEVICE string  //Ethernet Interface

var Config Configuration // the main configuration object

/*-------------------------*/

func init() {

	// Remove date and time from logs
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	log.SetOutput(os.Stdout)

	log.Printf("[Info  ] Initializing the System")

	Config = loadConfigs()

	/*-----------*/

	if os.Getenv("DEBUG_MODE") == "1" {
		log.Println("[Info  ]: DEBUG_MODE mode is activated.")
		DEBUG_MODE = true

	} else {

		log.Println("[Info  ]: DEBUG_MODE mode is NOT activated.")
		DEBUG_MODE = false
	}

	/*-----------*/

	WIFI_DEVICE = "wlan0"

	if os.Getenv("WIFI_DEVICE") != "" {
		WIFI_DEVICE = os.Getenv("WIFI_DEVICE")
	}

	ETH_DEVICE = "eth0"
	if os.Getenv("ETH_DEVICE") != "" {
		ETH_DEVICE = os.Getenv("ETH_DEVICE")
	}

	/*-----------*/

	TimezoneInit()
	BlackoutLoop()
	NetworkLoop()
	ButtonsLoop()
	OledLoop()
	FanLoop()

	/*-----------*/
}

/*-------------------------*/

// HomeLink implements GET / Just a test msg to see if it works
func HomeLink(resp http.ResponseWriter, req *http.Request, params routing.Params) {

	resp.Write([]byte("Salam Goloooo, It works!"))
}

/*-------------------------*/

// APIDocs API documents
func APIDocs(resp http.ResponseWriter, req *http.Request, params routing.Params) {
	// log.Println( req.URL.Path)
	http.FileServer(http.Dir("./")).ServeHTTP(resp, req)
}

/*-------------------------*/

// UI implements HTTP /ui
func UI(resp http.ResponseWriter, req *http.Request, params routing.Params) {
	// log.Println( req.URL.Path)
	// log.Println( params.ByName( "file_path"))

	http.FileServer(http.Dir("./")).ServeHTTP(resp, req)
}

/*-------------------------*/
