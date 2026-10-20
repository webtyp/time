package time_test

import (
	"testing"

	"webtyp.com/time"
)

// jan15 is 2024-01-15 15:30:45 UTC.
const jan15 time.Unix = 1705332645

func FormatShared(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"FormatDate", time.FormatDate(jan15), "2024-01-15"},
		{"FormatTime", time.FormatTime(jan15), "15:30:45"},
		{"FormatDateTime", time.FormatDateTime(jan15), "2024-01-15 15:30:45"},
		{"FormatDateTimeShort", time.FormatDateTimeShort(jan15), "2024-01-15 15:30"},
		{"FormatISO8601", time.FormatISO8601(jan15), "2024-01-15T15:30:45Z"},
		{"FormatCompact", time.FormatCompact(jan15), "20240115153045"},
		{"FormatDate epoch", time.FormatDate(0), "1970-01-01"},
		{"FormatDateTimeShort epoch", time.FormatDateTimeShort(0), "1970-01-01 00:00"},
		{"FormatCompact epoch", time.FormatCompact(0), "19700101000000"},
		{"FormatMinutes", time.FormatMinutes(510), "08:30"},
		{"FormatMinutes midnight", time.FormatMinutes(0), "00:00"},
		{"FormatMinutes last", time.FormatMinutes(1439), "23:59"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
}

// FormatOffsetShared: the local formatters apply the timezone offset; ISO and compact stay UTC.
func FormatOffsetShared(t *testing.T) {
	initial := time.GetTimeZoneOffset()
	time.SetTimeZoneOffset(-3)
	defer time.SetTimeZoneOffset(initial)

	if got := time.FormatDateTime(jan15); got != "2024-01-15 12:30:45" {
		t.Errorf("FormatDateTime at UTC-3 = %q, want 2024-01-15 12:30:45", got)
	}
	if got := time.FormatISO8601(jan15); got != "2024-01-15T15:30:45Z" {
		t.Errorf("FormatISO8601 must ignore the offset, got %q", got)
	}
}

func ClocksShared(t *testing.T) {
	unix, nano := time.NowUnix(), time.NowNano()
	if unix <= 0 || nano <= 0 {
		t.Fatalf("clocks must be positive: %d %d", unix, nano)
	}
	if diff := nano.Unix() - unix; diff < -2 || diff > 2 {
		t.Errorf("NowNano().Unix() and NowUnix() must agree within 2s, differ by %d", diff)
	}
	if got := time.Nano(1705332645123456789).Unix(); got != jan15 {
		t.Errorf("Nano.Unix truncates to the second: got %d, want %d", got, jan15)
	}
	if got := jan15.Nano(); got != 1705332645000000000 {
		t.Errorf("Unix.Nano = %d, want 1705332645000000000", got)
	}
	if time.Day != 86400 || time.Hour != 3600 || time.Minute != 60 {
		t.Errorf("durations in seconds: Day %d Hour %d Minute %d", time.Day, time.Hour, time.Minute)
	}
}

func ParseDateShared(t *testing.T) {
	got, err := time.ParseDate("2024-01-15")
	if err != nil || got != 1705276800 {
		t.Errorf("ParseDate(2024-01-15) = %d, %v; want 1705276800 (midnight UTC, seconds)", got, err)
	}
	for _, bad := range []string{"invalid", "2024-02-30", "2024-1-1", "20240-1-15"} {
		if _, err := time.ParseDate(bad); err == nil {
			t.Errorf("ParseDate(%q) must fail", bad)
		}
	}
}

func ParseTimeShared(t *testing.T) {
	for _, c := range []struct {
		in   string
		want int16
	}{{"08:30", 510}, {"08:30:45", 510}} {
		if got, err := time.ParseTime(c.in); err != nil || got != c.want {
			t.Errorf("ParseTime(%q) = %d, %v; want %d", c.in, got, err, c.want)
		}
	}
	for _, bad := range []string{"invalid", "25:00", "xx:00", "-1:00", "00:xx", "00:-1", "00:60"} {
		if _, err := time.ParseTime(bad); err == nil {
			t.Errorf("ParseTime(%q) must fail", bad)
		}
	}
}

func ParseDateTimeShared(t *testing.T) {
	for _, tm := range []string{"08:30", "08:30:00"} {
		if got, err := time.ParseDateTime("2024-01-15", tm); err != nil || got != 1705307400 {
			t.Errorf("ParseDateTime(2024-01-15, %s) = %d, %v; want 1705307400", tm, got, err)
		}
	}
	if _, err := time.ParseDateTime("invalid", "08:30"); err == nil {
		t.Error("ParseDateTime(invalid date) must fail")
	}
	if _, err := time.ParseDateTime("2024-01-15", "invalid"); err == nil {
		t.Error("ParseDateTime(invalid time) must fail")
	}
}

func RelativeShared(t *testing.T) {
	now := time.NowUnix()
	if !time.IsToday(now) || time.IsToday(now-time.Day) || time.IsToday(now+time.Day) {
		t.Error("IsToday: now is today, yesterday and tomorrow are not")
	}
	if !time.IsPast(now-time.Minute) || time.IsPast(now+time.Minute) {
		t.Error("IsPast")
	}
	if !time.IsFuture(now+time.Minute) || time.IsFuture(now-time.Minute) {
		t.Error("IsFuture")
	}
	a, b := time.Unix(1705276800), time.Unix(1705881600) // 2024-01-15, 2024-01-22
	if time.DaysBetween(a, b) != 7 || time.DaysBetween(b, a) != -7 || time.DaysBetween(a, a) != 0 {
		t.Errorf("DaysBetween: %d %d %d", time.DaysBetween(a, b), time.DaysBetween(b, a), time.DaysBetween(a, a))
	}
}

func CalendarShared(t *testing.T) {
	weekdays := []struct {
		at   time.Unix
		want int
	}{
		{1609459200, 5}, // 2021-01-01 Friday
		{0, 4},          // 1970-01-01 Thursday
		{1704585600, 0}, // 2024-01-07 Sunday
		{-time.Day, 3},  // 1969-12-31 Wednesday
		{-5 * time.Day, 6},
	}
	for _, c := range weekdays {
		if got := time.Weekday(c.at); got != c.want {
			t.Errorf("Weekday(%d) = %d, want %d", c.at, got, c.want)
		}
	}
	midnights := []struct{ at, want time.Unix }{
		{1609459200 + time.Hour, 1609459200},
		{1609459200, 1609459200},
		{1609545599, 1609459200},
		{-1, -time.Day},
		{-time.Day, -time.Day},
	}
	for _, c := range midnights {
		if got := time.MidnightUTC(c.at); got != c.want {
			t.Errorf("MidnightUTC(%d) = %d, want %d", c.at, got, c.want)
		}
	}
}

// LocalMinutesToUnixUTCShared: the tz argument is authoritative on every backend; the global offset must never leak
// into a conversion that names its zone.
func LocalMinutesToUnixUTCShared(t *testing.T) {
	initialOffset := time.GetTimeZoneOffset()
	time.SetTimeZoneOffset(7)
	defer time.SetTimeZoneOffset(initialOffset)

	cases := []struct {
		name     string
		date     time.Unix
		minutes  int
		tz       string
		expected time.Unix
	}{
		{"UTC", 1609459200, 9 * 60, "UTC", 1609459200 + 9*time.Hour},
		{"New York winter (UTC-5)", 1609459200, 9 * 60, "America/New_York", 1609459200 + 14*time.Hour},
		{"New York summer (UTC-4)", 1625097600, 9 * 60, "America/New_York", 1625097600 + 13*time.Hour},
		{"Santiago summer (UTC-3)", 1799971200, 9 * 60, "America/Santiago", 1799971200 + 12*time.Hour},
		{"Santiago winter (UTC-4)", 1812672000, 9 * 60, "America/Santiago", 1812672000 + 13*time.Hour},
		{"invalid zone falls back to UTC", 1609459200, 9 * 60, "Invalid/Zone", 1609459200 + 9*time.Hour},
	}
	for _, c := range cases {
		if got := time.LocalMinutesToUnixUTC(c.date, c.minutes, c.tz); got != c.expected {
			t.Errorf("%s: got %d, want %d (off by %+ds)", c.name, got, c.expected, got-c.expected)
		}
	}
}
