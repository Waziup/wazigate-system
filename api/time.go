package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	routing "github.com/julienschmidt/httprouter"
)

// GetTime implements GET /time
func GetTime(resp http.ResponseWriter, req *http.Request, params routing.Params) {

	zone, _ := getSystemTimezone()

	loc, _ := time.LoadLocation(zone)
	currentTime := time.Now().In(loc)
	timeStr := currentTime.Format("2006-01-02 15:04:05")

	utcTime := time.Now().UTC()
	utcStr := utcTime.Format("2006-01-02 15:04:05")

	// zoneAuto, _ := getIPBasedTimezone() // it takes some time and the time will not be accurate then

	//

	out := map[string]interface{}{
		"time": timeStr,
		"utc":  utcStr,
		"zone": zone,
		// "zone_auto": zoneAuto,
	}

	outJSON, err := json.Marshal(out)
	if err != nil {
		log.Printf("[ERR  ] %s", err.Error())
	}

	resp.Write([]byte(outJSON))
}

//

func getSystemTimezone() (string, error) {

	cmd := "timedatectl show | grep Timezone"
	stdout, err := execOnHost(cmd)
	if err != nil {
		return "", err // There is an error running the command
	}

	return strings.Replace(stdout, "Timezone=", "", 1), nil
}

//

// GetTimeZone implements GET /timezone
func GetTimeZone(resp http.ResponseWriter, req *http.Request, params routing.Params) {
	zone, _ := getSystemTimezone()

	outJSON, err := json.Marshal(zone)
	if err != nil {
		log.Printf("[ERR  ] %s", err.Error())
	}

	resp.Write([]byte(outJSON))
}

//

// GetTimeZones implements GET /timezones and provide a list of available timezones
func GetTimeZones(resp http.ResponseWriter, req *http.Request, params routing.Params) {

	cmd := "timedatectl list-timezones"
	stdout, err := execOnHost(cmd)
	if err != nil {
		log.Printf("[ERR  ] %s", err.Error())
		http.Error(resp, "[ Error ]: "+err.Error(), http.StatusInternalServerError)
		return
	}
	list := strings.Split(stdout, "\n")

	resp.Write([]byte("["))
	for i, item := range list {
		if i == 0 {
			resp.Write([]byte("\"" + item + "\""))
		} else {
			resp.Write([]byte(",\"" + item + "\""))
		}
	}
	resp.Write([]byte("]"))

}

//

// Implements POST|PUT /timezone
func SetTimeZone(resp http.ResponseWriter, req *http.Request, params routing.Params) {

	var newTimezone string
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&newTimezone)
	if err != nil {
		log.Printf("[ERR  ] %s", err.Error())
		http.Error(resp, "[ Error ]: "+err.Error(), http.StatusBadRequest)
		return
	}

	Config.LocalTimezone = newTimezone

	if newTimezone == "auto" {
		newTimezone, _ = getIPBasedTimezone()
	}

	err = setSystemTimezone(newTimezone)
	if err != nil {
		log.Printf("[ERR  ] %s", err.Error())
		http.Error(resp, "[ Error ]: "+err.Error(), http.StatusInternalServerError)
		return
	}

	saveConfig(Config)

	// resp.Write([]byte(`"OK"`))
}

//

func setSystemTimezone(newTimezone string) error {

	// set in host
	cmd := "sudo timedatectl set-timezone \"" + newTimezone + "\""
	_, err := execOnHost(cmd)
	if err != nil {
		return err
	}

	// set in go
	location, err := time.LoadLocation(newTimezone)
	if err != nil {
		log.Printf("[ERR  ] Error could not set timezone in golang: %s", err)
	} else {
		time.Local = location
	}

	// make api call to edge to update time @runtime
	err = exec.Command("curl", "-X", "POST", "-d", newTimezone, "http://wazigate-edge/sys/timezone").Run()
	if err != nil {
		return err
	}

	return err
}

//

// GetTimeZoneAuto implements GET /timezone/auto
func GetTimeZoneAuto(resp http.ResponseWriter, req *http.Request, params routing.Params) {
	zone, _ := getIPBasedTimezone()

	outJSON, err := json.Marshal(zone)
	if err != nil {
		log.Printf("[ERR  ] %s", err.Error())
	}

	resp.Write([]byte(outJSON))
}

//

// This function attempts to retrieve the timezone of the gateway
// based on its IP address.
// It uses an online service: `http://ip-api.com/json/`
func getIPBasedTimezone() (string, error) {

	url := "http://ip-api.com/json/"

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}

	//

	var GeoData struct {
		Timezone string `json:"timezone"`
	}

	err = json.Unmarshal(body, &GeoData)
	if err != nil {
		return "", err
	}

	//

	return GeoData.Timezone, nil
}

//

// This function initializes the timezone configuration
// and it is called on startup to set the timezone according to the configs
func RunTimezoneManager() error {
	go func() {

		// Wait for the host to come up before sending any command
		for {
			if hostReady() {
				log.Println("[     ] HOST is ready.")
				break
			}
			log.Println("[     ] Waiting for the HOST...")
			time.Sleep(2 * time.Second)
		}

		systemTimezone, _ := getSystemTimezone()
		newTimezone := Config.LocalTimezone

		if newTimezone == "auto" {
			newTimezone, _ = getIPBasedTimezone()
		}

		// Let's set it if different
		if systemTimezone != newTimezone {
			log.Println("[     ] Setting new timezone to [ " + newTimezone + " ]")
			err := setSystemTimezone(newTimezone)
			if err != nil {
				log.Printf("[ERR  ] %s", err.Error())
			}
		}
	}()

	log.Println("[     ] Timezone settings initialized.")
	return nil
}

// Change time manually

func SetSystemTime(newDate string) error {
	cmd := "sudo date -s \"" + newDate + "\""
	_, err := execOnHost(cmd)

	return err
}

// Implements POST|PUT /time
func SetTime(resp http.ResponseWriter, req *http.Request, params routing.Params) {

	var newTime string
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&newTime)
	if err != nil {
		log.Printf("[ERR  ] %s", err.Error())
		http.Error(resp, "[ Error ]: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = SetSystemTime(newTime)
	if err != nil {
		log.Printf("[ERR  ] %s", err.Error())
		http.Error(resp, "[ Error ]: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
