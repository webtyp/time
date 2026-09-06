//go:build wasm

package time_test

import (
	"testing"

	"webtyp.com/time"
)

// WASM tests run in real browser via wasmbrowsertest.
// Cannot use time.Sleep - it would freeze the browser UI.
// We can only test synchronous behavior:
// - Timer returns non-nil
// - Stop() returns true for active timer
// - Stop() returns false for already-stopped timer
//
// The actual callback execution is tested only in backend tests.

func TestAfterFunc_ReturnsTimer(t *testing.T) {
	timer := time.AfterFunc(1000, func() {
		// Won't execute during test - we can't wait
	})

	if timer == nil {
		t.Error("AfterFunc should return non-nil Timer")
	}

	// Clean up
	timer.Stop()
	t.Log("AfterFunc returns Timer - passed")
}

func TestAfterFunc_StopReturnsTrue(t *testing.T) {
	timer := time.AfterFunc(1000, func() {})

	wasActive := timer.Stop()
	if !wasActive {
		t.Error("Stop() should return true for active timer")
	}
	t.Log("AfterFunc Stop returns true - passed")
}

func TestAfterFunc_DoubleStopReturnsFalse(t *testing.T) {
	timer := time.AfterFunc(1000, func() {})

	timer.Stop()              // First stop
	wasActive := timer.Stop() // Second stop

	if wasActive {
		t.Error("Stop() should return false for already-stopped timer")
	}
	t.Log("AfterFunc double Stop returns false - passed")
}

func TestAfterFunc_CallbackLogic(t *testing.T) {
	executed := false

	timer := time.AfterFunc(1000, func() {
		executed = true
	})

	// Manually trigger the callback logic
	FireTimer(timer)

	if !executed {
		t.Error("Callback should have been executed by FireTimer")
	}

	// Verify that Stop() returns false after firing
	if timer.Stop() {
		t.Error("Stop() should return false after timer fired")
	}

	// Fire again - should be no-op (safe to call on inactive timer)
	FireTimer(timer)
	if executed == false {
		t.Error("executed flag should still be true")
	}

	t.Log("AfterFunc callback logic - passed")
}

func TestAfterFunc_NilCallback(t *testing.T) {
	// AfterFunc with nil callback should not panic
	timer := time.AfterFunc(1000, nil)

	if timer == nil {
		t.Error("AfterFunc should return non-nil Timer even with nil callback")
	}

	// Manually trigger - should not panic
	FireTimer(timer)

	// Clean up
	timer.Stop()
	t.Log("AfterFunc nil callback - passed")
}
