package time

import "webtyp.com/date"

// Timer represents a cancelable timer.
type Timer interface {
	// Stop prevents the timer from firing. Returns true if the timer was active.
	Stop() bool
}

// Now retrieves the current Unix timestamp in nanoseconds in UTC.
func Now() int64 {
	return provider.UnixNano()
}

// FormatTime formats a value into a time string "HH:MM:SS" applying the timezone offset.
func FormatTime(value any) string {
	return provider.FormatTime(value)
}

// FormatDate formats a value into a date string "YYYY-MM-DD" applying the timezone offset.
func FormatDate(value any) string {
	return provider.FormatDate(value)
}

// FormatDateTime formats a value into a date-time string "YYYY-MM-DD HH:MM:SS" applying the timezone offset.
func FormatDateTime(value any) string {
	return provider.FormatDateTime(value)
}

// FormatDateTimeShort formats a value into a short date-time string "YYYY-MM-DD HH:MM".
func FormatDateTimeShort(value any) string {
	return provider.FormatDateTimeShort(value)
}

// FormatISO8601 formats a UnixNano timestamp into an ISO 8601 string (UTC).
// Format: "YYYY-MM-DDTHH:MM:SSZ"
func FormatISO8601(nano int64) string {
	return provider.FormatISO8601(nano)
}

// FormatCompact formats a UnixNano timestamp into a compact string "YYYYMMDDHHmmss" (UTC).
// Useful for PDF metadata dates, file naming, and compact timestamps.
func FormatCompact(nano int64) string {
	return provider.FormatCompact(nano)
}

// ParseDate parses a date string ("YYYY-MM-DD") into a UnixNano timestamp (UTC).
func ParseDate(dateStr string) (int64, error) {
	return provider.ParseDate(dateStr)
}

// ParseTime parses a time string into minutes since midnight (UTC).
func ParseTime(timeStr string) (int16, error) {
	return provider.ParseTime(timeStr)
}

// ParseDateTime combines date and time strings into a single UnixNano timestamp (UTC).
func ParseDateTime(dateStr, timeStr string) (int64, error) {
	return provider.ParseDateTime(dateStr, timeStr)
}

// IsToday checks if the given UnixNano timestamp is today according to the current timezone offset.
func IsToday(nano int64) bool {
	return provider.IsToday(nano)
}

// IsPast checks if the given UnixNano timestamp is in the past.
func IsPast(nano int64) bool {
	return provider.IsPast(nano)
}

// IsFuture checks if the given UnixNano timestamp is in the future.
func IsFuture(nano int64) bool {
	return provider.IsFuture(nano)
}

// DaysBetween calculates the number of full days between two UnixNano timestamps.
func DaysBetween(nano1, nano2 int64) int {
	return DaysBetweenShared(nano1, nano2)
}

// Weekday returns the day of the week (0=Sunday … 6=Saturday) for a Unix
// timestamp in seconds (UTC). Based on the fact that 1970-01-01 was a Thursday (4).
func Weekday(unixSec int64) int {
	days := MidnightUTC(unixSec) / 86400
	w := int((days + 4) % 7)
	if w < 0 {
		w += 7
	}
	return w
}

// MidnightUTC returns the Unix timestamp in seconds for midnight UTC of the
// day that contains the given Unix timestamp in seconds.
func MidnightUTC(unixSec int64) int64 {
	days := unixSec / 86400
	if unixSec < 0 && unixSec%86400 != 0 {
		days--
	}
	return days * 86400
}

// Date returns the Unix timestamp in seconds for the given UTC date and time.
// Pure arithmetic, zero allocations, deterministic in WASM and backend.
func Date(year, month, day, hour, min, sec int) int64 {
	days := date.DaysSinceEpoch(year, month, day)
	return days*86400 + int64(hour*3600+min*60+sec)
}

// CivilDate converts a Unix timestamp in seconds (UTC) into (year, month, day, hour, min, sec).
// Pure arithmetic, zero allocations, deterministic in WASM and backend.
func CivilDate(unixSec int64) (year, month, day, hour, min, sec int) {
	days := unixSec / 86400
	rem := unixSec % 86400
	if rem < 0 {
		days--
		rem += 86400
	}
	year, month, day = date.FromDaysSinceEpoch(days)
	hour = int(rem / 3600)
	rem %= 3600
	min = int(rem / 60)
	sec = int(rem % 60)
	return
}

// DateKeyToUnix parses a date key "YYYY-MM-DD" and returns the Unix timestamp in seconds
// for midnight UTC. Returns 0 if dateKey is invalid.
func DateKeyToUnix(dateKey string) int64 {
	y, m, d := date.ParseDateKey(dateKey)
	if y == 0 {
		return 0
	}
	return Date(y, m, d, 0, 0, 0)
}

// LocalMinutesToUnixUTC converts minutes-from-midnight expressed in a local
// timezone into a UTC Unix timestamp in seconds for the given date.
// dateSec is a Unix timestamp in seconds (UTC) that identifies the target date.
// localMinutes is minutes elapsed since midnight in the local timezone.
// tz is an IANA timezone name (e.g. "America/New_York"); falls back to UTC if invalid.
func LocalMinutesToUnixUTC(dateSec int64, localMinutes int, tz string) int64 {
	return provider.LocalMinutesToUnixUTC(dateSec, localMinutes, tz)
}

// AfterFunc waits for the specified milliseconds then calls f.
func AfterFunc(milliseconds int, f func()) Timer {
	return provider.AfterFunc(milliseconds, f)
}

// Internal interface for the singleton provider
type timeProvider interface {
	UnixNano() int64
	FormatDate(value any) string
	FormatTime(value any) string
	FormatDateTime(value any) string
	FormatDateTimeShort(value any) string
	FormatISO8601(nano int64) string
	FormatCompact(nano int64) string
	ParseDate(dateStr string) (int64, error)
	ParseTime(timeStr string) (int16, error)
	ParseDateTime(dateStr, timeStr string) (int64, error)
	IsToday(nano int64) bool
	IsPast(nano int64) bool
	IsFuture(nano int64) bool
	LocalMinutesToUnixUTC(dateSec int64, localMinutes int, tz string) int64
	AfterFunc(milliseconds int, f func()) Timer
}

var provider timeProvider
