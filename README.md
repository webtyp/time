# TinyTime
<img src="docs/img/badges.svg">

A minimal, portable time utility for Go and TinyGo with WebAssembly support. Automatically handles timezones and uses JavaScript Date APIs in WASM environments to keep binaries small.

## The two units — read this first

Why, and how `time` and `calendar` divide the work: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

Every instant has its unit in its **type**, so mixing them does not compile:

| Type | Unit | Use it for | Clock |
|---|---|---|---|
| `time.Unix` | seconds since 1970-01-01 UTC | **dates and deadlines**: an expiry, the day of a booking, the end of a session, a JWT `exp` | `time.NowUnix()` |
| `time.Nano` | nanoseconds since 1970-01-01 UTC | **the instant an event happened**: `created_at`, `updated_at`; orders two events of the same second | `time.NowNano()` |

- Durations are typed seconds: `time.Minute`, `time.Hour`, `time.Day` (`time.Unix`). `time.NowUnix() + 30*time.Day`
  compiles; `time.NowNano() + 30*time.Day` does not.
- An event compared with a date or formatted: `n.Unix()`; a date compared with an event: `u.Nano()`. Never divide or
  multiply by 1e9.
- A model field that holds an instant is declared `model.UnixTime()` or `model.NanoTime()`; `ormc` generates it as
  `time.Unix` or `time.Nano`.
- The only arithmetic conversion left is at the border of an external format (a JWT, an OS API), written once in the
  library that speaks that format.

## Quick Start

```go
import "webtyp.com/time"

expiry := time.NowUnix() + 30*time.Day     // a deadline
println(time.FormatDate(expiry))           // "YYYY-MM-DD", local offset

created := time.NowNano()                   // an event
println(time.FormatDateTime(created.Unix())) // "YYYY-MM-DD HH:MM:SS", local offset

day, err := time.ParseDate("2024-01-15")    // midnight UTC, time.Unix
if err != nil {
    panic(err)
}
println(time.Weekday(day), time.IsPast(day))
```

## Timezone Support

The display formatters apply the detected local offset:
- **Standard Go**: `time.Now().Zone()`.
- **WASM**: `Date.getTimezoneOffset()`.

`SetTimeZoneOffset(hours int)` sets it by hand; `GetTimeZoneOffset() int` reads it. `LocalMinutesToUnixUTC` instead
takes an explicit IANA zone and never uses this offset.

---

## API Reference

| Function | Does |
|---|---|
| `NowUnix() Unix`, `NowNano() Nano` | the two clocks |
| `(Nano) Unix() Unix`, `(Unix) Nano() Nano` | explicit conversion between the two units |
| `FormatDate(Unix)`, `FormatTime(Unix)`, `FormatDateTime(Unix)`, `FormatDateTimeShort(Unix)` | "YYYY-MM-DD", "HH:MM:SS", both, "YYYY-MM-DD HH:MM", **local offset** |
| `FormatISO8601(Unix)`, `FormatCompact(Unix)` | "YYYY-MM-DDTHH:MM:SSZ" and "YYYYMMDDHHmmss", **UTC** |
| `FormatMinutes(int)` | minutes since midnight → "HH:MM" |
| `ParseDate(string) (Unix, error)` | "YYYY-MM-DD" → its midnight UTC |
| `ParseTime(string) (int16, error)` | "HH:MM[:SS]" → minutes since midnight |
| `ParseDateTime(date, time string) (Unix, error)` | both, read as UTC |
| `IsToday(Unix)`, `IsPast(Unix)`, `IsFuture(Unix)` | relative to now (today in the local offset) |
| `DaysBetween(a, b Unix) int` | full days from a to b |
| `Weekday(Unix) int`, `MidnightUTC(Unix) Unix` | 0 = Sunday; start of the UTC day |
| `Date(y, m, d, h, min, s int) Unix`, `CivilDate(Unix)` | civil date ↔ instant, pure arithmetic |
| `LocalMinutesToUnixUTC(date Unix, minutes int, tz string) Unix` | a local wall-clock time in an IANA zone → UTC instant |

---

### Timers

#### `AfterFunc(milliseconds int, f func()) Timer`
Waits for the specified milliseconds then calls `f`. Returns a `Timer` that can be used to cancel the call.

**Note:** In WASM environments, the callback runs in the JavaScript event loop. Keep callbacks lightweight to avoid blocking the UI.

```go
// Start a timer
timer := time.AfterFunc(1000, func() {
    println("1 second passed!")
})

// Stop the timer before it fires
timer.Stop()
```

---

## WebAssembly Usage

When compiled for WebAssembly (`GOOS=js GOARCH=wasm`), tinytime automatically uses JavaScript's native Date APIs instead of bundling Go's `time` package.

```bash
# Build for WebAssembly
GOOS=js GOARCH=wasm go build -o app.wasm .
```

## Testing

Always use `gotest` to run tests:
```bash
gotest
```

## Dependencies
- `webtyp.com/fmt`

---
## [Contributing](https://github.com/webtyp/cdvelop/blob/main/CONTRIBUTING.md)

## License

[MIT](LICENSE).
