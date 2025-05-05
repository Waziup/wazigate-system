package nm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
		err = ap.Update(gonetworkmanager.ConnectionSettings{
			"connection": map[string]interface{}{
				"id":             accessPointId,
				"interface-name": "wlan0",
				"type":           "802-11-wireless",
			},
			"802-11-wireless": map[string]interface{}{
				"ssid": []byte(ssid),
			},
			"802-11-wireless-security": map[string]interface{}{
				"psk":      psk,
				"key-mgmt": "wpa-psk",
				"pairwise": []string{"ccmp"},
				"group":    []string{"ccmp"},
				"proto":    []string{"wpa"},
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
