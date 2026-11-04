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
		{"FormatISO8601", time.FormatISO8601(jan15), "2024-01-15T15:30:45Z"},
		{"FormatCompact", time.FormatCompact(jan15), "20240115153045"},
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
	got, err := time.ParseDateUTC("2024-01-15")
	if err != nil || got != 1705276800 {
		t.Errorf("ParseDate(2024-01-15) = %d, %v; want 1705276800 (midnight UTC, seconds)", got, err)
	}
	for _, bad := range []string{"invalid", "2024-02-30", "2024-1-1", "20240-1-15"} {
		if _, err := time.ParseDateUTC(bad); err == nil {
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

func RelativeShared(t *testing.T) {
	now := time.NowUnix()
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
		if got := time.WeekdayUTC(c.at); got != c.want {
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

// ZoneShared: every local date and time comes from a loaded zone, with that zone's offset at that instant.
func ZoneShared(t *testing.T) {
	if _, err := time.LoadZone("Invalid/Zone"); err != time.ErrUnknownZone {
		t.Errorf("LoadZone(invalid) = %v, want ErrUnknownZone (never UTC in silence)", err)
	}
	if _, err := time.LoadZone(""); err != time.ErrUnknownZone {
		t.Errorf("LoadZone(\"\") = %v, want ErrUnknownZone", err)
	}
	scl, err := time.LoadZone("America/Santiago")
	if err != nil {
		t.Fatalf("LoadZone(America/Santiago): %v", err)
	}
	ny, _ := time.LoadZone("America/New_York")
	utc, _ := time.LoadZone("UTC")

	at := []struct {
		name    string
		zone    time.Zone
		date    string
		minutes int
		want    time.Unix
	}{
		{"UTC", utc, "2021-01-01", 9 * 60, 1609459200 + 9*time.Hour},
		{"New York winter (UTC-5)", ny, "2021-01-01", 9 * 60, 1609459200 + 14*time.Hour},
		{"New York summer (UTC-4)", ny, "2021-07-01", 9 * 60, 1625097600 + 13*time.Hour},
		{"Santiago summer (UTC-3)", scl, "2027-01-15", 9 * 60, 1799971200 + 12*time.Hour},
		{"Santiago winter (UTC-4)", scl, "2027-06-11", 9 * 60, 1812672000 + 13*time.Hour},
		{"Santiago winter night: next UTC day", scl, "2019-07-15", 21*60 + 30, 1563240600},
	}
	for _, c := range at {
		got, err := c.zone.At(c.date, c.minutes)
		if err != nil || got != c.want {
			t.Errorf("At %s: got %d %v, want %d (off by %+ds)", c.name, got, err, c.want, got-c.want)
		}
	}

	night := time.Unix(1563240600) // 2019-07-16 01:30 UTC = 2019-07-15 21:30 in Santiago (winter)
	if scl.Date(night) != "2019-07-15" || scl.Clock(night) != "21:30" || scl.DateTime(night) != "2019-07-15 21:30" {
		t.Errorf("Santiago night = %s %s, want 2019-07-15 21:30", scl.Date(night), scl.Clock(night))
	}
	if scl.Minutes(night) != 21*60+30 || scl.Weekday(night) != 1 || time.WeekdayUTC(night) != 2 {
		t.Errorf("Minutes %d Weekday %d (want 1290, Monday=1; UTC is Tuesday=2)", scl.Minutes(night), scl.Weekday(night))
	}
	if scl.Offset(night) != -4*3600 || scl.Offset(1799971200) != -3*3600 {
		t.Errorf("Offset winter %d summer %d, want -14400 -10800", scl.Offset(night), scl.Offset(1799971200))
	}

	from, to, err := scl.Day("2019-07-15")
	if err != nil || from != 1563163200 || to != 1563249600 || !(from <= night && night < to) {
		t.Errorf("Day(2019-07-15) = [%d, %d) %v; want [1563163200, 1563249600) containing the night booking", from, to, err)
	}
	if _, err := scl.At("2019-07-15", 24*60); err == nil {
		t.Error("At with minutes 1440 must fail")
	}
	if _, err := scl.At("2019-02-30", 0); err == nil {
		t.Error("At with an impossible date must fail")
	}

	now := time.NowUnix()
	if !scl.IsToday(now) || scl.IsToday(now-time.Day) || scl.IsToday(now+time.Day) {
		t.Error("IsToday: now is today, yesterday and tomorrow are not")
	}
	if time.MachineZone().Name() == "" {
		t.Error("MachineZone must report a name")
	}

	defer func() {
		if recover() == nil {
			t.Error("a Zone not loaded must panic, never act as UTC")
		}
	}()
	var zero time.Zone
	zero.Date(now)
}
