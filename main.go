/*
	@author: mojtaba.eskandari@waziup.org
	@author: johann.forster@waziup.org
	@Wazigate System Management
*/
package main

import (
	"log"
	"os"

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
	api.RunLEDManager,
	api.RunTimezoneManager,
	api.RunBlackoutManager,
	api.RunButtonsManager,
	api.RunOLEDManager,
	api.RunFanManager,
	api.GoMonitor,
}

////////////////////////////////////////////////////////////////////////////////

//go:embed package.json
var packageJSON []byte

const packageJSONFile = "/var/lib/waziapp/package.json"

func movePackageJSON() error {
	if err := os.WriteFile(packageJSONFile, packageJSON, 0777); err != nil {
		log.Println("Make sure to run this container with the mapped volume '/var/lib/waziapp'.")
		log.Println("See the Waziapp documentation for more details on running Waziapps.")
		return err
	}
	api.PackageJSON = packageJSON
	return nil
}
