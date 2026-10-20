//go:build !wasm

package time

import (
	"time"
)

func init() {
	provider = &timeServer{}
}

// timeServer implements timeProvider for standard Go.
type timeServer struct{}

func (ts *timeServer) UnixNano() int64 {
	return time.Now().UTC().UnixNano()
}

func (ts *timeServer) applyOffset(t time.Time) time.Time {
	offset := getOffsetMinutes()
	return t.Add(time.Duration(offset) * time.Minute)
}

func (ts *timeServer) local(u Unix) time.Time { return ts.applyOffset(time.Unix(int64(u), 0).UTC()) }

func (ts *timeServer) FormatDate(u Unix) string     { return ts.local(u).Format("2006-01-02") }
func (ts *timeServer) FormatTime(u Unix) string     { return ts.local(u).Format("15:04:05") }
func (ts *timeServer) FormatDateTime(u Unix) string { return ts.local(u).Format("2006-01-02 15:04:05") }
func (ts *timeServer) FormatDateTimeShort(u Unix) string {
	return ts.local(u).Format("2006-01-02 15:04")
}
func (ts *timeServer) FormatISO8601(u Unix) string {
	return time.Unix(int64(u), 0).UTC().Format(time.RFC3339)
}
func (ts *timeServer) FormatCompact(u Unix) string {
	return time.Unix(int64(u), 0).UTC().Format("20060102150405")
}

func (ts *timeServer) ParseDate(dateStr string) (Unix, error) {
	t, err := time.ParseInLocation("2006-01-02", dateStr, time.UTC)
	if err != nil {
		return 0, err
	}
	return Unix(t.Unix()), nil
}

func (ts *timeServer) ParseTime(timeStr string) (int16, error) {
	return parseTime(timeStr)
}

func (ts *timeServer) ParseDateTime(dateStr, timeStr string) (Unix, error) {
	layout := "2006-01-02 15:04:05"
	if len(timeStr) == 5 {
		layout = "2006-01-02 15:04"
	}
	t, err := time.ParseInLocation(layout, dateStr+" "+timeStr, time.UTC)
	if err != nil {
		return 0, err
	}
	return Unix(t.Unix()), nil
}

func (ts *timeServer) IsToday(u Unix) bool {
	t := ts.local(u)
	now := ts.applyOffset(time.Now().UTC())
	return t.Year() == now.Year() && t.YearDay() == now.YearDay()
}

func (ts *timeServer) LocalMinutesToUnixUTC(dateUTC Unix, localMinutes int, tz string) Unix {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	utcDate := time.Unix(int64(dateUTC), 0).UTC()
	localTime := time.Date(utcDate.Year(), utcDate.Month(), utcDate.Day(), localMinutes/60, localMinutes%60, 0, 0, loc)
	return Unix(localTime.Unix())
}

type timerWrapper struct {
	timer *time.Timer
}

func (tw *timerWrapper) Stop() bool {
	return tw.timer.Stop()
}

func (ts *timeServer) AfterFunc(milliseconds int, f func()) Timer {
	t := time.AfterFunc(time.Duration(milliseconds)*time.Millisecond, f)
	return &timerWrapper{timer: t}
}
