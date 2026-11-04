# Last plan executed locally — explicit time zones (`time.Zone`), breaking

Wave: `webtyp/docs/TIME_ZONE_MASTER_PLAN.md`, phase T1.

## Ecosystem search

- `grep -rn "FormatDate\|FormatTime\|IsToday\|LocalMinutesToUnixUTC\|TimeZoneOffset" webtyp veltylabs`: 20 files in 16
  repos used the implicit offset; `LocalMinutesToUnixUTC` (explicit, 2 repos) proved the zone-aware path already worked.
- Both platforms already resolved an IANA zone at an instant (`time.LoadLocation`; `Intl` in `frontWasm.go`) → **reused**
  as the single platform primitive `zoneOffset(name, instant)`.
- `appkit.App.Timezone` / `Deps.Timezone` already carry the installation's zone → **reused** as the source (T2, T4).
- No `Zone` type existed anywhere → **new**, in `webtyp/time` (owner of time).

## Design gate

1. **Prior art:** JavaScript Temporal `ZonedDateTime` requires `timeZone` and was created because `Date`'s implicit
   local zone caused exactly these bugs; Go `t.In(loc)`; Java `ZoneId`; RFC 5545 `DTSTART;TZID=`; HL7 FHIR
   `Appointment.start` is an instant (all unverified links). Shared standard: an instant plus an explicit zone;
   followed.
2. **Names:** `LoadZone`, `MachineZone`, `zone.At`, `zone.Day`, `zone.Date`, `zone.Clock`. UTC functions renamed so
   the name says it: `DateUTC`, `CivilUTC`, `WeekdayUTC`, `ParseDateUTC` (a wrong name is renamed in this change).
3. **Ledger:** concepts +1 (`Zone`); ways to get a local date or time 2 → **1**; `FormatDate`, `FormatTime`,
   `FormatDateTime`, `FormatDateTimeShort`, `IsToday`, `LocalMinutesToUnixUTC`, `SetTimeZoneOffset`,
   `GetTimeZoneOffset`, `ParseDateTime` and `offset*.go` **deleted**, no deprecation.
4. **Where:** `webtyp/time`; platform files answer only `zoneOffset` and `machineZoneName`; `zone.go` is shared, pure.
5. **Deletes:** the fixed offset and its silent UTC fallback (an unknown zone is now `ErrUnknownZone`; a zero `Zone`
   panics instead of acting as UTC).

## Done

`zone.go`; providers rewritten; offset files removed; shared tests (`ZoneShared`) pass on the server and in WASM,
including a Santiago winter night booking (2019-07-15 21:30 = 2019-07-16 01:30 UTC) and `Day` ranges. `gotest` green.
