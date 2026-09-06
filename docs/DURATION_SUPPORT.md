# Implementation Plan: Duration Type and Backend Compat Extensions

## Context

The `webtyp/time` package provides isomorphic time utilities but is missing the `Duration`
type and related constants/helpers that backend libraries (e.g., `webtyp/agent`) need as
a replacement for stdlib `time`. Currently these use cases require falling back to `import "time"`:

| Stdlib usage | Missing from webtyp/time |
|-------------|---------------------------|
| `time.Duration` type | No `Duration` type |
| `time.Second`, `time.Minute`, `time.Millisecond` | No time-unit constants |
| `time.Now().Unix()` (seconds) | `Now()` returns nanoseconds only |
| `time.Since(start).Milliseconds()` | No `Since()` or `Milliseconds()` |
| `http.Client{Timeout: 30 * time.Second}` | Can't set stdlib Duration fields |

---

## Design: What Goes Where

The existing architecture already separates code correctly:

| File | Purpose |
|------|---------|
| `api.go` | Shared public API — delegates to `provider` or pure logic |
| `shared.go` | Shared pure helpers (no `provider` needed) |
| `backStlib.go` | Backend-specific implementation |
| `frontWasm.go` | WASM-specific implementation |

**Rule:** Only code that MUST differ per platform goes in build-tag files.

| Addition | File | Reason |
|----------|------|--------|
| `type Duration` + constants | `backStlib.go` AND `frontWasm.go` | Type IS different: `= stdlib.Duration` vs `int64` |
| `UnixSec() int64` | `api.go` | `Now() / 1e9` — uses existing `Now()`, no platform code |
| `Since(int64) int64` | `api.go` | `Now() - nanoStart` — uses existing `Now()`, no platform code |
| `Milliseconds(int64) int64` | `api.go` | Pure math, no platform dependency |

> `Duration` and constants **cannot** be shared: the type alias (`= stdlib.Duration`) is
> only valid on backend. On WASM it would import stdlib `time`, which is unavailable in TinyGo.

---

## Changes: `api.go` (shared — append)

```go
// UnixSec returns the current UTC time as Unix seconds (integer).
// Equivalent to time.Now().Unix() in stdlib.
// Uses the existing Now() (nanoseconds) and divides by 1e9 — no platform code needed.
func UnixSec() int64 { return Now() / 1_000_000_000 }

// Since returns the nanoseconds elapsed since nanoStart (a UnixNano timestamp).
// Equivalent to time.Since(start).Nanoseconds() in stdlib.
//
//	start := twtime.Now()
//	// ... work ...
//	elapsed := twtime.Since(start) // nanoseconds
func Since(nanoStart int64) int64 { return Now() - nanoStart }

// Milliseconds converts a nanosecond value to milliseconds (integer truncation).
// Equivalent to time.Duration.Milliseconds() when working with int64 timestamps.
//
//	ms := twtime.Milliseconds(twtime.Since(start))
func Milliseconds(ns int64) int64 { return ns / 1_000_000 }
```

---

## Changes: `backStlib.go` (backend — append)

```go
// Duration is a type alias for stdlib time.Duration on backend (!wasm) builds.
// Being a type alias (=) means Duration values are interchangeable with time.Duration
// everywhere — http.Client.Timeout, context.WithTimeout, etc. accept them directly.
//
// On WASM builds, Duration is an independent int64-based type (nanoseconds).
type Duration = time.Duration

// Time unit constants — identical to stdlib on backend (Duration IS time.Duration).
const (
	Nanosecond  Duration = 1
	Microsecond Duration = 1000 * Nanosecond
	Millisecond Duration = 1000 * Microsecond
	Second      Duration = 1000 * Millisecond
	Minute      Duration = 60 * Second
	Hour        Duration = 60 * Minute
)
```

> `backStlib.go` already imports `"time"` for its `timeServer` implementation. No new import needed.

---

## Changes: `frontWasm.go` (WASM — append)

```go
// Duration represents a time duration in nanoseconds on WASM builds.
// On backend (!wasm) builds, Duration is a type alias for stdlib time.Duration.
//
// Values are nanoseconds — identical numeric scale to stdlib time.Duration.
type Duration int64

// Time unit constants — same numeric values as stdlib time constants.
const (
	Nanosecond  Duration = 1
	Microsecond Duration = 1000 * Nanosecond
	Millisecond Duration = 1000 * Microsecond
	Second      Duration = 1000 * Millisecond
	Minute      Duration = 60 * Second
	Hour        Duration = 60 * Minute
)
```

---

## Test Additions: `data_test.go` (shared test runner)

Add a `DurationShared(t *testing.T)` function called from both `backStlib_test.go` and `frontWasm_test.go`:

```go
func DurationShared(t *testing.T) {
    t.Run("Constants", func(t *testing.T) {
        if Second != 1_000_000_000 {
            t.Errorf("Second: expected 1e9 ns, got %d", Second)
        }
        if Minute != 60*Second {
            t.Errorf("Minute: expected 60*Second, got %d", Minute)
        }
        if Millisecond != 1_000_000 {
            t.Errorf("Millisecond: expected 1e6 ns, got %d", Millisecond)
        }
    })

    t.Run("UnixSec", func(t *testing.T) {
        before := Now() / 1_000_000_000
        got := UnixSec()
        after := Now() / 1_000_000_000
        if got < before || got > after {
            t.Errorf("UnixSec %d outside expected range [%d, %d]", got, before, after)
        }
    })

    t.Run("Since", func(t *testing.T) {
        start := Now()
        elapsed := Since(start)
        if elapsed < 0 {
            t.Errorf("Since returned negative value: %d", elapsed)
        }
    })

    t.Run("Milliseconds", func(t *testing.T) {
        if Milliseconds(1_500_000) != 1 {
            t.Error("expected 1 ms from 1.5M ns (truncation)")
        }
        if Milliseconds(int64(Second)) != 1000 {
            t.Error("expected 1000 ms from 1 Second")
        }
    })
}
```

### Backend-only type alias test (`backStlib_test.go`)

```go
func TestDuration_IsStdlibAlias(t *testing.T) {
    // Compile-time proof: Duration is assignable to time.Duration
    var _ time.Duration = Second
    var _ time.Duration = 30 * Second
}
```

---

## Verification

```bash
cd /home/cesar/Dev/Pkg/webtyp/time && gotest .
```

All existing tests must still pass. Then publish:

```bash
cd /home/cesar/Dev/Pkg/webtyp/time && gopush
```

---

## Usage After This Plan (webtyp/agent example)

```go
import twtime "webtyp.com/time"

// Duration for http.Client — works because Duration = time.Duration on backend (type alias)
client := &http.Client{Timeout: 30 * twtime.Second}

// Nanosecond value for twctx.WithTimeout
ctx, cancel := twctx.WithTimeout(parent, int64(30*twtime.Second))

// Timestamp in seconds
msg.CreatedAt = twtime.UnixSec()

// Elapsed milliseconds
start := twtime.Now()
// ... work ...
durationMS := twtime.Milliseconds(twtime.Since(start))
```
