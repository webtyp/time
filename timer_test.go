package time_test

import (
	"testing"

	"webtyp.com/time"
)

// AfterFuncStopShared tests that Stop() prevents callback execution.
// Returns: timer, pointer to executed flag
func AfterFuncStopSetup() (time.Timer, *bool) {
	executed := false
	timer := time.AfterFunc(100, func() {
		executed = true
	})
	return timer, &executed
}

// AfterFuncStopVerify checks that the callback was NOT executed after Stop()
func AfterFuncStopVerify(t *testing.T, executed *bool) {
	if *executed {
		t.Error("callback should not have executed after Stop()")
	}
	t.Log("AfterFunc_Stop test passed")
}
