//go:build !wasm

package time

import (
	"fmt"
	"time"

	. "webtyp.com/fmt"
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

func (ts *timeServer) FormatDate(value any) string {
	switch v := value.(type) {
	case int64:
		t := time.Unix(0, v).UTC()
		return ts.applyOffset(t).Format("2006-01-02")
	case string:
		if _, err := time.Parse("2006-01-02", v); err == nil {
			return v
		}
	}
	return ""
}

func (ts *timeServer) FormatTime(value any) string {
	switch v := value.(type) {
	case int64: // UnixNano
		t := time.Unix(0, v).UTC()
		return ts.applyOffset(t).Format("15:04:05")
	case int16: // Minutes since midnight
		hours := v / 60
		minutes := v % 60
		return fmt.Sprintf("%02d:%02d", hours, minutes)
	case string:
		if nano, err := Convert(v).Int64(); err == nil {
			t := time.Unix(0, nano).UTC()
			return ts.applyOffset(t).Format("15:04:05")
		}
		if Count(v, ":") >= 1 {
			return v
		}
	}
	return ""
}

func (ts *timeServer) FormatDateTime(value any) string {
	switch v := value.(type) {
	case int64:
		t := time.Unix(0, v).UTC()
		return ts.applyOffset(t).Format("2006-01-02 15:04:05")
	case string:
		if _, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			return v
		}
	}
	return ""
}

func (ts *timeServer) FormatDateTimeShort(value any) string {
	switch v := value.(type) {
	case int64:
		t := time.Unix(0, v).UTC()
		return ts.applyOffset(t).Format("2006-01-02 15:04")
	case string:
		if _, err := time.Parse("2006-01-02 15:04", v); err == nil {
			return v
		}
	}
	return ""
}

func (ts *timeServer) FormatISO8601(nano int64) string {
	t := time.Unix(0, nano).UTC()
	return t.Format(time.RFC3339)
}

func (ts *timeServer) FormatCompact(nano int64) string {
	t := time.Unix(0, nano).UTC()
	return t.Format("20060102150405")
}

func (ts *timeServer) ParseDate(dateStr string) (int64, error) {
	t, err := time.ParseInLocation("2006-01-02", dateStr, time.UTC)
	if err != nil {
		return 0, err
	}
	return t.UnixNano(), nil
}

func (ts *timeServer) ParseTime(timeStr string) (int16, error) {
	return parseTime(timeStr)
}

func (ts *timeServer) ParseDateTime(dateStr, timeStr string) (int64, error) {
	layout := "2006-01-02 15:04:05"
	if len(timeStr) == 5 {
		layout = "2006-01-02 15:04"
	}
	t, err := time.ParseInLocation(layout, dateStr+" "+timeStr, time.UTC)
	if err != nil {
		return 0, err
	}
	return t.UnixNano(), nil
}

func (ts *timeServer) IsToday(nano int64) bool {
	t := time.Unix(0, nano).UTC()
	now := time.Now().UTC()
	tLocal := ts.applyOffset(t)
	nowLocal := ts.applyOffset(now)
	return tLocal.Year() == nowLocal.Year() && tLocal.YearDay() == nowLocal.YearDay()
}

func (ts *timeServer) IsPast(nano int64) bool {
	return nano < ts.UnixNano()
}

func (ts *timeServer) IsFuture(nano int64) bool {
	return nano > ts.UnixNano()
}

func (ts *timeServer) LocalMinutesToUnixUTC(dateSec int64, localMinutes int, tz string) int64 {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	utcDate := time.Unix(dateSec, 0).UTC()
	hour := localMinutes / 60
	minute := localMinutes % 60
	localTime := time.Date(utcDate.Year(), utcDate.Month(), utcDate.Day(), hour, minute, 0, 0, loc)
	return localTime.Unix()
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
