# time
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

zone, err := time.LoadZone("America/Santiago") // the place's clock; unknown name → ErrUnknownZone
if err != nil {
    panic(err)
}
start, err := zone.At("2019-07-15", 21*60+30)   // a local date and time → the instant (2019-07-16 01:30 UTC)
println(zone.Date(start), zone.Clock(start))    // "2019-07-15" "21:30" — winter, UTC-4, read correctly any day
from, to, err := zone.Day("2019-07-15")          // the day's instants: search with from <= start < to

expiry := time.NowUnix() + 30*time.Day          // a deadline is an instant, in no place
println(time.FormatISO8601(expiry), time.IsPast(expiry))
```

## Time zones — one way, always explicit

There is **no implicit "local time"**: every local date or hour comes from a `Zone` loaded by name. The offset is the
zone's **at that instant** (summer and winter time included), never "the offset of the machine when it started".

- **Server:** the IANA database is embedded (`time/tzdata`), so a Docker image with no zone files and no `TZ` still
  knows `America/Santiago`.
- **Browser:** the browser's `Intl` database, at no cost in the `.wasm` file. A clinic's schedule shows the clinic's
  hour even to a visitor in another country.
- **An app** loads its installation's zone once (`appkit.App.Timezone`). `MachineZone()` — this computer's zone — is
  only for tools (logs, a terminal UI).
- What is UTC says so in its name: `DateUTC`, `CivilUTC`, `WeekdayUTC`, `ParseDateUTC`, `MidnightUTC`, `FormatISO8601`.

---

## API Reference

| Function | Does |
|---|---|
| `NowUnix() Unix`, `NowNano() Nano` | the two clocks |
| `(Nano) Unix() Unix`, `(Unix) Nano() Nano` | explicit conversion between the two units |
| `LoadZone(name) (Zone, error)`, `MachineZone() Zone` | a place's clock / this computer's clock (tools only) |
| `(Zone) At(date, minutes) (Unix, error)` | a local date ("YYYY-MM-DD") and minutes since its midnight → the instant |
| `(Zone) Day(date) (from, to Unix, error)` | the instants of a local day, `[from, to)` |
| `(Zone) Date(Unix)`, `Clock(Unix)`, `DateTime(Unix)` | "YYYY-MM-DD", "HH:MM", both — in the zone |
| `(Zone) Minutes(Unix)`, `Weekday(Unix)`, `IsToday(Unix)`, `Offset(Unix)`, `Name()` | minutes since local midnight; 0 = Sunday; today there; seconds from UTC |
| `FormatISO8601(Unix)`, `FormatCompact(Unix)` | "YYYY-MM-DDTHH:MM:SSZ" and "YYYYMMDDHHmmss", **UTC** |
| `FormatMinutes(int)` | minutes since midnight → "HH:MM" |
| `ParseDateUTC(string) (Unix, error)` | "YYYY-MM-DD" → its midnight UTC (a date as a key) |
| `ParseTime(string) (int16, error)` | "HH:MM[:SS]" → minutes since midnight |
| `IsPast(Unix)`, `IsFuture(Unix)`, `DaysBetween(a, b Unix) int` | relative to now; full days from a to b |
| `WeekdayUTC(Unix) int`, `MidnightUTC(Unix) Unix` | 0 = Sunday, in UTC; start of the UTC day |
| `DateUTC(y, m, d, h, min, s int) Unix`, `CivilUTC(Unix)` | UTC civil date ↔ instant, pure arithmetic |

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

When compiled for WebAssembly (`GOOS=js GOARCH=wasm`), `webtyp.com/time` automatically uses JavaScript's native Date APIs instead of bundling Go's stdlib `time` package.

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
