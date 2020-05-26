package main

import (
	// "fmt"
	"log"
	"net"
	"net/http"
	"os"

	// "encoding/json"
	// "strings"
	// "strconv"

	// "os"
	// "os/exec"
	// "path/filepath"
	// "io/ioutil"

	"github.com/Waziup/wazigate-system/api"
	routing "github.com/julienschmidt/httprouter"
)

var router = routing.New()

// Please do not change this line
const sockAddr = "/root/app/proxy.sock"

func init() {

	router.GET("/", api.HomeLink)

	// router.GET( "/ui/", api.UI)
	router.GET("/ui/*file_path", api.UI)

	router.GET("/docs/", api.APIDocs)
	router.GET("/docs/:file_path", api.APIDocs)

	router.GET("/docker", api.DockerStatus)
	router.GET("/docker/:cId", api.DockerStatusById)
	router.POST("/docker/:cId/:action", api.DockerAction)
	router.PUT("/docker/:cId/:action", api.DockerAction)
	router.GET("/docker/:cId/logs", api.DockerLogs)
	router.GET("/docker/:cId/logs/:tail", api.DockerLogs)

	router.GET("/time", api.GetTime)
	router.GET("/timezones", api.GetTimeZones)
	router.GET("/timezone/auto", api.GetTimeZoneAuto) // based on IP address
	router.GET("/timezone", api.GetTimeZone)
	router.PUT("/timezone", api.SetTimeZone)
	router.POST("/timezone", api.SetTimeZone)

	router.GET("/usage", api.ResourceUsage)
	router.GET("/blackout", api.BlackoutEnabled)

	router.GET("/conf", api.GetSystemConf)
	router.POST("/conf", api.SetSystemConf)
	router.PUT("/conf", api.SetSystemConf)

	router.GET("/net", api.GetNetInfo)
	router.GET("/gwid", api.GetGWID)

	router.GET("/net/wifi", api.GetNetWiFi)
	router.POST("/net/wifi", api.SetNetWiFi)
	router.PUT("/net/wifi", api.SetNetWiFi)

	router.GET("/internet", api.InternetAccessible)

	router.GET("/net/wifi/scanning", api.NetWiFiScan)
	router.GET("/net/wifi/scan", api.NetWiFiScan)

	router.GET("/net/wifi/ap", api.GetNetAP)
	router.POST("/net/wifi/ap", api.SetNetAP)
	router.PUT("/net/wifi/ap", api.SetNetAP)

	router.POST("/net/wifi/mode/ap", api.SetNetAPMode)
	router.PUT("/net/wifi/mode/ap", api.SetNetAPMode)

	router.GET("/update", api.SystemUpdate)
	router.POST("/update", api.SystemUpdate)
	router.PUT("/update", api.SystemUpdate)
	router.GET("/update/status", api.SystemUpdateStatus)
	router.GET("/version", api.FirmwareVersion)

	router.POST("/shutdown", api.SystemShutdown)
	router.PUT("/shutdown", api.SystemShutdown)
	router.POST("/reboot", api.SystemReboot)
	router.PUT("/reboot", api.SystemReboot)

	router.POST("/oled", api.OledWriteMessage)
	router.PUT("/oled", api.OledWriteMessage)
}

/*-------------------------*/

// ListenAndServeHTTP serves the APIs and the ui
func ListenAndServeHTTP() {

	log.Printf("Initializing...")

	if err := os.RemoveAll(sockAddr); err != nil {
		log.Fatal(err)
	}

	server := http.Server{
		Handler: router,
	}

	l, e := net.Listen("unix", sockAddr)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	log.Printf("Serving... on socket: [%v]", sockAddr)
	server.Serve(l)

	// addr := os.Getenv( "WAZIGATE_SYSTEM_ADDR")
	// if addr == "" {
	// 	addr = ":5000"
	// }

	// if( api.DEBUG_MODE){
	// 	log.Printf( "[Info  ] Serving on %s", addr)
	// }

	// log.Fatal( http.ListenAndServe( addr, router))
}

/*-------------------------*/
