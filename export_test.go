//go:build wasm

package time_test

import (
	"webtyp.com/time"
)

// FireTimer triggers the timer callback manually for testing purposes.
// This allows verifying the callback logic without waiting for the browser event loop.
func FireTimer(t time.Timer) {
	if wt, ok := t.(*time.WasmTimer); ok {
		wt.Fire()
	}
}
