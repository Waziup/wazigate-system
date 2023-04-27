package api

import (
	// "fmt"
	"context"
	"encoding/json"
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

//

// GetGWID implements GET /gwid
// This is Deprecated, now we call an API on the `wazigate-edge` to get the gateway Id
// func GetGWID(resp http.ResponseWriter, req *http.Request, params routing.Params) {

// 	interfs, err := net.Interfaces()
// 	if err != nil {
// 		log.Printf("[ERR  ] %s", err.Error())

// 		if DEBUG_MODE {
// 			http.Error(resp, "[ Error ]: "+err.Error(), http.StatusInternalServerError)
// 		}
// 	}

// 	localID := ""

// 	for _, interf := range interfs {
// 		addr := interf.HardwareAddr.String()
// 		if addr != "" {
// 			localID = strings.ReplaceAll(addr, ":", "")
// 			break
// 		}
// 	}
// 	resp.Write([]byte(localID))
// }

//

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

	// out, err := getNetWiFi()
	// if err != nil {
	// 	log.Printf("[ERR  ] %s", err.Error())
	// 	http.Error(resp, "[ Error ]: "+err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// outJSON, err := json.Marshal(out)
	// if err != nil {
	// 	log.Printf("[ERR  ] %s", err.Error())
	// }

	// resp.Write([]byte(outJSON))
}

//

// func getWiFiClientStatus() (map[string]interface{}, error) {

// 	cmd := "wpa_cli status -i " + WIFI_DEVICE
// 	stdout, err := execOnHost(cmd)
// 	if err != nil {
// 		return nil, err // The WiFi is not connected
// 	}

// 	/*--------*/

// 	re := regexp.MustCompile(`([\w]+)=(.*)`)
// 	var status map[string]string = make(map[string]string)

// 	subMatchAll := re.FindAllStringSubmatch(string(stdout), -1)
// 	for _, element := range subMatchAll {
// 		status[element[1]] = element[2]
// 	}

// 	// All possible keys:
// 	// bssid=9c:c8:fc:29:e5:e0
// 	// freq=2412
// 	// ssid=GoliNet
// 	// id=0
// 	// mode=station
// 	// pairwise_cipher=CCMP
// 	// group_cipher=CCMP
// 	// key_mgmt=WPA2-PSK
// 	// wpa_state=COMPLETED
// 	// ip_address=192.168.200.1  /* Not very accurate*/
// 	// p2p_device_address=f6:8d:01:5d:ae:28
// 	// address=b8:27:eb:49:66:e2
// 	// uuid=087d50a2-7a1c-589c-bcec-cd5acde1ff57

// 	ssid := status["ssid"]
// 	freq := status["freq"]
// 	state := status["wpa_state"]

// 	/*--------*/

// 	return map[string]interface{}{
// 		"ssid":  ssid,
// 		"freq":  freq,
// 		"state": state,
// 	}, nil
// }

//

// func getNetWiFi() (map[string]interface{}, error) {

// 	/*iwifi, err := net.InterfaceByName( WIFI_DEVICE)
// 	if err != nil {
// 		log.Printf( "[ERR  ] %s", err.Error())
// 		if( DEBUG_MODE){
// 			http.Error( resp, "[ Error ]: "+ err.Error(), http.StatusInternalServerError)
// 		}

// 	}

// 	addrs, err := iwifi.Addrs();
// 	if err != nil {
// 		log.Printf( "[ERR  ] %s", err.Error())
// 		if( DEBUG_MODE){
// 			http.Error( resp, "[ Error ]: "+ err.Error(), http.StatusInternalServerError)
// 		}
// 	}

// 	ip := ""
// 	if len( addrs) > 0 {
// 		ip = addrs[0].(*net.IPNet).IP.String()
// 	} /**/

// 	/*-----*/

// 	cmd := "ip -4 addr show " + WIFI_DEVICE + " | awk '$1 == \"inet\" {gsub(/\\/.*$/, \"\", $2); print $2}'"
// 	ip, _ := exeCmd(cmd)

// 	/*-----*/

// 	cmd = "ip link show up " + WIFI_DEVICE
// 	outc, _ := exeCmd(cmd)
// 	enabled := outc != ""

// 	/*-----*/

// 	cmd = "iw " + WIFI_DEVICE + " info | grep ssid | awk '{print $2\" \"$3\" \"$4\" \"$5\" \"$6}'"
// 	outc, _ = exeCmd(cmd)
// 	ssid := outc

// 	/*-----*/

// 	cmd = "systemctl is-active --quiet hostapd && echo 1"
// 	outc, err := execOnHost(cmd)
// 	apMode := outc == "1"
// 	if err != nil {
// 		apMode = false // we may chnage this
// 	}

// 	/*-----*/

// 	wifiClientStatus, err := getWiFiClientStatus()
// 	state := ""
// 	if err == nil {
// 		state = wifiClientStatus["state"].(string)
// 	}

// 	//

// 	return map[string]interface{}{
// 		"ip":      ip,
// 		"enabled": enabled,
// 		"ssid":    ssid,
// 		"ap_mode": apMode,
// 		"state":   state,
// 	}, nil

// }

//

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

//

// func startWiFiClient() error {

// 	oledWrite("\nConnecting to\n   WiFi...")
// 	stdout, err := execOnHost("sudo bash start_wifi.sh")

// 	if err != nil {
// 		log.Printf("[HOST  ] %s \t %s", err.Error(), stdout)

// 	} else {

// 		log.Printf("[     ] %s", stdout)
// 	}
// 	oledWrite("") // Clean the OLED msg

// 	return err
// }

//

// This function determines if the gateway is in the Access Point Mode
// func apMode(withLogs bool) bool {

// 	apAtive, _ := execOnHostWithLogs("systemctl is-active --quiet hostapd && echo 1", withLogs)
// 	return apAtive == "1"
// }

//

// CheckWlanConn checks the status of WiFi and takes proper actions
// Return: fasle = AP Mode , true = WiFi Client mode
// func CheckWlanConn() bool {

// 	wifiOperation.Lock()

// 	if apMode(true) {
// 		ActivateAPMode()
// 		wifiOperation.Unlock()
// 		return false // AP mode is active
// 	}

// 	wifiOperation.Unlock()

// 	// Give it some time to connect, then we check
// 	time.Sleep(2 * time.Second)
// 	for i := 0; i < 10; i++ {

// 		oledWrite("\nChecking WiFi." + strings.Repeat(".", i))

// 		// cmd = "iw " + WIFI_DEVICE + " info | grep ssid | awk '{print $2\" \"$3\" \"$4\" \"$5\" \"$6}'"
// 		// wifiRes, err := execOnHost("iwgetid")
// 		// if err == nil && wifiRes != "" {
// 		// 	oledWrite("") // Clean the OLED
// 		// 	return true
// 		// }

// 		wifiOperation.Lock()

// 		wifiClientStatus, err := getWiFiClientStatus()

// 		wifiOperation.Unlock()
// 		state := ""
// 		if err == nil {
// 			state = wifiClientStatus["state"].(string)
// 			if state == "COMPLETED" {
// 				return true
// 			}

// 			// time.Sleep(1 * time.Second)
// 			oledWrite("\nWiFi state:\n  " + state)
// 		}

// 		time.Sleep(5 * time.Second)
// 		// oledWrite("")
// 		// time.Sleep(2 * time.Second)
// 	}

// 	//Could no conenct, need to revert to AP setting

// 	if DEBUG_MODE {
// 		log.Printf("[     ] Could not connect!\nReverting the settings...")
// 	}
// 	oledWrite("Cannot Connect\n\nReverting to \n  Access point...")
// 	time.Sleep(2 * time.Second)

// 	wifiOperation.Lock()

// 	ActivateAPMode()

// 	wifiOperation.Unlock()

// 	return false
// }

//

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

//

// Implements GET /net/wifi/ap
// func GetNetAP(resp http.ResponseWriter, req *http.Request, params routing.Params) {
// 	data, err := nm.Device("wlan0")
// 	if err != nil {
// 		resp.WriteHeader(http.StatusInternalServerError)
// 		resp.Write([]byte(err.Error()))
// 		return
// 	}
// 	resp.Header().Set("Content-Type", "application/json")
// 	resp.Write(data)
// }

//

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

//

// Checks if Waziup cloud is accessible
func CloudAccessible(withLogs bool) bool {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get("http://www.waziup.io/")
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

//

// This function retrieves the IP addesses of all connected network interfaces (e.g. Wlan, Ethernet)
// It is usually used by the OLED controller
// func GetAllIPs() (string, string, string, string) {

// 	cmd := "iw " + WIFI_DEVICE + " info | grep ssid | awk '{print $2\" \"$3\" \"$4\" \"$5\" \"$6}'"
// 	ssid, _ := exeCmdWithLogs(cmd, false)

// 	cmd = "status=$(ip addr show " + WIFI_DEVICE + " | grep \"state UP\"); if [ \"$status\" == \"\" ]; then echo \"\"; else echo $(ip -4 addr show " + WIFI_DEVICE + " | awk '$1 == \"inet\" {gsub(/\\/.*$/, \"\", $2); print $2}');  fi;"
// 	wip, _ := exeCmdWithLogs(cmd, false)
// 	aip := wip

// 	cmd = "status=$(ip addr show " + ETH_DEVICE + " | grep \"state UP\"); if [ \"$status\" == \"\" ]; then echo \"NOT Connected\"; else echo $(ip -4 addr show " + ETH_DEVICE + " | awk '$1 == \"inet\" {gsub(/\\/.*$/, \"\", $2); print $2}');  fi;"
// 	eip, _ := exeCmdWithLogs(cmd, false)

// 	if apMode(false) {
// 		wip = ""
// 	} else {
// 		aip = ""
// 	}

// 	return eip, wip, aip, ssid
// }

//

// NetworkLoop provides a constant connectivity check
// func NetworkLoop() {

// 	// Connecting might take some time, so throw it into another thread ;)
// 	go func() {

// 		// Wait for the host to come up before sending any command
// 		for {
// 			if hostReady() {
// 				oledWrite("\n \n    HOST READY")
// 				break
// 			}
// 			log.Println("[     ] Waiting for the HOST...")
// 			oledWrite("\n Waiting\n     for \n   the HOST")
// 			time.Sleep(2 * time.Second)
// 		}

// 		// Check WiFi Connectivity
// 		if CheckWlanConn() {
// 			oledWrite("\n WiFi Connected ")
// 			if DEBUG_MODE {
// 				log.Println("[     ] WiFi Connected.")
// 			}
// 		}

// 		// In order to provide a stable connectivity,
// 		// let's not rely on the OS and check the WiFi connection periodically
// 		time.Sleep(60 * time.Second)
// 		for {
// 			// We need to avoid race condition
// 			wifiOperation.Lock()

// 			wifiStatus, err := getNetWiFi()

// 			if err == nil {
// 				if apMode, ok := wifiStatus["ap_mode"]; ok {

// 					SSID, okSSID := wifiStatus["ssid"]
// 					IP, okIP := wifiStatus["ip"]

// 					if (okSSID && SSID == "") || (okIP && IP == "") {

// 						// Reconnect
// 						oledWrite("\n WiFi\n Reconnecting...")
// 						if apMode == true {
// 							ActivateAPMode()

// 						} else {

// 							startWiFiClient()
// 						}
// 					}
// 				}
// 			}

// 			wifiOperation.Unlock()

// 			time.Sleep(30 * time.Second)
// 		}
// 	}()
// }

//

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
