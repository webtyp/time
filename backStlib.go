//go:build !wasm

package time

import (
	"sync"
	"time"
	_ "time/tzdata" // every IANA zone inside the binary: no dependence on the OS or a Docker image's files
)

func init() {
	provider = &timeServer{}
}

// timeServer implements timeProvider for standard Go.
type timeServer struct{}

func (ts *timeServer) UnixNano() int64 {
	return time.Now().UTC().UnixNano()
}

func (ts *timeServer) FormatISO8601(u Unix) string {
	return time.Unix(int64(u), 0).UTC().Format(time.RFC3339)
}
func (ts *timeServer) FormatCompact(u Unix) string {
	return time.Unix(int64(u), 0).UTC().Format("20060102150405")
}

func (ts *timeServer) ParseDateUTC(dateStr string) (Unix, error) {
	t, err := time.ParseInLocation("2006-01-02", dateStr, time.UTC)
	if err != nil {
		return 0, err
	}
	return Unix(t.Unix()), nil
}

func (ts *timeServer) ParseTime(timeStr string) (int16, error) {
	return parseTime(timeStr)
}

// zones caches each loaded location: LoadLocation reads the embedded database once per name.
var zones sync.Map // name → *time.Location

func (ts *timeServer) zoneOffset(name string, u Unix) (int, error) {
	loc := time.Local
	if name != "" {
		cached, ok := zones.Load(name)
		if !ok {
			l, err := time.LoadLocation(name)
			if err != nil {
				return 0, err
			}
			cached, _ = zones.LoadOrStore(name, l)
		}
		loc = cached.(*time.Location)
	}
	_, off := time.Unix(int64(u), 0).In(loc).Zone()
	return off, nil
}

// machineZoneName is the process's zone: the TZ variable when set, else what the system reports.
func (ts *timeServer) machineZoneName() string { return time.Local.String() }

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
