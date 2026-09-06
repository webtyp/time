//go:build !wasm

package time_test

import (
	"sync/atomic"
	"testing"
	stlib "time"

	"webtyp.com/time"
)

func TestAfterFunc(t *testing.T) {
	var executed atomic.Bool

	time.AfterFunc(50, func() {
		executed.Store(true)
	})

	if executed.Load() {
		t.Error("callback executed too early")
	}

	stlib.Sleep(100 * stlib.Millisecond)

	if !executed.Load() {
		t.Error("callback was not executed")
	}
	t.Log("AfterFunc test passed")
}

func TestAfterFunc_Stop(t *testing.T) {
	timer, executed := AfterFuncStopSetup()

	wasActive := timer.Stop()
	if !wasActive {
		t.Error("timer should have been active")
	}

	stlib.Sleep(200 * stlib.Millisecond)
	AfterFuncStopVerify(t, executed)
}
