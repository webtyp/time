package time

import (
	. "webtyp.com/fmt"
)

// Names and prefixes are data, never words to translate.
var (
	pkgPrefix   = "time:"
	exampleZone = "America/Santiago"
)

// ErrUnknownZone: the name is not a time zone this platform knows. Never replaced by UTC in silence.
var ErrUnknownZone = Err(pkgPrefix, "unknown time zone — use an IANA name such as", exampleZone)

// errZoneNotLoaded is a programming error: a Zone used without LoadZone or MachineZone.
const errZoneNotLoaded = "time: Zone not loaded — create it with time.LoadZone or time.MachineZone"

// Zone is a place's clock: it turns an instant (Unix, UTC) into that place's date and time, and back, with the
// offset that place had at that instant (summer and winter time included). Every local date or time in webtyp comes
// from a Zone; there is no implicit "local" anywhere.
//
// Example:
//
//	zone, err := time.LoadZone("America/Santiago")
//	start, err := zone.At("2019-07-15", 21*60+30) // 2019-07-16 01:30 UTC (winter, UTC-4)
//	zone.Clock(start)                             // "21:30"
type Zone struct {
	name    string
	machine bool
}

// LoadZone loads an IANA time zone by name ("America/Santiago"). An unknown or empty name returns ErrUnknownZone.
//
// Example: zone, err := time.LoadZone(deps.Timezone)
func LoadZone(name string) (Zone, error) {
	if name == "" {
		return Zone{}, ErrUnknownZone
	}
	if _, err := provider.zoneOffset(name, NowUnix()); err != nil {
		return Zone{}, ErrUnknownZone
	}
	return Zone{name: name}, nil
}

// MachineZone is the zone of this computer: the process's on a server, the visitor's in a browser. Only for tools
// (logs, a terminal UI); an app shows its installation's zone, loaded with LoadZone.
//
// Example: stamp := time.MachineZone().Clock(time.NowUnix())
func MachineZone() Zone { return Zone{machine: true} }

// Name is the IANA name of the zone ("America/Santiago"); for MachineZone, the name the platform reports.
func (z Zone) Name() string {
	z.mustBeLoaded()
	if z.machine {
		return provider.machineZoneName()
	}
	return z.name
}

// Offset is the zone's distance from UTC at the instant u, in seconds (-14400 for UTC-4).
func (z Zone) Offset(u Unix) int {
	z.mustBeLoaded()
	key := z.name
	if z.machine {
		key = ""
	}
	off, err := provider.zoneOffset(key, u)
	if err != nil {
		panic(errZoneNotLoaded) // a loaded zone always resolves
	}
	return off
}

// wall is the instant shifted to the zone's clock, read as UTC.
func (z Zone) wall(u Unix) Unix { return u + Unix(z.Offset(u)) }

// Date is the zone's calendar date of the instant: "2019-07-15".
func (z Zone) Date(u Unix) string {
	y, m, d, _, _, _ := CivilUTC(z.wall(u))
	return Sprintf("%04d-%02d-%02d", y, m, d)
}

// Clock is the zone's hour and minute of the instant: "21:30".
func (z Zone) Clock(u Unix) string {
	_, _, _, h, mi, _ := CivilUTC(z.wall(u))
	return Sprintf("%02d:%02d", h, mi)
}

// DateTime is Date and Clock together: "2019-07-15 21:30".
func (z Zone) DateTime(u Unix) string { return z.Date(u) + " " + z.Clock(u) }

// Minutes is the minutes since the zone's midnight at the instant (21:30 → 1290).
func (z Zone) Minutes(u Unix) int {
	_, _, _, h, mi, _ := CivilUTC(z.wall(u))
	return h*60 + mi
}

// Weekday is the zone's day of the week of the instant (0 = Sunday … 6 = Saturday).
func (z Zone) Weekday(u Unix) int { return WeekdayUTC(z.wall(u)) }

// IsToday reports whether the instant falls on today's date in the zone.
func (z Zone) IsToday(u Unix) bool { return z.Date(u) == z.Date(NowUnix()) }

// At is the instant of a local date ("YYYY-MM-DD") and minutes since its midnight (0–1439) in the zone. In the hour a
// clock skips when summer time starts, the result is the same minute one hour later.
//
// Example: start, err := zone.At("2019-07-15", 21*60+30)
func (z Zone) At(date string, minutes int) (Unix, error) {
	z.mustBeLoaded()
	midnight, err := ParseDateUTC(date)
	if err != nil {
		return 0, err
	}
	if minutes < 0 || minutes >= 24*60 {
		return 0, Err(pkgPrefix, "minutes must be between", 0, "and", 24*60-1)
	}
	wall := midnight + Unix(minutes)*Minute
	// Two passes: the offset at the first guess can differ from the offset at the answer across a DST change.
	guess := wall - Unix(z.Offset(wall))
	return wall - Unix(z.Offset(guess)), nil
}

// Day is the range of instants of a local date in the zone: from (its midnight, included) to the next midnight
// (excluded). Search "the bookings of a day" with from <= start < to.
//
// Example: from, to, err := zone.Day("2019-07-15")
func (z Zone) Day(date string) (from, to Unix, err error) {
	if from, err = z.At(date, 0); err != nil {
		return 0, 0, err
	}
	midnight, _ := ParseDateUTC(date)
	y, m, d, _, _, _ := CivilUTC(midnight + Day)
	to, err = z.At(Sprintf("%04d-%02d-%02d", y, m, d), 0)
	return from, to, err
}

func (z Zone) mustBeLoaded() {
	if z.name == "" && !z.machine {
		panic(errZoneNotLoaded)
	}
}
