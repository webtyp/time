//go:build wasm

package time

import (
	"syscall/js"

	. "webtyp.com/fmt"
)

func init() {
	provider = &timeClient{
		dateCtor: js.Global().Get("Date"),
	}
}

// timeClient implements timeProvider for WASM/JS environments using the JavaScript Date API.
type timeClient struct {
	dateCtor js.Value
}

func (tc *timeClient) UnixNano() int64 {
	jsDate := tc.dateCtor.New()
	msTimestamp := jsDate.Call("getTime").Float()
	return int64(msTimestamp) * 1000000
}

func (tc *timeClient) applyOffset(nano int64) js.Value {
	offsetMs := float64(getOffsetMinutes()) * 60000
	return tc.dateCtor.New(float64(nano)/1e6 + offsetMs)
}

func (tc *timeClient) FormatDate(value any) string {
	switch v := value.(type) {
	case int64:
		jsDate := tc.applyOffset(v)
		return jsDate.Call("toISOString").String()[0:10]
	case string:
		if len(v) == 10 && v[4] == '-' && v[7] == '-' {
			return v
		}
	}
	return ""
}

func (tc *timeClient) FormatTime(value any) string {
	switch v := value.(type) {
	case int64: // UnixNano
		jsDate := tc.applyOffset(v)
		hours := jsDate.Call("getUTCHours").Int()
		minutes := jsDate.Call("getUTCMinutes").Int()
		seconds := jsDate.Call("getUTCSeconds").Int()
		return Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	case int16: // Minutes since midnight
		hours := v / 60
		minutes := v % 60
		return Sprintf("%02d:%02d", hours, minutes)
	case string:
		if nano, err := Convert(v).Int64(); err == nil {
			jsDate := tc.applyOffset(nano)
			hours := jsDate.Call("getUTCHours").Int()
			minutes := jsDate.Call("getUTCMinutes").Int()
			seconds := jsDate.Call("getUTCSeconds").Int()
			return Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
		}
		if Count(v, ":") >= 1 {
			return v
		}
	}
	return ""
}

func (tc *timeClient) FormatDateTime(value any) string {
	switch v := value.(type) {
	case int64:
		jsDate := tc.applyOffset(v)
		iso := jsDate.Call("toISOString").String()
		return iso[0:10] + " " + iso[11:19]
	case string:
		if len(v) == 19 && v[4] == '-' && v[7] == '-' && v[10] == ' ' && v[13] == ':' && v[16] == ':' {
			return v
		}
	}
	return ""
}

func (tc *timeClient) FormatDateTimeShort(value any) string {
	switch v := value.(type) {
	case int64:
		jsDate := tc.applyOffset(v)
		iso := jsDate.Call("toISOString").String()
		return iso[0:10] + " " + iso[11:16]
	case string:
		if len(v) == 16 && v[4] == '-' && v[7] == '-' && v[10] == ' ' && v[13] == ':' {
			return v
		}
	}
	return ""
}

func (tc *timeClient) FormatISO8601(nano int64) string {
	// We MUST NOT use applyOffset here. We need raw UTC.
	jsDate := tc.dateCtor.New(float64(nano) / 1e6)

	// 'toISOString' natively outputs 'YYYY-MM-DDTHH:mm:ss.sssZ'
	iso := jsDate.Call("toISOString").String()

	// We slice out the milliseconds to strictly match 'YYYY-MM-DDTHH:MM:SSZ'
	return iso[0:19] + "Z"
}

func (tc *timeClient) FormatCompact(nano int64) string {
	jsDate := tc.dateCtor.New(float64(nano) / 1e6)
	iso := jsDate.Call("toISOString").String()
	// iso = "YYYY-MM-DDTHH:mm:ss.sssZ"
	// extraer: YYYY(0:4) MM(5:7) DD(8:10) HH(11:13) mm(14:16) ss(17:19)
	return iso[0:4] + iso[5:7] + iso[8:10] + iso[11:13] + iso[14:16] + iso[17:19]
}

func (tc *timeClient) ParseDate(dateStr string) (int64, error) {
	if len(dateStr) != 10 || dateStr[4] != '-' || dateStr[7] != '-' {
		return 0, Errf("invalid date format: %s (expected YYYY-MM-DD)", dateStr)
	}
	jsDate := tc.dateCtor.New(dateStr + "T00:00:00Z")
	if jsDate.Call("toString").String() == "Invalid Date" {
		return 0, Errf("invalid date format: %s", dateStr)
	}
	year := jsDate.Call("getUTCFullYear").Int()
	month := jsDate.Call("getUTCMonth").Int() + 1
	day := jsDate.Call("getUTCDate").Int()
	expected := Sprintf("%04d-%02d-%02d", year, month, day)
	if expected != dateStr {
		return 0, Errf("invalid date: %s (auto-corrected to %s)", dateStr, expected)
	}
	ms := jsDate.Call("getTime").Float()
	return int64(ms) * 1000000, nil
}

func (tc *timeClient) ParseTime(timeStr string) (int16, error) {
	return parseTime(timeStr)
}

func (tc *timeClient) ParseDateTime(dateStr, timeStr string) (int64, error) {
	if len(timeStr) == 5 {
		timeStr += ":00"
	}
	isoStr := dateStr + "T" + timeStr + "Z"
	jsDate := tc.dateCtor.New(isoStr)
	if jsDate.Call("toString").String() == "Invalid Date" {
		return 0, Errf("invalid date/time format: %s %s", dateStr, timeStr)
	}
	ms := jsDate.Call("getTime").Float()
	return int64(ms) * 1000000, nil
}

func (tc *timeClient) IsToday(nano int64) bool {
	jsDateLocal := tc.applyOffset(nano)
	nowLocal := tc.applyOffset(tc.UnixNano())
	return jsDateLocal.Call("toDateString").String() == nowLocal.Call("toDateString").String()
}

func (tc *timeClient) IsPast(nano int64) bool {
	return nano < tc.UnixNano()
}

func (tc *timeClient) IsFuture(nano int64) bool {
	return nano > tc.UnixNano()
}

func (tc *timeClient) LocalMinutesToUnixUTC(dateSec int64, localMinutes int, tz string) int64 {
	offsetSec := int64(getOffsetMinutes()) * 60
	midnight := MidnightUTC(dateSec)
	return midnight + int64(localMinutes)*60 - offsetSec
}

type WasmTimer struct {
	id     js.Value
	active bool
	jsFunc js.Func
	f      func()
}

func (wt *WasmTimer) Stop() bool {
	if !wt.active {
		return false
	}
	js.Global().Call("clearTimeout", wt.id)
	wt.active = false
	wt.jsFunc.Release()
	return true
}

func (wt *WasmTimer) Fire() {
	if !wt.active {
		return
	}
	wt.active = false
	wt.jsFunc.Release()
	if wt.f != nil {
		wt.f()
	}
}

func (tc *timeClient) AfterFunc(milliseconds int, f func()) Timer {
	wt := &WasmTimer{
		active: true,
		f:      f,
	}
	wt.jsFunc = js.FuncOf(func(this js.Value, args []js.Value) any {
		wt.Fire()
		return nil
	})
	wt.id = js.Global().Call("setTimeout", wt.jsFunc, milliseconds)
	return wt
}
