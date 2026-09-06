# TinyTime - Timer Support Feature

## Issue ID: FEAT_TIMER_SUPPORT

## Problem Statement

### Context

El paquete `crudp` necesita funcionalidades de temporizadores para implementar un sistema de batching (broker) que agrupa paquetes antes de enviarlos. Actualmente el código propuesto usa el paquete `time` de la biblioteca estándar, que **NO está optimizado para TinyGo/WASM**.

### Código actual problemático (crudp/broker.go):

```go
import (
    "sync"
    "time"  // ❌ No optimizado para TinyGo
)

type broker struct {
    timer       *time.Timer  // ❌ No disponible en WASM
    batchWindow int
    // ...
}

func (b *broker) resetTimerLocked() {
    if b.timer != nil {
        b.timer.Stop()
    }
    // ❌ time.AfterFunc no está optimizado para WASM
    b.timer = time.AfterFunc(time.Duration(b.batchWindow)*time.Millisecond, b.flush)
}
```

### Funcionalidades Requeridas

| Funcionalidad | Uso en crudp | Estado en tinytime |
|---------------|--------------|-------------------|
| `Timer` | Temporizador cancelable | ❌ Falta |
| `AfterFunc(ms, f)` | Ejecutar función después de delay | ❌ Falta |

> **Nota:** `UnixMilli()` NO es necesario. Se puede derivar: `UnixNano() / 1_000_000`

---

## Proposed API

### Timer Interface and AfterFunc

**Purpose:** Ejecutar funciones después de un delay, con capacidad de cancelar.

```go
// Timer represents a cancelable timer
type Timer interface {
    // Stop prevents the timer from firing. Returns true if the timer was active.
    Stop() bool
}

// TimeProvider additions
type TimeProvider interface {
    // ... existing methods ...
    
    // AfterFunc waits for the specified milliseconds then calls f.
    // Returns a Timer that can be used to cancel the call.
    // WARNING: In WASM, callback runs in JS event loop - keep it lightweight.
    AfterFunc(milliseconds int, f func()) Timer
}
```

> **IMPORTANT:** No `Sleep()` method. Blocking is not possible in WASM without freezing UI.
> Use callbacks instead of channels for async operations.

**Backend Implementation (backStlib.go):**

```go
// timerWrapper wraps time.Timer to implement Timer interface
type timerWrapper struct {
    timer *time.Timer
}

func (tw *timerWrapper) Stop() bool {
    return tw.timer.Stop()
}

func (ts *timeServer) AfterFunc(milliseconds int, f func()) Timer {
    t := time.AfterFunc(time.Duration(milliseconds)*time.Millisecond, f)
    return &timerWrapper{timer: t}
}
```

**Frontend Implementation (frontWasm.go):**

```go
// WasmTimer implements Timer for WASM using setTimeout
type WasmTimer struct {
    id     int
    active bool
    jsFunc js.Func // Store to release later
}

func (wt *WasmTimer) Stop() bool {
    if !wt.active {
        return false
    }
    js.Global().Call("clearTimeout", wt.id)
    wt.active = false
    wt.jsFunc.Release() // Free memory
    return true
}

func (tc *timeClient) AfterFunc(milliseconds int, f func()) Timer {
    wt := &WasmTimer{
        active: true,
    }
    
    wt.jsFunc = js.FuncOf(func(this js.Value, args []js.Value) any {
        wt.active = false
        wt.jsFunc.Release() // Free memory after execution
        if f != nil {
            f()
        }
        return nil
    })
    
    wt.id = js.Global().Call("setTimeout", wt.jsFunc, milliseconds).Int()
    return wt
}
```

**Tests (backend only - uses time.Sleep for waiting):**

```go
//go:build !wasm

func TestAfterFunc(t *testing.T) {
    tp := NewTimeProvider()
    
    executed := false
    
    tp.AfterFunc(50, func() {
        executed = true
    })
    
    // Should not have executed yet
    if executed {
        t.Error("callback executed too early")
    }
    
    // Wait for execution (backend only)
    time.Sleep(100 * time.Millisecond)
    
    if !executed {
        t.Error("callback was not executed")
    }
}

func TestAfterFunc_Stop(t *testing.T) {
    tp := NewTimeProvider()
    
    executed := false
    
    timer := tp.AfterFunc(100, func() {
        executed = true
    })
    
    // Stop immediately
    wasActive := timer.Stop()
    if !wasActive {
        t.Error("timer should have been active")
    }
    
    // Wait to ensure callback doesn't Fire
    time.Sleep(200 * time.Millisecond)
    
    if executed {
        t.Error("callback should not have executed after Stop()")
    }
}
```

---

## Testing Strategy

### Test File Structure (matches existing pattern)

Following the existing test pattern in tinytime:

```
tinytime/
├── data_test.go           # Shared test functions (no build tags)
├── timer_test.go          # NEW: Shared timer test helpers (no build tags)  
├── backStlib_test.go      # Backend tests (!wasm) - calls shared functions
├── backStlib_timer_test.go # NEW: Backend timer tests (!wasm) - uses time.Sleep
├── frontWasm_test.go      # WASM tests (wasm) - calls shared functions
└── frontWasm_timer_test.go # NEW: WASM timer tests (wasm) - callback-based
```

### Shared Test Helper (timer_test.go - no build tags)

```go
package tinytime_test

import (
	"testing"

	"webtyp.com/time"
)

// AfterFuncStopShared tests that Stop() prevents callback execution.
// The waiting mechanism is platform-specific, so this only sets up the test.
// Returns: timer, pointer to executed flag
func AfterFuncStopSetup(tp tinytime.TimeProvider) (tinytime.Timer, *bool) {
	executed := false
	timer := tp.AfterFunc(100, func() {
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
```

### Backend Timer Test (backStlib_timer_test.go)

```go
//go:build !wasm

package tinytime_test

import (
	"testing"
	"time"

	"webtyp.com/time"
)

func TestAfterFunc(t *testing.T) {
	tp := tinytime.NewTimeProvider()
	executed := false

	tp.AfterFunc(50, func() {
		executed = true
	})

	if executed {
		t.Error("callback executed too early")
	}

	time.Sleep(100 * time.Millisecond)

	if !executed {
		t.Error("callback was not executed")
	}
	t.Log("AfterFunc test passed")
}

func TestAfterFunc_Stop(t *testing.T) {
	tp := tinytime.NewTimeProvider()
	timer, executed := AfterFuncStopSetup(tp)

	wasActive := timer.Stop()
	if !wasActive {
		t.Error("timer should have been active")
	}

	time.Sleep(200 * time.Millisecond)
	AfterFuncStopVerify(t, executed)
}
```

### WASM Timer Test (frontWasm_timer_test.go)

```go
//go:build wasm

package tinytime_test

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
	tp := tinytime.NewTimeProvider()

	timer := tp.AfterFunc(1000, func() {
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
	tp := tinytime.NewTimeProvider()

	timer := tp.AfterFunc(1000, func() {})

	wasActive := timer.Stop()
	if !wasActive {
		t.Error("Stop() should return true for active timer")
	}
	t.Log("AfterFunc Stop returns true - passed")
}

func TestAfterFunc_DoubleStopReturnsFalse(t *testing.T) {
	tp := tinytime.NewTimeProvider()

	timer := tp.AfterFunc(1000, func() {})

	timer.Stop() // First stop
	wasActive := timer.Stop() // Second stop

	if wasActive {
		t.Error("Stop() should return false for already-stopped timer")
	}
	t.Log("AfterFunc double Stop returns false - passed")
}
```

### Run Tests

```bash
# Run all tests (uses existing test.sh with wasmbrowsertest)
./test.sh
```

> **Note:** WASM tests run in a real browser via `wasmbrowsertest`. 
> We cannot use `time.Sleep` to wait for callbacks - it freezes the browser.
> Callback execution is verified only in backend tests.

# WASM  
GOOS=js GOARCH=wasm go test -v ./...
```

After implementing these features, the broker code would change to:

```go
package crudp

import (
    "sync"
    "webtyp.com/time"
)

type queuedPacket struct {
    packet    Packet
    timestamp int64
}

type broker struct {
    mu          sync.Mutex
    queue       []queuedPacket
    batchWindow int
    timer       tinytime.Timer  // ✅ Uses tinytime.Timer
    tp          tinytime.TimeProvider
    codec       Codec
    onFlush     func([]byte)
}

func newBroker(cfg *Config, codec Codec) *broker {
    return &broker{
        queue:       make([]queuedPacket, 0),
        batchWindow: cfg.BatchWindow,
        tp:          tinytime.NewTimeProvider(),  // ✅ Platform-independent
        codec:       codec,
    }
}

func (b *broker) Enqueue(handlerID uint8, action byte, reqID string, data []byte) {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    // ... consolidation logic ...
    
    b.queue = append(b.queue, queuedPacket{
        packet: Packet{
            Action:    action,
            HandlerID: handlerID,
            ReqID:     reqID,
            Data:      [][]byte{data},
        },
        timestamp: b.tp.UnixNano() / 1_000_000,  // ✅ Derive milli from nano
    })
    b.resetTimerLocked()
}

func (b *broker) resetTimerLocked() {
    if b.timer != nil {
        b.timer.Stop()
    }
    b.timer = b.tp.AfterFunc(b.batchWindow, b.flush)  // ✅ Uses tinytime
}
```

---

## Dependencies

- No new dependencies required
- Uses existing `syscall/js` for WASM implementation
- Uses existing `time` package for backend implementation

---

## Breaking Changes

None. All new methods are additions to the existing interface.

---

## Estimated Effort

- **Timer interface**: 30 min  
- **AfterFunc() backend**: 30 min
- **AfterFunc() WASM**: 1-2 hours
- **Tests**: 1 hour
- **Total**: ~3-4 hours

---

## Notes

### WASM Limitations

1. **No blocking/Sleep**: Cannot block in WASM - freezes UI. Use callbacks only.

2. **No channels for waiting**: Channels block goroutines, which blocks the JS event loop in WASM.

3. **Memory management**: `js.Func` must be released with `Release()`. Implementation handles this automatically on timer Fire or stop.

4. **Callbacks must be lightweight**: They run in the JS event loop. Heavy work should be deferred.

### API Design Decision

Using `func()` callback instead of interface:
- ✅ Simpler API
- ✅ Works with closures (can capture state)
- ✅ No need to define separate interface
- ✅ Matches Go's `time.AfterFunc` signature