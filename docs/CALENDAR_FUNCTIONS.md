# Calendar Functions

This document describes the calendar-related functions added to `webtyp/time`.

## Functions

### `Weekday(unixSec int64) int`

Returns the day of the week (0=Sunday … 6=Saturday) for a Unix timestamp in seconds (UTC).

- **Implementation**: Pure arithmetic.
- **Behavior**: Identical on both standard Go (backend) and WASM (frontend).
- **Example**: `Weekday(1609459200)` (2021-01-01) returns `5` (Friday).

### `MidnightUTC(unixSec int64) int64`

Returns the Unix timestamp in seconds for midnight UTC of the day that contains the given Unix timestamp in seconds.

- **Implementation**: Pure arithmetic.
- **Behavior**: Identical on both standard Go (backend) and WASM (frontend).
- **Example**: `MidnightUTC(1609462800)` (2021-01-01 01:00:00 UTC) returns `1609459200` (2021-01-01 00:00:00 UTC).

### `LocalMinutesToUnixUTC(dateSec int64, localMinutes int, tz string) int64`

Converts minutes-from-midnight expressed in a local timezone into a UTC Unix timestamp in seconds for the given date.

- **Parameters**:
  - `dateSec`: A Unix timestamp in seconds (UTC) that identifies the target date.
  - `localMinutes`: Minutes elapsed since midnight in the local timezone.
  - `tz`: An IANA timezone name (e.g., "America/New_York").
- **Backend Behavior**: Uses full IANA resolution via the `time` standard library. Falls back to UTC if the timezone is invalid.
- **Frontend (WASM) Behavior**: Uses the global offset set via `SetTimeZoneOffset` as the approximation. The `tz` parameter is ignored.
- **Example**: `LocalMinutesToUnixUTC(1609459200, 540, "America/New_York")` (2021-01-01, 09:00 local) returns `1609509600` (14:00 UTC).
