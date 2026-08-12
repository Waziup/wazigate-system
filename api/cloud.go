package api

import (
	"log"
	"net/http"
	"sync"
	"time"
)

//

// The cloud reachability probe is the only periodic request of this service
// that leaves the gateway through wlan0 or the GSM modem. It used to be sent by
// every consumer on its own (the LED and the OLED indicator), which added up to
// about 1700 requests a day. Now a single goroutine probes and everyone else
// reads the cached result.

// The probe URL answers with a bare `204 No Content`, so one probe is a DNS
// lookup, a TCP handshake and two small packets, roughly 1 KB in total.
// It is a variable so that the test can point it at a local server.
var cloudProbeURL = "http://www.waziup.io/generate_204"

const cloudProbeTimeout = 10 * time.Second

// How often the cloud is probed in the background.
const cloudCheckInterval = 5 * time.Minute

// GET /internet may send its own probe to stay accurate, but not more often
// than this, so that a dashboard polling the endpoint can not multiply the
// traffic on a metered uplink.
const cloudProbeMinInterval = 30 * time.Second

//

var cloud struct {
	mutex      sync.Mutex
	accessible bool
	checkedAt  time.Time // Zero value means "not probed yet"
	probing    bool
}

//

// This function starts the single goroutine that probes the cloud.
// Everything else reads the cached result via CloudAccessible.
func RunCloudMonitor() error {

	go func() {
		for {
			cloudRefresh(cloudCheckInterval, false)
			time.Sleep(cloudCheckInterval)
		}
	}()

	log.Printf("[     ] Cloud monitor initialized. Probing every %v.", cloudCheckInterval)
	return nil
}

//

// CloudAccessible returns the last known reachability of the Waziup cloud.
// It never sends a request itself, so it is safe to call from a display loop.
func CloudAccessible(withLogs bool) bool {

	accessible, _ := cloudStatus()
	return accessible
}

//

// This function returns the last known reachability and whether the cloud has
// been probed at all yet, so that callers can tell "not reachable" apart from
// "we do not know yet" while the gateway is starting up.
func cloudStatus() (accessible bool, known bool) {

	cloud.mutex.Lock()
	defer cloud.mutex.Unlock()

	return cloud.accessible, !cloud.checkedAt.IsZero()
}

//

// This function sends a probe if the cached result is older than `maxAge` and
// returns the (possibly refreshed) result. If another goroutine is probing at
// that moment, the cached value is returned right away instead of waiting for
// it, so no caller is ever blocked for longer than one probe.
func cloudRefresh(maxAge time.Duration, withLogs bool) bool {

	cloud.mutex.Lock()

	if cloud.probing || time.Since(cloud.checkedAt) < maxAge {
		accessible := cloud.accessible
		cloud.mutex.Unlock()
		return accessible
	}

	cloud.probing = true
	wasAccessible, wasKnown := cloud.accessible, !cloud.checkedAt.IsZero()

	cloud.mutex.Unlock()

	//

	accessible := cloudProbe(withLogs)

	//

	cloud.mutex.Lock()

	cloud.accessible = accessible
	cloud.checkedAt = time.Now()
	cloud.probing = false

	cloud.mutex.Unlock()

	//

	// State changes are rare and worth a log line, a probe that just confirms
	// what we already knew is not.
	if !wasKnown || accessible != wasAccessible {
		if accessible {
			log.Printf("[     ] Cloud is reachable.")
		} else {
			log.Printf("[WARN ] Cloud is not reachable.")
		}
	}

	return accessible
}

//

// This function performs the actual request. Use `cloudRefresh` instead, it
// keeps the number of requests down.
func cloudProbe(withLogs bool) bool {

	client := http.Client{
		Timeout: cloudProbeTimeout,
	}

	resp, err := client.Get(cloudProbeURL)
	if err != nil {
		if withLogs && DEBUG_MODE {
			log.Printf("[     ] Cloud probe failed: %s", err.Error())
		}
		return false
	}

	resp.Body.Close()
	return true
}

//
