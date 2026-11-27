/*
	@author: mojtaba.eskandari@waziup.org
	@author: johann.forster@waziup.org
	@Wazigate System Management
*/
package main

import (
	"log"
	"os"
	"path/filepath"

	_ "embed"

	"github.com/Waziup/wazigate-system/api"
	"github.com/Waziup/wazigate-system/pkg/nm"
)

func main() {
	// Remove date and time from logs
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	log.SetOutput(os.Stdout)

	log.Println("[SETUP] Performing setup, please wait ...")

	for i, fn := range setup {
		if err := fn(); err != nil {
			log.Fatalf("[ERR  ] Setup %d/%d failed: %v", i, len(setup), err)
		}
	}

	log.Println("[SETUP] Setup completed successfully.")

	api.ListenAndServeHTTP()
}

var setup = []func() error{
	movePackageJSON,
	nm.Connect,
	api.Init,
	api.RunCloudMonitor,
	optional("LED manager", api.RunLEDManager),
	api.RunTimezoneManager,
	optional("blackout monitor", api.RunBlackoutManager),
	optional("buttons manager", api.RunButtonsManager),
	optional("OLED manager", api.RunOLEDManager),
	optional("fan manager", api.RunFanManager),
	api.GoMonitor,
}

// optional wraps a setup step that depends on hardware which may simply not be
// fitted, or which a kernel change has moved out of reach. Such a step logs a
// warning and the service carries on without that feature.
//
// Previously every step was fatal, so one unavailable GPIO pin was enough to
// take down networking, the API and the whole container with it: on the trixie
// kernel RunBlackoutManager could not get an edge-triggered input and
// log.Fatalf put the container into a permanent restart loop, with nothing
// else in it ever starting. Losing a button is worth a warning, not an outage.
func optional(name string, fn func() error) func() error {
	return func() error {
		if err := fn(); err != nil {
			log.Printf("[WARN ] %s unavailable, continuing without it: %v", name, err)
		}
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////

//go:embed package.json
var packageJSON []byte

const defaultWaziappFolder = "/var/lib/waziapp"

var waziappFolder = getWaziappFolder()

func getWaziappFolder() string {
	waziappFolder := os.Getenv("WAZIAPP_FOLDER")
	if waziappFolder != "" {
		return waziappFolder
	}
	return defaultWaziappFolder
}

var packageJSONFile = filepath.Join(waziappFolder, "package.json")

func movePackageJSON() error {
	if err := os.WriteFile(packageJSONFile, packageJSON, 0777); err != nil {
		log.Println("Make sure to run this container with the mapped volume '/var/lib/waziapp'.")
		log.Println("See the Waziapp documentation for more details on running Waziapps.")
		return err
	}
	api.PackageJSON = packageJSON
	return nil
}
