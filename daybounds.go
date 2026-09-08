package time

// DayBounds is the usable window of one date, in minutes from midnight.
//
// It is the value a day-window constraint crosses a module boundary as: the
// module that computes it (an establishment's calendar) and the module that
// obeys it (a booking schedule) agree through this type instead of importing
// each other.
//
// Open == false means the date is unusable and OpenMin/CloseMin carry no
// meaning. The zero value is therefore "closed", which is the safe default:
// the absence of a computed window never means "all day".
//
// It deliberately carries NO reason. Why a day is closed — a holiday, a local
// closure, a weekly rule — is the producer's business and stays with the
// producer; a consumer that only has to skip the day does not need it, and
// putting it here would drag one domain's vocabulary into every other's.
type DayBounds struct {
	Open     bool
	OpenMin  int
	CloseMin int
}

// Unbounded is the whole day, 00:00 to 24:00. It is what a caller with no
// institutional calendar uses, so "no constraint" is an explicit value rather
// than a nil check repeated at every call site.
func Unbounded() DayBounds { return DayBounds{Open: true, OpenMin: 0, CloseMin: 1440} }

// Contains reports whether a minute-of-day falls inside the window.
// Half-open: OpenMin is inside, CloseMin is not.
func (b DayBounds) Contains(minute int) bool {
	return b.Open && minute >= b.OpenMin && minute < b.CloseMin
}

// Clamp narrows [startMin, endMin) to what b allows, and reports whether
// anything survived. It is the operation a scheduler actually performs: a
// professional's block trimmed to the establishment's opening window.
func (b DayBounds) Clamp(startMin, endMin int) (int, int, bool) {
	if !b.Open {
		return 0, 0, false
	}
	if startMin < b.OpenMin {
		startMin = b.OpenMin
	}
	if endMin > b.CloseMin {
		endMin = b.CloseMin
	}
	if startMin >= endMin {
		return 0, 0, false
	}
	return startMin, endMin, true
}
