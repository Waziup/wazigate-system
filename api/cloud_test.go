package api

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// This function points the cloud probe at a local server and returns a counter
// of how many requests actually left for it, plus a cleanup function.
func mockCloudProbe(t *testing.T, status int) (*int32, func()) {

	t.Helper()

	var requests int32

	server := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		atomic.AddInt32(&requests, 1)
		resp.WriteHeader(status)
	}))

	originalURL := cloudProbeURL
	cloudProbeURL = server.URL

	cloud.mutex.Lock()
	cloud.accessible = false
	cloud.checkedAt = time.Time{}
	cloud.probing = false
	cloud.mutex.Unlock()

	return &requests, func() {
		cloudProbeURL = originalURL
		server.Close()
	}
}

//

// The whole point of the cache is that many consumers cause one request, so
// that the LED, the OLED and the API do not each probe on their own.
func TestCloudRefreshProbesOnceWithinMaxAge(t *testing.T) {

	requests, cleanup := mockCloudProbe(t, http.StatusNoContent)
	defer cleanup()

	if _, known := cloudStatus(); known {
		t.Fatal("Expected the cloud state to be unknown before the first probe")
	}

	for i := 0; i < 20; i++ {
		if !cloudRefresh(time.Minute, false) {
			t.Fatalf("Expected the cloud to be reachable on call %d", i)
		}
	}

	// Reading the cached state must never cause a request.
	for i := 0; i < 20; i++ {
		if !CloudAccessible(false) {
			t.Fatalf("Expected the cached state to be reachable on call %d", i)
		}
	}

	if got := atomic.LoadInt32(requests); got != 1 {
		t.Errorf("Expected 1 request for 40 calls, got %d", got)
	}

	if _, known := cloudStatus(); !known {
		t.Error("Expected the cloud state to be known after a probe")
	}
}

//

// Once the cached result is older than maxAge, a new probe has to be sent,
// otherwise a gateway would never notice that it is back online.
func TestCloudRefreshProbesAgainWhenStale(t *testing.T) {

	requests, cleanup := mockCloudProbe(t, http.StatusNoContent)
	defer cleanup()

	cloudRefresh(time.Minute, false)

	// Pretend the cached result is two minutes old.
	cloud.mutex.Lock()
	cloud.checkedAt = time.Now().Add(-2 * time.Minute)
	cloud.mutex.Unlock()

	cloudRefresh(time.Minute, false)

	if got := atomic.LoadInt32(requests); got != 2 {
		t.Errorf("Expected 2 requests, got %d", got)
	}
}

//

// A failing probe has to be remembered just like a successful one, otherwise an
// unreachable cloud would be retried on every single call.
func TestCloudRefreshCachesFailure(t *testing.T) {

	requests, cleanup := mockCloudProbe(t, http.StatusNoContent)
	defer cleanup()

	// An address that nothing listens on, so the probe fails without waiting
	// for the timeout.
	cloudProbeURL = "http://127.0.0.1:1"

	for i := 0; i < 5; i++ {
		if cloudRefresh(time.Minute, false) {
			t.Fatalf("Expected the cloud to be unreachable on call %d", i)
		}
	}

	if got := atomic.LoadInt32(requests); got != 0 {
		t.Errorf("Expected no request to reach the test server, got %d", got)
	}

	accessible, known := cloudStatus()
	if accessible {
		t.Error("Expected the cached state to be unreachable")
	}
	if !known {
		t.Error("Expected a failed probe to count as known state")
	}
}

//

// Concurrent consumers must not turn into concurrent requests.
func TestCloudRefreshConcurrent(t *testing.T) {

	requests, cleanup := mockCloudProbe(t, http.StatusNoContent)
	defer cleanup()

	var wg sync.WaitGroup
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cloudRefresh(time.Minute, false)
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(requests); got != 1 {
		t.Errorf("Expected 1 request for 25 concurrent calls, got %d", got)
	}
}
