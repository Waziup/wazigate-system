package api

import (
	// "fmt"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"io/ioutil"
	"log"
	"net/http"

	"sync"
	"time"

	"github.com/Waziup/wazigate-system/pkg/nm"
	"github.com/Waziup/wazigate-system/pkg/wazigate"
	routing "github.com/julienschmidt/httprouter"
	// In future we will use something like this lib to handle wifi stuff
	// wifi "github.com/mark2b/wpa-connect"
)

var wifiOperation = &sync.Mutex{}
const WAZIUP_CLOUD_URL string = "https://api.waziup.io/api/v2"

//

// GetNetInfo implements GET /net
func GetNetInfo(resp http.ResponseWriter, req *http.Request, params routing.Params) {

	r, err := nm.Devices()
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(err.Error()))
		return
	}
	resp.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(resp)
	err = encoder.Encode(r)
	if err != nil {
		log.Printf("[ERR  ] Can not encode nm.Devices: %v", err)
	}
}


// GetNetWiFi implements GET /net/wifi
func GetNetWiFi(resp http.ResponseWriter, req *http.Request, params routing.Params) {
	data, err := nm.Device("wlan0")
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(err.Error()))
		return
	}
	resp.Header().Set("Content-Type", "application/json")
	resp.Write(data)

}


type WifiReq struct {
	Enabled     bool   `json:"enabled"` // legacy, remove
	SSID        string `json:"ssid"`
	Autoconnect bool   `json:"autoConnect"`
	Password    string `json:"password"`
}

// SetNetWiFi implements POST/PUT /net/wifi
func SetNetWiFi(resp http.ResponseWriter, req *http.Request, params routing.Params) {

	var r WifiReq
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&r)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(err.Error()))
		return
	}
	if err := nm.Wifi(r.SSID, r.Password, r.Autoconnect); err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(err.Error()))
		return
	}
}

type DeleteWifiReq struct {
	SSID string `json:"ssid"`
}

// DeleteNetWiFi implements DELETE /net/wifi
func DeleteNetWiFi(resp http.ResponseWriter, req *http.Request, params routing.Params) {

	var r DeleteWifiReq
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&r)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(err.Error()))
		return
	}
	if err := nm.DeleteWifi(r.SSID); err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(err.Error()))
		return
	}
}


// Activate Access Point Mode
func ActivateAPMode() error {
	err := ioutil.WriteFile("/etc/do_not_reconnect_wifi", nil, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return nm.Hotspot("", "")
}

//

// implements POST /net/wifi/mode/ap
func SetNetAPMode(resp http.ResponseWriter, req *http.Request, params routing.Params) {
	if err := ActivateAPMode(); err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.Write([]byte("true"))
}

//

// Implements GET /net/wifi/scanning
func NetWiFiScan(resp http.ResponseWriter, req *http.Request, params routing.Params) {

	_, err := execOnHost("iwlist wlan0 scan")
	if err != nil {
		log.Printf("[ERR  ] Can not scan wifi: %v", err)
	}

	points, err := nm.ScanWifi()
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(err.Error()))
		return
	}

	if DEBUG_MODE {
		log.Printf("[     ] WiFi Scan: %v", points)
	}

	resp.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(resp)
	if err := encoder.Encode(points); err != nil {
		log.Printf("[ERR  ] Can not encode json: %v", err)
	}
}

// Implements GET /net/conns
func NetConns(resp http.ResponseWriter, req *http.Request, params routing.Params) {

	conns, err := nm.Connections()
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(resp)
	if err := encoder.Encode(conns); err != nil {
		log.Printf("[ERR  ] Can not encode json: %v", err)
	}
}

type AccessPointRequest struct {
	SSID     string `json:"ssid"`
	Password string `json:"password"`
}

// Implements POST|PUT /net/wifi/ap
func SetNetAP(resp http.ResponseWriter, req *http.Request, params routing.Params) {

	var r AccessPointRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&r)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(err.Error()))
		return
	}
	if err := nm.Hotspot(r.SSID, r.Password); err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte(err.Error()))
		return
	}
}



type VPNRequest struct {
	Enabled bool `json:"enabled"`
}

type VPNResponse struct {
	Connected bool   `json:"connected"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}


//

// Checks if Waziup cloud is accessible
func CloudAccessible(withLogs bool) bool {
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get("http://www.waziup.io/generate_204")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}

//

// InternetAccessible implements GET /internet
func InternetAccessible(resp http.ResponseWriter, req *http.Request, params routing.Params) {

	if CloudAccessible(true) {

		resp.Write([]byte("1"))

	} else {

		resp.Write([]byte("0"))
	}

}

func GoMonitor() error {
	ctx := context.Background()
	messages := make(chan interface{}, 1)
	go nm.Monitor(ctx, messages)
	go Monitor(messages)
	return nil
}

func Monitor(messages chan interface{}) {
	for msg := range messages {
		switch m := msg.(type) {
		case *nm.EventDeviceStateChanged:
			data, err := json.Marshal(m)
			if err != nil {
				log.Fatalf("[ERR  ] Can not marshal *EventDeviceStateChanged: %v", err)
			}
			if err := wazigate.Publish("waziup.wazigate-system/network-manager/device/"+m.Device, data); err != nil {
				log.Fatalf("[ERR  ] Can not publish MQTT message: %v", err)
			}
		}
	}
}
//========================================VPN FUNCTIONS=============================
// GetVPNStatus implements GET /net/vpn
func GetVPNStatus(resp http.ResponseWriter, req *http.Request, params routing.Params) {
	gatewayID, err := getGatewayID()

	if err != nil {
		errorResponse(resp,http.StatusInternalServerError,err.Error())
		return
	}
	connected, state, banner, err := nm.CheckVPNStatus(gatewayID)
	if err !=nil {
		errorResponse(resp,http.StatusBadRequest,"Invalid request body: " + err.Error())
		return 
	}
	log.Printf("VPN State: %v\n", state)
	
	if banner != "" {
		log.Printf("Server message: %s\n", banner)
	}
	connectedMessage :="VPN is not connected"
	if(connected){
		connectedMessage="VPN connection active"
	}

	resultResponse(resp,http.StatusOK, VPNResponse{
		Connected: connected,
		Message: connectedMessage+". VPN banner: "+banner,
	})
}
// PostVPN implements POST /net/vpn
func PostVPN(resp http.ResponseWriter, req *http.Request,  params routing.Params){
	var reqBody VPNRequest
	decoder :=json.NewDecoder(req.Body)
	err := decoder.Decode(&reqBody)
	if err !=nil {
		errorResponse(resp,http.StatusBadRequest,"could not decode ")
		return 
	}
	gatewayID, err := getGatewayID()

	if err != nil {
		log.Printf("error encountered %v",err)
		errorResponse(resp,http.StatusInternalServerError,err.Error())
		return
	}
	
	connected, activeConn, err := nm.IsVPNConnected(gatewayID)
	if err != nil {
		errorResponse(resp,http.StatusInternalServerError,fmt.Sprintf("error checking VPN status: %v", err.Error()))
		return
	}

	if !reqBody.Enabled {
		if !connected {
			log.Println("VPN is not connected")
			errorResponse(resp,http.StatusBadRequest,"VPN is not connected")
			return 
		}
		if err := nm.DisconnectVPN(activeConn);err !=nil{
			log.Printf("error encountered %v",err)
			errorResponse(resp,http.StatusInternalServerError,fmt.Sprintf("failed to disconnect VPN %s",err.Error()))
			return 
		}
	}
	if connected {
		log.Println(" VPN is already active")
		errorResponse(resp,http.StatusBadRequest,"VPN is already active",)
		return
	}
	conn, exists,err := nm.VpnProfileExists(gatewayID)
	if err != nil {
		errorResponse(resp,http.StatusInternalServerError,fmt.Sprintf("error getting active vpn %s",err.Error()))
		return
	}
	if !exists {
		configFile := gatewayID +".ovpn"
		if err := downloadVPNConfig(gatewayID, configFile); err !=nil {
			log.Printf("could not download vpn file %s",err.Error())
			errorResponse(resp,http.StatusInternalServerError,"could not download vpn file")
			return 
		}
		conn, err = nm.ImportVPN(configFile)
		if err !=nil {
			errorResponse(resp,http.StatusInternalServerError,fmt.Sprintf("could not import vpn profile %s",err.Error()))
			return
		}
	}
	err = nm.ConnectVPN(conn)
	if err != nil {
		errorResponse(resp,http.StatusInternalServerError,fmt.Sprintf("could not connect to vpn network %s",err.Error()))
		return
	}

	
	action := "disabled"
	if reqBody.Enabled {
		action = "enabled"
	}
	resultResponse(resp,http.StatusOK,VPNResponse{
		Connected: reqBody.Enabled,
		Message: "VPN "+action +" successfully.",
	})
}
func getGatewayID() (string, error) {
	resp, err := http.Get("http://waziup.wazigate-edge/device/id")
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read error: %v", err)
	}
	cleanId := strings.ToLower(strings.ReplaceAll(string(body),`"`,""))
	log.Printf("Gateway id is %s.",cleanId)
	return "wazigate-" + cleanId, nil

}

func downloadVPNConfig(gatewayID, outputFile string) error {
	url := fmt.Sprintf("%s/gateways/%s/vpn",WAZIUP_CLOUD_URL, gatewayID)
	log.Printf("Getch url =%s",url)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch config: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error getting VPN config file status %s. code: %d",resp.Status, resp.StatusCode, )
	}

	out, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}

	log.Printf("Downloaded config: %s", outputFile)
	return nil
}
func resultResponse(w http.ResponseWriter, code int, payload interface{}){
	data, err := json.Marshal(payload)
	if err !=nil {
		log.Printf("Failed to marshal JSON response %v",payload)
		w.WriteHeader(500)
		return
	}
	w.Header().Add("Content-Type","application/json")
	w.WriteHeader(code)
	w.Write(data)
}
func errorResponse(resp http.ResponseWriter, code int, msg  string)  {
	resultResponse(resp, code, VPNResponse{
		Error: msg,
		Connected: false,
	})
}
