// This package handles all the APIs provided by `wazigate-system`
package api

import (
	"bytes"
	"io"
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

var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

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

var client = &http.Client{}

func SSH(resp http.ResponseWriter, req *http.Request, params routing.Params) {

	// Proxy from client to host:
	// /ssh/index.html -> http://localhost:4200/index.html
	reqURL := "http://wazigate:4200" + req.RequestURI[4:]

	var body io.Reader = req.Body
	if req.Method == "POST" {
		buf, _ := io.ReadAll(req.Body)
		body = bytes.NewBuffer(buf)
	}

	req2, err := http.NewRequest(req.Method, reqURL, body)
	if err != nil {
		log.Println("SSH Proxy Error", err)
		http.Error(resp, "Error in Request", http.StatusInternalServerError)
		return
	}

	copyHeader(req2.Header, req.Header)
	delHopHeaders(req2.Header)
	req2.Header.Set("Connection", "keep-alive")

	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		log.Println("SSH Proxy Error", err)
		http.Error(resp, "Server Error", http.StatusInternalServerError)
		return
	}

	// Proxy back to client:
	delHopHeaders(resp2.Header)
	copyHeader(resp.Header(), resp2.Header)
	resp.WriteHeader(resp2.StatusCode)
	io.Copy(resp, resp2.Body)

	req.Body.Close()
	resp2.Body.Close()
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

//
