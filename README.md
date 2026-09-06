# TinyTime
<img src="docs/img/badges.svg">

A minimal, portable time utility for Go and TinyGo with WebAssembly support. Automatically handles timezones and uses JavaScript Date APIs in WASM environments to keep binaries small.

## Quick Start

```go
import "webtyp.com/time"

func main() {
    // Get current Unix timestamp in nanoseconds (UTC)
    nano := time.Now()
    println("Current time:", nano)

    // Format dates and times (Using detected local timezone)
    date := time.FormatDate(nano)
    println("Local Date:", date)

    timeStr := time.FormatTime(nano)
    println("Local Time:", timeStr)

    // Set custom timezone offset manually (e.g. UTC-3)
    time.SetTimeZoneOffset(-180) 
    println("New Local Time:", time.FormatTime(nano))

    // Parse date and time strings (Always UTC)
    parsedNano, err := time.ParseDate("2024-01-15")
    if err != nil {
        panic(err)
    }
    println("Parsed UTC Nano:", parsedNano)

    // Perform date calculations
    isToday := time.IsToday(nano)
    println("Is Today (Local)?", isToday)

    // Compact format (UTC, no separators)
    compact := time.FormatCompact(nano)
    println("Compact:", compact) // "20260402153045"
}
```

## Timezone Support

TinyTime automatically detects the local timezone offset:
- **Standard Go**: Uses `time.Now().Zone()`.
- **WASM**: Uses JavaScript `Date.getTimezoneOffset()`.

### `SetTimeZoneOffset(offsetMinutes int)`
Sets the manual timezone offset in minutes from UTC.

### `GetTimeZoneOffset() int`
Returns the current active timezone offset in minutes.

---

## API Reference

### Display Formatting
**⚠️ IMPORTANT:** All formatting functions below (`FormatDate`, `FormatTime`, `FormatDateTime`, etc.) automatically apply the current timezone offset to display **local time**. If you need raw UTC time, use `FormatISO8601` instead.

#### `FormatDate(value any) string`
Formats a value into a date string: "YYYY-MM-DD".
- **`int64`**: UnixNano timestamp.
- **`string`**: Valid date string (passthrough).

#### `FormatTime(value any) string`
Formats a value into a time string "HH:MM:SS" (or "HH:MM" for `int16`).
- **`int64`**: UnixNano timestamp.
- **`int16`**: Minutes since midnight.
- **`string`**: Valid time string (passthrough).

#### `FormatDateTime(value any) string`
Formats a value into a date-time string: "YYYY-MM-DD HH:MM:SS".

#### `FormatDateTimeShort(value any) string`
Formats a value into a short date-time string: "YYYY-MM-DD HH:MM".

#### `FormatISO8601(nano int64) string`
Formats a UnixNano timestamp into an ISO 8601 string: "YYYY-MM-DDTHH:MM:SSZ".
Unlike other formatting functions, this strictly outputs **UTC time** and ignores any local timezone offsets. Often used for DB records, HL7/FHIR, and REST APIs.

#### `FormatCompact(nano int64) string`
Formats a UnixNano timestamp into a compact string: "YYYYMMDDHHmmss".
Outputs **UTC time**, ignoring timezone offsets. Useful for PDF metadata dates, file naming, and compact timestamps.

---

### Parsing
All parsing functions assume UTC input and return UTC timestamps.

#### `ParseDate(dateStr string) (int64, error)`
Parses a date string ("YYYY-MM-DD") into a UnixNano timestamp at midnight UTC.

#### `ParseTime(timeStr string) (int16, error)`
Parses a time string ("HH:MM" or "HH:MM:SS") into minutes since midnight UTC.

#### `ParseDateTime(dateStr, timeStr string) (int64, error)`
Combines date and time strings into a single UnixNano timestamp (UTC).

---

### Current Time

#### `Now() int64`
Returns the current Unix timestamp in nanoseconds (UTC).

---

### Date Utilities

#### `IsToday(nano int64) bool`
Checks if the given UnixNano timestamp is today based on the local timezone offset.

#### `IsPast(nano int64) bool`
Checks if the given UnixNano timestamp is in the past.

#### `IsFuture(nano int64) bool`
Checks if the given UnixNano timestamp is in the future.

#### `DaysBetween(nano1, nano2 int64) int`
Calculates the number of full days between two UnixNano timestamps.

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
