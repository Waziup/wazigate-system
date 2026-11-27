package api

import (
	"time"

	"periph.io/x/periph/conn/gpio"
)

// Why this exists instead of gpio.PinIO.WaitForEdge:
//
// periph.io v3.6.7's bcm283x driver implements edge detection through the
// kernel's legacy sysfs GPIO interface (/sys/class/gpio), exporting a pin by
// its BCM number. Since Linux 6.x, gpiochip base numbers are allocated
// dynamically from the top of the number space rather than starting at 0, so
// BCM GPIO23 is no longer sysfs pin 23. On the trixie kernel every In() call
// that asks for an edge therefore fails with
//
//	failed to access power_in_pin bcm283x-gpio (GPIO23): pin 23 is not
//	exported by sysfs
//
// and, since main() treated a setup error as fatal, that took the whole
// container into a restart loop.
//
// Reading a pin needs none of that: periph reads the memory-mapped BCM
// registers through /dev/mem, which the kernel change does not affect. Polling
// is immune both to gpiochip renumbering and to the eventual removal of the
// sysfs GPIO ABI.

// gpioPollInterval is how often watchEdge samples a pin. The buttons it serves
// require a hold of whole seconds before anything happens, and a blackout
// signal stays asserted, so 50 ms is far quicker than needed while costing an
// unmeasurable amount of CPU across the three pins involved.
const gpioPollInterval = 50 * time.Millisecond

// watchEdge calls onEdge every time pin transitions to want, and returns once
// onEdge returns false. It blocks, so run it in its own goroutine exactly as
// the WaitForEdge loops it replaces did.
func watchEdge(pin gpio.PinIO, want gpio.Level, onEdge func() bool) {
	previous := pin.Read()
	for {
		time.Sleep(gpioPollInterval)
		current := pin.Read()
		if current == previous {
			continue
		}
		previous = current
		if current != want {
			continue
		}
		if !onEdge() {
			return
		}
	}
}
