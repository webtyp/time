---
PLAN: "feat: DayBounds — the neutral value a day-window constraint crosses a module boundary as"
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **GATE** for `veltylabs/appointment_booking` and `veltylabs/business_calendar`.
> Orchestrator:
> [webtyp/docs/AGENDA_DOMAIN_MASTER_PLAN.md](https://github.com/webtyp/webtyp/blob/main/docs/AGENDA_DOMAIN_MASTER_PLAN.md).

# Plan — `time.DayBounds`

## 1. The problem this solves

Two domain modules must agree on one sentence: **"on this date, only these
minutes are usable."**

- `business_calendar` computes it (an establishment's opening hours, minus
  holidays, minus closures).
- `appointment_booking` consumes it (a professional's blocks may not exceed it).

Today the only way to make them agree is for one to **import** the other. That
would make a *generic* booking module depend on a *specific* institutional
calendar. Verified cost: `appointment_booking/go.mod` has **zero veltylabs
dependencies today** — it is fully independent, and that import would be the
first to break it. Every app wanting appointment booking would then have to
adopt one particular model of "the establishment", including apps that have no
establishment at all (a freelancer's booking page).

The value that crosses is tiny and carries no business meaning. It belongs in a
neutral leaf both modules already import.

## 2. Why `webtyp/time` and not `webtyp/date`

Both were considered. `date` is the wrong leaf:

| | `webtyp/date` | `webtyp/time` |
|---|---|---|
| Vocabulary | `(year, month, day)` arithmetic — leap years, days in month, weekday from a triple, month/weekday **names**, date-key parsing | unix seconds, `Weekday(unixSec)`, `MidnightUTC(unixSec)`, **`LocalMinutesToUnixUTC(dateSec, localMinutes, tz)`** |
| Knows "minutes within a day"? | **no** | **yes — already its currency** |
| Knows a date as unix seconds? | **no** | **yes** |
| Imported by `appointment_booking` today | no | **yes** (`tinytime "webtyp.com/time"`, `service.go:9`) |
| Imported by `business_calendar` today | no | **yes** (`interfaces.go:6`, `module.go:9`) |
| Dependencies | `webtyp/fmt` only | `webtyp/fmt` only |

`time` already converts *minutes within a date* to an instant. A day's usable
window is the same vocabulary, one step up. And because **both modules already
import it under the same alias**, the bridge adds **zero new dependencies to
either side**.

`date` would be a new import for both, to hold a type that mentions neither a
year nor a month name.

## 3. Design gate

New exported symbol. Per skill **api-design**:

### 3.1 Prior art

- **Go's `time.Duration` / `net.IPNet`** — small, meaning-free value types in a
  neutral package that unrelated libraries exchange without importing each
  other. That is exactly the role here.
- **iCalendar's `FREEBUSY` period** — a start/end pair with no reason attached;
  the *reason* a period is busy stays with whoever produced it. `DayBounds`
  copies that split deliberately (§4, why `ClosedBy` does not come along).
- **`java.time.LocalTime` ranges / Joda `Interval`** — the range is a
  time-library concept, not a domain one. Neither carries "why".

Why not follow the "the owner declares the port" rule here: that rule prevents a
consumer inventing a **data type** that already exists upstream. Here nothing
exists upstream — the type is genuinely missing, and putting it in the module
that happens to compute it first is what creates the coupling. The behavioural
**interface** stays with the consumer, which is the Go idiom; only the value
crosses through a leaf.

### 3.2 The novice-name test

`time.DayBounds{Open: true, OpenMin: 480, CloseMin: 1200}` — *"on this day it's
open from minute 480 to minute 1200."* No lookup needed. `Open` reads as a
question already answered; `OpenMin`/`CloseMin` match the `…Min` suffix
`LocalMinutesToUnixUTC` already established in this package.

### 3.3 Complexity ledger

| Row | Δ |
|---|---|
| Concepts the developer must learn | **+1** (`DayBounds`) — genuinely new: nothing named this |
| Module dependencies added | **0** — both consumers already import this package |
| Cross-module dependencies removed | **−1** — `appointment_booking` keeps zero veltylabs deps |
| Lines at the call site | **0** |
| Ways to do the same thing | **−1** — the import-the-neighbour path stops being reachable |
| Exported surface | **+1 type, +3 methods** |

### 3.4 Where it belongs

`webtyp/time` owns instants, dates-as-seconds and minutes-within-a-date. A day's
usable window is that vocabulary. It is **not** a business concern: nothing here
knows what a holiday, a clinic or a booking is.

### 3.5 What it deletes

The `github.com/veltylabs/business_calendar` import that
`appointment_booking`'s plan would otherwise have introduced, and with it the
requirement that every appointment-booking app adopt an institutional calendar.

## 4. Stage 1 — the type

**New file: `daybounds.go`** (this package keeps one concern per file).

```go
package time

// DayBounds is the usable window of one date, in minutes from midnight.
//
// It is the value a day-window constraint crosses a module boundary as: the
// module that computes it (an establishment's calendar) and the module that
// obeys it (a booking schedule) agree through this type instead of importing
// each other.
//
// Open == false means the date is unusable and OpenMin/CloseMin carry no
// meaning. The zero value is therefore "closed", which is the safe default:
// the absence of a computed window never means "all day".
//
// It deliberately carries NO reason. Why a day is closed — a holiday, a local
// closure, a weekly rule — is the producer's business and stays with the
// producer; a consumer that only has to skip the day does not need it, and
// putting it here would drag one domain's vocabulary into every other's.
type DayBounds struct {
	Open     bool
	OpenMin  int
	CloseMin int
}

// Unbounded is the whole day, 00:00 to 24:00. It is what a caller with no
// institutional calendar uses, so "no constraint" is an explicit value rather
// than a nil check repeated at every call site.
func Unbounded() DayBounds { return DayBounds{Open: true, OpenMin: 0, CloseMin: 1440} }

// Contains reports whether a minute-of-day falls inside the window.
// Half-open: OpenMin is inside, CloseMin is not.
func (b DayBounds) Contains(minute int) bool {
	return b.Open && minute >= b.OpenMin && minute < b.CloseMin
}

// Clamp narrows [startMin, endMin) to what b allows, and reports whether
// anything survived. It is the operation a scheduler actually performs: a
// professional's block trimmed to the establishment's opening window.
func (b DayBounds) Clamp(startMin, endMin int) (int, int, bool) {
	if !b.Open {
		return 0, 0, false
	}
	if startMin < b.OpenMin {
		startMin = b.OpenMin
	}
	if endMin > b.CloseMin {
		endMin = b.CloseMin
	}
	if startMin >= endMin {
		return 0, 0, false
	}
	return startMin, endMin, true
}
```

`Clamp` lives here, not in either consumer, because both would otherwise write
it: `appointment_booking` to trim blocks, `scheduleeditor`'s host to bound the
hour options. Two copies of an interval intersection drift, and the stale one is
what someone trusts.

## 5. Stage 2 — tests

**New file: `daybounds_test.go`.**

```go
func TestZeroDayBoundsIsClosed(t *testing.T)            // the safe default
func TestUnboundedCoversTheWholeDay(t *testing.T)
func TestContainsIsHalfOpen(t *testing.T)               // OpenMin in, CloseMin out
func TestClampTrimsBothEdges(t *testing.T)
func TestClampOnClosedDayReportsNothingSurvives(t *testing.T)
func TestClampReportsFalseWhenTheWindowsDoNotOverlap(t *testing.T)
func TestClampLeavesAnAlreadyInsideRangeUntouched(t *testing.T)
```

`TestClampReportsFalseWhenTheWindowsDoNotOverlap` is the one that matters: a
block entirely before opening or after closing must report `false`, not an
inverted range. An implementation that only trims edges without the
`startMin >= endMin` check passes every other test and returns a negative-length
window.

Only `testing` from stdlib, as everywhere in this repo.

## 6. Constraints

- **No stdlib** beyond `testing` in `_test.go`: this package's non-test code
  imports only `webtyp.com/fmt`. `DayBounds` needs neither.
- **No `map`.** This is a struct and three pure methods.
- This package compiles to WASM and TinyGo; keep it allocation-free.
- **No `TODO`, nothing deprecated.** Before closing:
  `grep -rn "TODO\|FIXME\|Deprecated" --include='*.go' .`
- `gotest`, never `go test`.

## 7. Acceptance criteria

| # | Check | Expected |
|---|-------|----------|
| 1 | `gotest ./...` | green, the seven tests in §5 present |
| 2 | `grep -n "type DayBounds" daybounds.go` | present |
| 3 | `grep -rn "holiday\|closure\|business" daybounds.go` | **empty** — no domain vocabulary leaked in |
| 4 | `go list -deps . \| grep -v webtyp.com/fmt \| grep webtyp` | no new dependency |
| 5 | `GOOS=js GOARCH=wasm go build ./...` | compiles |
| 6 | `grep -rn "TODO\|FIXME\|Deprecated" --include='*.go' .` | no hit introduced here |

## 8. Stages

| # | Stage | Files | Done when |
|---|-------|-------|-----------|
| 1 | The type | **new** `daybounds.go` | `DayBounds`, `Unbounded`, `Contains`, `Clamp` |
| 2 | Tests | **new** `daybounds_test.go` | §5 complete, including the no-overlap case |
