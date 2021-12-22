// This package handles all the APIs provided by `wazigate-system`
package api

import (
	"log"
	"net/http"
	"os"

	routing "github.com/julienschmidt/httprouter"
	"periph.io/x/periph/host"
)

//

var DEBUG_MODE bool    //DEBUG mode sends the errors via the HTTP responds
var WIFI_DEVICE string //Wifi Interface which can be set via env
var ETH_DEVICE string  //Ethernet Interface

var Config Configuration // the main configuration object

func Init() error {
	if _, err := host.Init(); err != nil {
		return err
	}

	Config = loadConfigs()

	//

	if os.Getenv("DEBUG_MODE") == "1" {
		log.Println("[     ] Debug Mode is activated.")
		DEBUG_MODE = true
	} else {
		DEBUG_MODE = false
	}

	//

	WIFI_DEVICE = "wlan0"

	if os.Getenv("WIFI_DEVICE") != "" {
		WIFI_DEVICE = os.Getenv("WIFI_DEVICE")
	}

	ETH_DEVICE = "eth0"
	if os.Getenv("ETH_DEVICE") != "" {
		ETH_DEVICE = os.Getenv("ETH_DEVICE")
	}
	return nil
}

//

// HomeLink implements GET / Just a test msg to see if it works
func HomeLink(resp http.ResponseWriter, req *http.Request, params routing.Params) {
	resp.Write([]byte("Salam Goloooo, It works!"))
}

var PackageJSON []byte // set by main.go

func packageJSON(resp http.ResponseWriter, req *http.Request, params routing.Params) {
	resp.Header().Set("Content-Type", "application/json")
	resp.Write(PackageJSON)
}

//

// APIDocs API documents (Swagger)
func APIDocs(resp http.ResponseWriter, req *http.Request, params routing.Params) {
	// log.Println( req.URL.Path)

	rootPath := os.Getenv("EXEC_PATH")
	if rootPath == "" {
		rootPath = "./"
	}

	http.FileServer(http.Dir(rootPath)).ServeHTTP(resp, req)
}

//

// UI implements HTTP /ui
func UI(resp http.ResponseWriter, req *http.Request, params routing.Params) {

	rootPath := os.Getenv("EXEC_PATH")
	if rootPath == "" {
		rootPath = "./"
	}

	http.FileServer(http.Dir(rootPath)).ServeHTTP(resp, req)
}

//
