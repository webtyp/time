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

// local is the JS Date of an instant shifted by the local offset, read with its UTC getters.
func (tc *timeClient) local(u Unix) js.Value {
	offsetMs := float64(getOffsetMinutes()) * 60000
	return tc.dateCtor.New(float64(u)*1000 + offsetMs)
}

func (tc *timeClient) iso(u Unix) string { return tc.local(u).Call("toISOString").String() }

func (tc *timeClient) FormatDate(u Unix) string { return tc.iso(u)[0:10] }
func (tc *timeClient) FormatTime(u Unix) string { return tc.iso(u)[11:19] }
func (tc *timeClient) FormatDateTime(u Unix) string {
	iso := tc.iso(u)
	return iso[0:10] + " " + iso[11:19]
}
func (tc *timeClient) FormatDateTimeShort(u Unix) string {
	iso := tc.iso(u)
	return iso[0:10] + " " + iso[11:16]
}

// utcISO is the raw UTC ISO string, with no offset applied.
func (tc *timeClient) utcISO(u Unix) string {
	return tc.dateCtor.New(float64(u) * 1000).Call("toISOString").String()
}

func (tc *timeClient) FormatISO8601(u Unix) string { return tc.utcISO(u)[0:19] + "Z" }
func (tc *timeClient) FormatCompact(u Unix) string {
	iso := tc.utcISO(u) // YYYY-MM-DDTHH:mm:ss.sssZ
	return iso[0:4] + iso[5:7] + iso[8:10] + iso[11:13] + iso[14:16] + iso[17:19]
}

func (tc *timeClient) ParseDate(dateStr string) (Unix, error) {
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
	return Unix(int64(jsDate.Call("getTime").Float()) / 1000), nil
}

func (tc *timeClient) ParseTime(timeStr string) (int16, error) {
	return parseTime(timeStr)
}

func (tc *timeClient) ParseDateTime(dateStr, timeStr string) (Unix, error) {
	if len(timeStr) == 5 {
		timeStr += ":00"
	}
	jsDate := tc.dateCtor.New(dateStr + "T" + timeStr + "Z")
	if jsDate.Call("toString").String() == "Invalid Date" {
		return 0, Errf("invalid date/time format: %s %s", dateStr, timeStr)
	}
	return Unix(int64(jsDate.Call("getTime").Float()) / 1000), nil
}

func (tc *timeClient) IsToday(u Unix) bool {
	return tc.FormatDate(u) == tc.FormatDate(NowUnix())
}

// LocalMinutesToUnixUTC resolves tz through the browser's Intl database, at the
// target date — never the browser's own offset, and never today's offset: a
// Santiago clinic read in June is UTC-4 even when "now" is UTC-3, whatever zone
// the viewer's browser is in. An invalid zone falls back to UTC, like the
// backend's time.LoadLocation path.
func (tc *timeClient) LocalMinutesToUnixUTC(dateUTC Unix, localMinutes int, tz string) Unix {
	wall := int64(MidnightUTC(dateUTC)) + int64(localMinutes)*60 // the local wall clock, read as if UTC
	// Two passes, as for any zone conversion: the offset at the first guess can
	// differ from the offset at the answer when a DST change lies in between.
	guess := wall - zoneOffsetSec(wall, tz)
	return Unix(wall - zoneOffsetSec(guess, tz))
}

// zoneOffsetSec is tz's UTC offset in seconds at the instant utcSec: the zone's
// wall clock at that instant (from Intl), read as UTC, minus the instant.
// 0 (UTC) when tz is not a zone the runtime knows.
func zoneOffsetSec(utcSec int64, tz string) (offset int64) {
	defer func() {
		if recover() != nil { // Intl throws RangeError on an unknown zone
			offset = 0
		}
	}()
	opts := js.Global().Get("Object").New()
	opts.Set("timeZone", tz)
	opts.Set("hourCycle", "h23")
	for _, field := range [...]string{"year", "month", "day", "hour", "minute", "second"} {
		opts.Set(field, "numeric")
	}
	fmtr := js.Global().Get("Intl").Get("DateTimeFormat").New("en-US", opts)
	parts := fmtr.Call("formatToParts", js.Global().Get("Date").New(float64(utcSec)*1000))

	var y, mo, d, h, mi, sec int
	number := js.Global().Get("Number")
	for i := 0; i < parts.Length(); i++ {
		p := parts.Index(i)
		v := number.Invoke(p.Get("value")).Int()
		switch p.Get("type").String() {
		case "year":
			y = v
		case "month":
			mo = v
		case "day":
			d = v
		case "hour":
			h = v
		case "minute":
			mi = v
		case "second":
			sec = v
		}
	}
	wallMs := js.Global().Get("Date").Call("UTC", y, mo-1, d, h, mi, sec).Float()
	return int64(wallMs/1000) - utcSec
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
