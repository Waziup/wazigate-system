package api

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

	// "github.com/Waziup/wazigate-system/api"
	routing "github.com/julienschmidt/httprouter"
)

// Please do not change this line
const sockAddr = "/root/app/proxy.sock"

func setupRouter() *routing.Router {

	var router = routing.New()

	router.GET("/", HomeLink)

	// router.GET( "/ui/", UI)
	router.GET("/ui/*file_path", UI)

	router.GET("/docs/", APIDocs)
	router.GET("/docs/:file_path", APIDocs)

	router.GET("/docker", DockerStatus)
	router.GET("/docker/:cId", DockerStatusById)
	router.POST("/docker/:cId/:action", DockerAction)
	router.PUT("/docker/:cId/:action", DockerAction)
	router.GET("/docker/:cId/logs", DockerLogs)
	router.GET("/docker/:cId/logs/:tail", DockerLogs)

	router.GET("/time", GetTime)
	router.GET("/timezones", GetTimeZones)
	router.GET("/timezone/auto", GetTimeZoneAuto) // based on IP address
	router.GET("/timezone", GetTimeZone)
	router.PUT("/timezone", SetTimeZone)
	router.POST("/timezone", SetTimeZone)

	router.GET("/usage", ResourceUsage)
	router.GET("/blackout", BlackoutEnabled)

	router.GET("/conf", GetSystemConf)
	router.POST("/conf", SetSystemConf)
	router.PUT("/conf", SetSystemConf)

	router.GET("/net", GetNetInfo)
	router.GET("/gwid", GetGWID)

	router.GET("/net/wifi", GetNetWiFi)
	router.POST("/net/wifi", SetNetWiFi)
	router.PUT("/net/wifi", SetNetWiFi)

	router.GET("/internet", InternetAccessible)

	router.GET("/net/wifi/scanning", NetWiFiScan)
	router.GET("/net/wifi/scan", NetWiFiScan)

	router.GET("/net/wifi/ap", GetNetAP)
	router.POST("/net/wifi/ap", SetNetAP)
	router.PUT("/net/wifi/ap", SetNetAP)

	router.POST("/net/wifi/mode/ap", SetNetAPMode)
	router.PUT("/net/wifi/mode/ap", SetNetAPMode)

	// router.GET("/update", SystemUpdate)
	// router.POST("/update", SystemUpdate)
	// router.PUT("/update", SystemUpdate)
	// router.GET("/update/status", SystemUpdateStatus)
	// router.GET("/version", FirmwareVersion)

	router.POST("/shutdown", SystemShutdown)
	router.PUT("/shutdown", SystemShutdown)
	router.POST("/reboot", SystemReboot)
	router.PUT("/reboot", SystemReboot)

	router.POST("/oled", OledWriteMessage)
	router.PUT("/oled", OledWriteMessage)

	return router
}

/*-------------------------*/

// ListenAndServeHTTP serves the APIs and the ui
func ListenAndServeHTTP() {

	log.Printf("Initializing...")

	router := setupRouter()

	if err := os.RemoveAll(sockAddr); err != nil {
		log.Fatal(err)
	}

	server := http.Server{
		Handler: router,
	}
	defer server.Close()

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

	// if( DEBUG_MODE){
	// 	log.Printf( "[Info  ] Serving on %s", addr)
	// }

	// log.Fatal( http.ListenAndServe( addr, router))
}

/*-------------------------*/
