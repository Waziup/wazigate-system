package api

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"

	routing "github.com/julienschmidt/httprouter"
)

// This is the unix socket file that Waziapps serve on (UIs and APIs).
// The waziapp directory is a mapped volume created by the Wazigate when running this app in a Docker container.
const waziappProxy = "/var/lib/waziapp/proxy.sock"

func setupRouter() *routing.Router {

	var router = routing.New()

	router.GET("/", HomeLink)
	router.GET("/package.json", packageJSON)

	// router.GET("/ui/*file_path", UI)
	router.GET("/ssh/*file_path", SSH)
	router.POST("/ssh/*file_path", SSH)
	router.OPTIONS("/ssh/*file_path", SSH)

	router.GET("/docs/", APIDocs)
	router.GET("/docs/:file_path", APIDocs)

	router.GET("/docker", DockerStatus)
	router.GET("/docker/:cId", DockerStatusById)
	router.POST("/docker/:cId/:action", DockerAction)
	router.PUT("/docker/:cId/:action", DockerAction)
	router.GET("/docker/:cId/logs", DockerLogs)
	router.GET("/docker/:cId/logs/:tail", DockerLogs)

	router.GET("/time", GetTime)
	router.PUT("/time", SetTime)
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

	router.GET("/net/wifi", GetNetWiFi)
	router.POST("/net/wifi", SetNetWiFi)
	router.PUT("/net/wifi", SetNetWiFi)
	router.DELETE("/net/wifi", DeleteNetWiFi)

	router.GET("/internet", InternetAccessible)

	// router.GET("/net/wifi/scanning", NetWiFiScan)
	router.GET("/net/wifi/scan", NetWiFiScan)

	router.GET("/net/conns", NetConns)

	// router.GET("/net/wifi/ap", GetNetAP)
	router.POST("/net/wifi/ap", SetNetAP)
	router.PUT("/net/wifi/ap", SetNetAP)

	router.POST("/net/wifi/mode/ap", SetNetAPMode)
	router.PUT("/net/wifi/mode/ap", SetNetAPMode)

	router.POST("/shutdown", SystemShutdown)
	router.PUT("/shutdown", SystemShutdown)
	router.POST("/reboot", SystemReboot)
	router.PUT("/reboot", SystemReboot)

	router.POST("/oled", OledWriteMessage)
	router.PUT("/oled", OledWriteMessage)

	return router
}

// ListenAndServeHTTP serves the APIs and the UI.
func ListenAndServeHTTP() {

	// Wait for the host to come up before sending any command
	for {
		if hostReady() {
			break
		}
		log.Println("[     ] Waiting for host to become ready ...")
		time.Sleep(5 * time.Second)
	}

	router := setupRouter()

	server := http.Server{
		Handler: router,
	}

	cleanupSocket()
	l, err := net.Listen("unix", waziappProxy) // unix-socket for deploy mode on RPI
	//l, err := net.Listen("tcp", ":8082") // tcp for debug mode on RPI
	if err != nil {
		log.Fatal("Listen error:", err)
	}
	defer cleanupSocket()

	log.Printf("[     ] Listening on 'unix:%s' ...", waziappProxy)
	if err := server.Serve(l); err != http.ErrServerClosed {
		log.Fatal("Server error:", err)
	}
}

func cleanupSocket() {
	if err := os.Remove(waziappProxy); err != nil && !os.IsNotExist(err) {
		log.Fatalln("Can not remove Waziapp proxy.sock:", err)
	}
}
