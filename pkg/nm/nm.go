package nm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/Wifx/gonetworkmanager"
)

var nm gonetworkmanager.NetworkManager
var settings gonetworkmanager.Settings
var wlan0 gonetworkmanager.Device
var eth0 gonetworkmanager.Device

// var wifi gonetworkmanager.DeviceWireless

var ap gonetworkmanager.Connection

var Version string

const accessPointId = "WAZIGATE-AP"

// const wifiIdPrefix = "wifi-"

func Connect() (err error) {

	nm, err = gonetworkmanager.NewNetworkManager()
	if err != nil {
		return err
	}

	Version, err = nm.GetPropertyVersion()
	if err != nil {
		return err
	}

	log.Println("[     ] Network Manager Version:", Version)

	settings, err = gonetworkmanager.NewSettings()
	if err != nil {
		return err
	}

	//

	wlan0, err = nm.GetDeviceByIpIface("wlan0")
	if err != nil {
		return fmt.Errorf("no wlan0 interface: %w", err)
	}

	eth0, err = nm.GetDeviceByIpIface("eth0")
	if err != nil {
		return fmt.Errorf("no et0 interface: %w", err)
	}

	//

	connections, err := settings.ListConnections()
	if err != nil {
		return err
	}

	for _, conn := range connections {
		settings, err := conn.GetSettings()
		if err != nil {
			return err
		}
		idInterf := settings["connection"]["id"]
		log.Printf("- %s", idInterf)
		if idInterf != nil {
			if id, ok := idInterf.(string); ok {
				if id == accessPointId {
					ap = conn
					break
				}
			}
		}
	}

	if ap == nil {
		log.Println("[WARN ] The Network Manager Access Point connection has not been found.")
	}
	return nil
}

var errNoHotspot = errors.New("the access point connection is not available")

func Hotspot(ssid string, psk string) (err error) {
	if ap == nil {
		return errNoHotspot
	}
	if ssid != "" {
		if len(psk) < 8 {
			return fmt.Errorf("psk must be at least 8 characters")
		}
		if len(psk) > 63 {
			return fmt.Errorf("psk must be at most 63 characters")
		}
		if len(ssid) > 32 {
			return fmt.Errorf("ssid must be at most 32 characters")
		}

		// cmd := exec.Command("nmcli", "dev", "wifi", "hotspot", "ifname", "wlan0", "ssid", ssid, "password", psk, "con-name", accessPointId)
		// err := cmd.Run()
		err = ap.Update(gonetworkmanager.ConnectionSettings{
			"connection": map[string]interface{}{
				"id":             accessPointId,
				"interface-name": "wlan0",
				"type":           "802-11-wireless",
			},
			"802-11-wireless": map[string]interface{}{
				"ssid": []byte(ssid),
				"mode": "ap",
			},
			"802-11-wireless-security": map[string]interface{}{
				"psk":      psk,
				"key-mgmt": "wpa-psk",
				"pairwise": []string{"ccmp"},
				"group":    []string{"ccmp"},
				"proto":    []string{"rsn"},
			},
			"ipv4": map[string]interface{}{
				"method": "shared",
			},
			"ipv6": map[string]interface{}{
				"method": "ignore",
			},
		})
		if err != nil {
			return err
		}
	}

	log.Printf("[     ] Wifi activating Access Point ...")
	_, err = nm.ActivateConnection(ap, wlan0, nil)
	return err
}

func Wifi(ssid string, psk string, autoconnect bool) (err error) {
	connections, err := settings.ListConnections()
	if err != nil {
		return err
	}

	for _, conn := range connections {
		settings, err := conn.GetSettings()
		if err != nil {
			return err
		}
		if settings["connection"]["type"].(string) == "802-11-wireless" {
			if string(settings["802-11-wireless"]["ssid"].([]byte)) == ssid {
				id := settings["connection"]["id"].(string)
				log.Printf("[     ] Wifi reactivating connection '%s' ...", ssid)
				return wifiReuseConn(conn, id, ssid, psk, autoconnect)
			}
		}
		// idInterf := settings["connection"]["id"]
		// if idInterf != nil {
		// 	if id, ok := idInterf.(string); ok {
		// 		if id == wifiIdPrefix+ssid {
		// 			log.Printf("[     ] Wifi reactivating connection '%s' ...", ssid)
		// 			return wifiReuseConn(conn, ssid, psk, autoconnect)
		// 		}
		// 	}
		// }
	}

	log.Printf("[     ] Wifi adding connection '%s' ...", ssid)
	return wifiNewConn(ssid, psk, autoconnect)
}

func DeleteWifi(ssid string) error {
	connections, err := settings.ListConnections()
	if err != nil {
		return err
	}
	var hasDeleted bool
	for _, conn := range connections {
		settings, err := conn.GetSettings()
		if err != nil {
			return err
		}
		if settings["connection"]["type"] == "802-11-wireless" {
			sid := string(settings["802-11-wireless"]["ssid"].([]byte))
			if sid == ssid {
				if err := conn.Delete(); err != nil {
					return err
				}
				hasDeleted = true
			}
		}
	}
	if !hasDeleted {
		return fmt.Errorf("npo connection with that ssid")
	}
	return nil
}

func wifiReuseConn(conn gonetworkmanager.Connection, id string, ssid string, psk string, autoconnect bool) (err error) {
	settings := gonetworkmanager.ConnectionSettings{
		"connection": map[string]interface{}{
			"id":          id,
			"autoconnect": autoconnect,
		},
		"802-11-wireless": map[string]interface{}{
			"ssid": []byte(ssid),
		},
		"802-11-wireless-security": map[string]interface{}{
			"auth-alg": "open",
			"key-mgmt": "wpa-psk",
		},
	}
	if psk != "" {
		settings["802-11-wireless-security"]["psk"] = psk
	}
	err = conn.Update(settings)
	if err != nil {
		return err
	}
	_, err = nm.ActivateConnection(conn, wlan0, nil)
	return err
}

func wifiNewConn(ssid string, psk string, autoconnect bool) (err error) {
	_, err = nm.AddAndActivateConnection(gonetworkmanager.ConnectionSettings{
		"connection": map[string]interface{}{
			"type":        "802-11-wireless",
			"autoconnect": autoconnect,
		},
		"802-11-wireless": map[string]interface{}{
			"ssid": []byte(ssid),
		},
		"802-11-wireless-security": map[string]interface{}{
			"auth-alg": "open",
			"key-mgmt": "wpa-psk",
			"psk":      psk,
		},
	}, wlan0)
	return err
}

const (
	DeviceStateChangedInterface = gonetworkmanager.DeviceInterface + ".StateChanged"
)

const (
	DevicesObjectPath = gonetworkmanager.NetworkManagerObjectPath + "/Devices"
)

type DeviceState = gonetworkmanager.NmDeviceState

type EventDeviceStateChanged struct {
	Device               string `json:"device"`
	OldState             string `json:"oldState"`
	NewState             string `json:"newState"`
	Reason               string `json:"reason"`
	ActiveConnectionId   string `json:"activeConnectionId,omitempty"`
	ActiveConnectionUUID string `json:"activeConnectionUUID,omitempty"`
}

func Monitor(ctx context.Context, c chan<- interface{}) (err error) {
	defer nm.Unsubscribe()
	signals := nm.Subscribe()
	for {
		select {
		case signal := <-signals:
			if strings.HasPrefix(string(signal.Path), DevicesObjectPath+"/") && signal.Name == DeviceStateChangedInterface {
				device, err := gonetworkmanager.NewDevice(signal.Path)
				if err != nil {
					return err
				}
				ev := new(EventDeviceStateChanged)

				ev.Device, err = device.GetPropertyInterface()
				if err != nil {
					return err
				}

				newState := DeviceState(signal.Body[0].(uint32))
				oldState := DeviceState(signal.Body[1].(uint32))
				reason := DeviceStateReason(signal.Body[2].(uint32))
				ev.NewState = newState.String()
				ev.OldState = oldState.String()
				ev.Reason = reason.String()

				if newState == gonetworkmanager.NmDeviceStatePrepare {
					conn, err := device.GetPropertyActiveConnection()
					if err != nil {
						return err
					}
					ev.ActiveConnectionId, err = conn.GetPropertyID()
					if err != nil {
						return err
					}
					ev.ActiveConnectionUUID, err = conn.GetPropertyUUID()
					if err != nil {
						return err
					}
				} //else if newState == gonetworkmanager.NmDeviceStateActivated {

				// }
				c <- ev
			}
		case <-ctx.Done():
			return context.Canceled
		}
	}
}

type AccessPoint struct {
	Flags      uint32                       `json:"flags"`
	Frequency  uint32                       `json:"freq"`
	HWAddress  string                       `json:"hwAddress"`
	MaxBitrate uint32                       `json:"maxBitrate"`
	Mode       gonetworkmanager.Nm80211Mode `json:"mode"`
	RSNFlags   uint32                       `json:"rsnFlags"`
	SSID       string                       `json:"ssid"`
	Strength   uint8                        `json:"strength"`
	WPAFlags   uint32                       `json:"wpaFlags"`
}

func ScanWifi() ([]AccessPoint, error) {
	if wlan0 == nil {
		return nil, fmt.Errorf("wlan0 unavailable")
	}
	wifi, err := gonetworkmanager.NewDeviceWireless(wlan0.GetPath())
	if err != nil {
		return nil, err
	}

	// lastScan, err := wifi.GetPropertyLastScan()
	// if err != nil {
	// 	return nil, err
	// }

	// if err := wifi.RequestScan(); err != nil {
	// 	return nil, err
	// }

	// for i := 0; i < 20; i++ {
	// 	currentScan, err := wifi.GetPropertyLastScan()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if currentScan != lastScan {
	// 		break
	// 	}
	// 	time.Sleep(100 * time.Millisecond)
	// }

	accessPoints, err := wifi.GetAccessPoints()
	if err != nil {
		return nil, err
	}
	aps := make([]AccessPoint, len(accessPoints))
	for i, accessPoint := range accessPoints {
		aps[i].Flags, _ = accessPoint.GetPropertyFlags()
		aps[i].Frequency, _ = accessPoint.GetPropertyFrequency()
		aps[i].HWAddress, _ = accessPoint.GetPropertyHWAddress()
		aps[i].MaxBitrate, _ = accessPoint.GetPropertyMaxBitrate()
		aps[i].Mode, _ = accessPoint.GetPropertyMode()
		aps[i].RSNFlags, _ = accessPoint.GetPropertyRSNFlags()
		aps[i].SSID, _ = accessPoint.GetPropertySSID()
		aps[i].Strength, _ = accessPoint.GetPropertyStrength()
		aps[i].WPAFlags, _ = accessPoint.GetPropertyWPAFlags()
	}

	return aps, nil
}

type ConnectionSettings map[string]map[string]interface{}

func Connections() (s []ConnectionSettings, err error) {
	conns, err := settings.ListConnections()
	if err != nil {
		return nil, err
	}
	s = make([]ConnectionSettings, len(conns))
	for i, conn := range conns {
		settings, err := conn.GetSettings()
		if err != nil {
			return nil, err
		}
		s[i] = ConnectionSettings(settings)
	}
	return s, nil
}

var errNoDevice = errors.New("no device with that interface name")

func Device(name string) (json.RawMessage, error) {

	devices, err := nm.GetPropertyAllDevices()
	if err != nil {
		return nil, err
	}

	for _, device := range devices {
		interf, err := device.GetPropertyInterface()
		if err != nil {
			return nil, err
		}
		if interf == name {
			jsonData, err := device.MarshalJSON()
			if err != nil {
				return nil, err
			}

			activeConn, err := device.GetPropertyActiveConnection()
			if err == nil && activeConn != nil {
				conn, err := activeConn.GetPropertyConnection()
				if err != nil {
					return nil, err
				}
				settings, err := conn.GetSettings()
				if err != nil {
					return nil, err
				}
				idInterf := settings["connection"]["id"]
				uuidInterf := settings["connection"]["uuid"]
				if idInterf != nil && uuidInterf != nil {
					var jsonMap map[string]interface{}
					json.Unmarshal(jsonData, &jsonMap)
					jsonMap["ActiveConnectionId"] = idInterf
					jsonMap["ActiveConnectionUUID"] = uuidInterf
					jsonData, _ = json.Marshal(jsonMap)
				}
			}
			return jsonData, nil
		}
	}
	return nil, errNoDevice
}

func Devices() (map[string]json.RawMessage, error) {
	wlan0, err := Device("wlan0")
	if err != nil {
		return nil, err
	}
	eth0, err := Device("eth0")
	if err != nil {
		return nil, err
	}
	return map[string]json.RawMessage{
		"wlan0": wlan0,
		"eth0":  eth0,
	}, nil
}
//===========================VPN functions=======================
func CheckVPNStatus()(bool, gonetworkmanager.NmVpnConnectionState,string , error){
	nm, err := gonetworkmanager.NewNetworkManager()
	if err!=nil {
		return false, 0, "", fmt.Errorf("failed connecting to Network manager: %v", err)
	}

	gatewayID, err := getGatewayID()
	if err != nil {
		return false,0,"",fmt.Errorf("failed to get gateway ID: %v", err)
	}
	connected, activeConn, err := isVPNConnected(nm,gatewayID)
	if err != nil || !connected {
		return false,0,"", err
	}

	vpnConn, err := gonetworkmanager.NewVpnConnection(activeConn.GetPath())
	if( err != nil || vpnConn ==nil) {
		return true, 0, "", fmt.Errorf("failed to create VPN connection object")
	}

	state, err := vpnConn.GetPropertyVpnState()

	if state == gonetworkmanager.NmVpnConnectionActivated {
		log.Println("VPN is fully connected!")
	}
	if err != nil {
		return true, 0, "", fmt.Errorf("failed to get VPN state: %v", err)
	}

	banner, err := vpnConn.GetPropertyBanner()
	if err != nil {
		banner = ""
	}

	return true, state, banner, nil
}
func EnableDisableVPN(enable bool) (error) {
	nm, err :=gonetworkmanager.NewNetworkManager()
	if err != nil {
		return fmt.Errorf("failed to connect to NetworkManager: %v", err)
	}
	version, err := nm.GetPropertyVersion()
	if err != nil {
		return err
	}
	log.Println("[     ] Network Manager Version:", version)
	gatewayID, err := getGatewayID()

	if err != nil {
		return fmt.Errorf("failed to get gateway ID: %v", err)
	}
	connected, activeConn, err := isVPNConnected(nm,gatewayID)
	if err != nil || !connected {
		return fmt.Errorf("error checking VPN status: %v", err)
	}

	if !enable {
		if !connected {
			log.Println("VPN is not connected")
			return nil
		}
		return disconnectVPN(nm, activeConn)
	}
	if connected {
		log.Println(" VPN is already active")
		return nil
	}
	conn, exists,err := vpnProfileExists(gatewayID)
	if err != nil {
		return fmt.Errorf("error getting active VPN: %v", err)
	}
	if !exists {
		configFile := gatewayID +".ovpn"
		if err :=downloadVPNConfig(gatewayID, configFile); err !=nil {
			return err
		}
		conn, err = importVPN(configFile)
		if err !=nil {
			return err
		}
	}
	return connectVPN(nm, conn)
}

func connectVPN(nm gonetworkmanager.NetworkManager, conn gonetworkmanager.Connection) error {
	log.Println("Connecting to VPN...")
	activeConn, err := nm.ActivateConnection(conn, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}

	vpnConn,err := gonetworkmanager.NewVpnConnection(activeConn.GetPath())
	if err != nil {
		state, err := vpnConn.GetPropertyVpnState()
		if err == nil {
			log.Printf("VPN State: %v", state)
		}
	}

	log.Println("VPN connected successfully!")
	return nil
}

func importVPN(configFile string) (gonetworkmanager.Connection, error) {
	log.Println("Importing VPN profile...")
	
	cmd := exec.Command("nmcli", "connection", "import", "type", "openvpn", "file", configFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to import VPN: %v - %s", err, output)
	}
	log.Println("VPN profile imported")

	connID := strings.TrimSuffix(configFile, ".ovpn")
	conn, exists, err := vpnProfileExists(connID)
	if err != nil || !exists {
		return nil, fmt.Errorf("failed to find imported connection: %v", err)
	}

	return conn, nil
}

func downloadVPNConfig(gatewayID, outputFile string) error {
	url := fmt.Sprintf("http://3.125.6.177:5000/gateways/%s/vpn", gatewayID)
	
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

func vpnProfileExists(gatewayID string) (gonetworkmanager.Connection, bool, error) {
	settings, err := gonetworkmanager.NewSettings()
	if err !=nil{
		return nil,false,fmt.Errorf("failed to get settings: %v",err)
	}
	connections,err :=settings.ListConnections()
	if err !=nil {
		return nil, false, fmt.Errorf("failed to list connections: %v",err)
	}
	for _, connection := range connections {
		connSettings,err :=connection.GetSettings()
		if err !=nil {
			continue
		}
		if connSettings["connection"]["id"]==gatewayID {
			return connection, true,nil
		}
	}
	return nil, false, nil
}

func disconnectVPN(nm gonetworkmanager.NetworkManager, activeConn gonetworkmanager.ActiveConnection) error {
	log.Println(" Disconnecting VPN...")

	err := nm.DeactivateConnection(activeConn)
	if err !=nil {
		return fmt.Errorf("failed to disconnected %v",err)
	}
	log.Println("VPN disconnected successfully")
	return nil
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

	return "wazigate-" + string(body), nil
	
}
func isVPNConnected(nm gonetworkmanager.NetworkManager, gatewayId string) (bool, gonetworkmanager.ActiveConnection, error){
	activeConnections, err :=nm.GetPropertyActiveConnections()
	if err !=nil {
		return false, nil, fmt.Errorf("failed to get active connections %v",err)
	}
	for _, ac :=range activeConnections {
		isVPN, err :=ac.GetPropertyVPN()
		if err !=nil || isVPN {
			continue
		}
		id, err :=ac.GetPropertyID()
		if err !=nil{
			continue
		}
		if id== gatewayId {
			return true,ac,nil
		}
	}
	return false, nil, nil
}