package time

import "webtyp.com/date"

// Unix is an instant in whole seconds since 1970-01-01 UTC: the unit of dates and deadlines (an expiry, the day of a
// booking, the end of a session, the exp of a JWT). Every calendar function and every formatter takes a Unix.
type Unix int64

// Nano is an instant in nanoseconds since 1970-01-01 UTC: the unit of the instant an event happened (created_at,
// updated_at), fine enough to order two events of the same second. Convert it to a date with its Unix method.
type Nano int64

// Durations in seconds, typed: NowUnix() + 30*Day compiles; NowNano() + 30*Day does not, because it mixes units.
const (
	Minute Unix = 60
	Hour   Unix = 3600
	Day    Unix = 86400
)

const nanosPerSecond = 1000000000

// Unix truncates an event instant to the second, for comparing it with a date or formatting it.
func (n Nano) Unix() Unix { return Unix(int64(n) / nanosPerSecond) }

// Nano is the first nanosecond of a date's second, for comparing it with an event instant.
func (u Unix) Nano() Nano { return Nano(int64(u) * nanosPerSecond) }

// Timer represents a cancelable timer.
type Timer interface {
	// Stop prevents the timer from firing. Returns true if the timer was active.
	Stop() bool
}

// NowUnix is the current instant in seconds, for dates and deadlines.
func NowUnix() Unix { return Unix(provider.UnixNano() / nanosPerSecond) }

// NowNano is the current instant in nanoseconds, for the instant of an event.
func NowNano() Nano { return Nano(provider.UnixNano()) }

// FormatTime formats an instant as "HH:MM:SS" in the local timezone offset.
func FormatTime(u Unix) string { return provider.FormatTime(u) }

// FormatMinutes formats minutes since midnight as "HH:MM".
func FormatMinutes(minutes int) string {
	return twoDigits(minutes/60) + ":" + twoDigits(minutes%60)
}

// FormatDate formats an instant as "YYYY-MM-DD" in the local timezone offset.
func FormatDate(u Unix) string { return provider.FormatDate(u) }

// FormatDateTime formats an instant as "YYYY-MM-DD HH:MM:SS" in the local timezone offset.
func FormatDateTime(u Unix) string { return provider.FormatDateTime(u) }

// FormatDateTimeShort formats an instant as "YYYY-MM-DD HH:MM" in the local timezone offset.
func FormatDateTimeShort(u Unix) string { return provider.FormatDateTimeShort(u) }

// FormatISO8601 formats an instant as "YYYY-MM-DDTHH:MM:SSZ" (UTC).
func FormatISO8601(u Unix) string { return provider.FormatISO8601(u) }

// FormatCompact formats an instant as "YYYYMMDDHHmmss" (UTC): PDF metadata, file names.
func FormatCompact(u Unix) string { return provider.FormatCompact(u) }

// ParseDate parses "YYYY-MM-DD" into the instant of its midnight UTC.
func ParseDate(dateStr string) (Unix, error) { return provider.ParseDate(dateStr) }

// ParseTime parses "HH:MM" or "HH:MM:SS" into minutes since midnight.
func ParseTime(timeStr string) (int16, error) { return provider.ParseTime(timeStr) }

// ParseDateTime parses a date ("YYYY-MM-DD") and a time ("HH:MM" or "HH:MM:SS") as UTC.
func ParseDateTime(dateStr, timeStr string) (Unix, error) {
	return provider.ParseDateTime(dateStr, timeStr)
}

// IsToday reports whether the instant falls on today's date in the local timezone offset.
func IsToday(u Unix) bool { return provider.IsToday(u) }

// IsPast reports whether the instant is before now.
func IsPast(u Unix) bool { return u < NowUnix() }

// IsFuture reports whether the instant is after now.
func IsFuture(u Unix) bool { return u > NowUnix() }

// DaysBetween is the number of full days from a to b (negative when b is before a).
func DaysBetween(a, b Unix) int { return int((b - a) / Day) }

// Weekday returns the day of the week (0=Sunday … 6=Saturday) of an instant, in UTC. The rule is date.Weekday's: this
// function only turns the instant into its civil date.
func Weekday(u Unix) int {
	y, m, d, _, _, _ := CivilDate(u)
	return date.Weekday(y, m, d)
}

// MidnightUTC returns midnight UTC of the day that contains the instant.
func MidnightUTC(u Unix) Unix {
	days := u / Day
	if u < 0 && u%Day != 0 {
		days--
	}
	return days * Day
}

// Date returns the instant of a UTC date and time. Pure arithmetic, deterministic in WASM and backend.
func Date(year, month, day, hour, min, sec int) Unix {
	days := date.DaysSinceEpoch(year, month, day)
	return Unix(days*86400 + int64(hour*3600+min*60+sec))
}

// CivilDate splits an instant (UTC) into year, month, day, hour, minute and second.
func CivilDate(u Unix) (year, month, day, hour, min, sec int) {
	days := int64(u) / 86400
	rem := int64(u) % 86400
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

// LocalMinutesToUnixUTC converts minutes since midnight in the IANA zone tz, on the date that contains dateUTC, into a
// UTC instant. An unknown zone falls back to UTC.
func LocalMinutesToUnixUTC(dateUTC Unix, localMinutes int, tz string) Unix {
	return provider.LocalMinutesToUnixUTC(dateUTC, localMinutes, tz)
}

// AfterFunc waits for the specified milliseconds then calls f.
func AfterFunc(milliseconds int, f func()) Timer { return provider.AfterFunc(milliseconds, f) }

func twoDigits(n int) string {
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}

// timeProvider is what each platform (standard Go, WASM) supplies.
type timeProvider interface {
	UnixNano() int64
	FormatDate(u Unix) string
	FormatTime(u Unix) string
	FormatDateTime(u Unix) string
	FormatDateTimeShort(u Unix) string
	FormatISO8601(u Unix) string
	FormatCompact(u Unix) string
	ParseDate(dateStr string) (Unix, error)
	ParseTime(timeStr string) (int16, error)
	ParseDateTime(dateStr, timeStr string) (Unix, error)
	IsToday(u Unix) bool
	LocalMinutesToUnixUTC(dateUTC Unix, localMinutes int, tz string) Unix
	AfterFunc(milliseconds int, f func()) Timer
}

var provider timeProvider
