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

type VPNStatus struct {
	Connected bool   `json:"connected"`
	Name      string `json:"name,omitempty"`
}

// GetVPNStatus implements GET /vpn
func GetVPNStatus(resp http.ResponseWriter, req *http.Request, params routing.Params) {
	connected, name, err := IsVPNConnected()
	if err != nil {
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(resp).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	status := VPNStatus{
		Connected: connected,
		Name:      name,
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)
	json.NewEncoder(resp).Encode(status)
}

type VPNRequest struct {
	Enabled bool `json:"enabled"`
}

type VPNResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// PostVPN implements POST /vpn
func PostVPN(resp http.ResponseWriter, req *http.Request, params routing.Params) {
	var reqBody VPNRequest
	
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&reqBody)
	if err != nil {
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(resp).Encode(VPNResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	// Execute VPN operation in background to avoid blocking
	err = EnableDisableVPN(reqBody.Enabled)

	if err != nil {
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(resp).Encode(VPNResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Prepare response message
	action := "disabled"
	if reqBody.Enabled {
		action = "enabled"
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)
	json.NewEncoder(resp).Encode(VPNResponse{
		Success: true,
		Message: "VPN " + action + " successfully",
	})
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
func EnableDisableVPN(enable bool) error {
	gatewayID, err := getGatewayID()
	if err != nil {
		return fmt.Errorf("failed to get gateway ID: %v", err)
	}

	connected, activeName, err := IsVPNConnected()
	if err != nil {
		return fmt.Errorf("error checking VPN status: %v", err)
	}

	// Handle disconnect
	if !enable {
		if !connected {
			fmt.Println("ℹ️  VPN is not connected")
			return nil
		}
		return disconnectVPN(gatewayID)
	}

	if connected {
		fmt.Printf("⚠️  VPN '%s' is already active\n", activeName)
		return nil
	}

	// Check if VPN profile exists
	exists, err := vpnProfileExists(gatewayID)
	if err != nil {
		return fmt.Errorf("error checking VPN profile: %v", err)
	}

	// Only download and import if profile doesn't exist
	if !exists {
		configFile := gatewayID + ".ovpn"
		if err := downloadVPNConfig(gatewayID, configFile); err != nil {
			fmt.Println("Error:", err)
			return err
		}

		if err := importVPN(configFile); err != nil {
			fmt.Println("Error:", err)
			return err
		}
	} else {
		log.Println("VPN profile already exists, connecting...")
	}


	return connectVPN(gatewayID)

}

// isVPNConnected checks if any VPN is active
func IsVPNConnected() (bool, string, error) {
	out, err := runCommand("nmcli", "-t", "-f", "NAME,TYPE", "connection", "show", "--active")
	if err != nil {
		return false, "", fmt.Errorf("failed to check VPN: %v", err)
	}

	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, ":vpn") {
			name := strings.Split(line, ":")[0]
			return true, name, nil
		}
	}
	return false, "", nil
}

// vpnProfileExists checks if the VPN profile is configured in NetworkManager
func vpnProfileExists(vpnName string) (bool, error) {
	out, err := runCommand("nmcli", "-t", "-f", "NAME", "connection", "show")
	if err != nil {
		return false, fmt.Errorf("failed to list connections: %v", err)
	}

	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == vpnName {
			return true, nil
		}
	}
	return false, nil
}

// disconnectVPN disconnects the active VPN
func disconnectVPN(vpnName string) error {
	if _, err := runCommand("nmcli", "con", "down", vpnName); err != nil {
		return fmt.Errorf("failed to disconnect: %v", err)
	}
	return nil
}

// downloadVPNConfig downloads the .ovpn file from the server
func downloadVPNConfig(gatewayID, outputFile string) error {
	url := fmt.Sprintf("http://3.71.4.83:5000/gateways/%s/vpn", gatewayID)
	
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch config: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
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

// importVPN imports the VPN configuration into NetworkManager
func importVPN(configFile string) error {
	log.Println("Importing VPN profile...")
	if _, err := runCommand("nmcli", "connection", "import", "type", "openvpn", "file", configFile); err != nil {
		return fmt.Errorf("failed to import: %v", err)
	}
	log.Println("VPN profile imported")
	return nil
}

// connectVPN connects to the VPN
func connectVPN(vpnName string) error {
	log.Println("Connecting to VPN...")
	if _, err := runCommand("nmcli", "connection", "up", "id", vpnName); err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	return nil
}

// getGatewayID fetches the gateway ID from the edge service
func getGatewayID() (string, error) {
	resp, err := http.Get("http://wazigate-edge/device/id")
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

	return "wazigate-" + string(body), nil
}

// runCommand executes a shell command and returns output
func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}