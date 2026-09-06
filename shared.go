package time

import (
	. "webtyp.com/fmt"
)

// parseTime is a shared helper function for parsing time strings ("HH:MM" or "HH:MM:SS").
func parseTime(timeStr string) (int16, error) {
	parts := Convert(timeStr).Split(":")
	if len(parts) < 2 {
		return 0, Errf("invalid time format: %s", timeStr)
	}
	hours, err := Convert(parts[0]).Int()
	if err != nil || hours < 0 || hours > 23 {
		return 0, Errf("invalid hours: %s", parts[0])
	}
	minutes, err := Convert(parts[1]).Int()
	if err != nil || minutes < 0 || minutes > 59 {
		return 0, Errf("invalid minutes: %s", parts[1])
	}
	return int16(hours*60 + minutes), nil
}

// DaysBetweenShared is a shared helper function for calculating the number of full days between two timestamps.
func DaysBetweenShared(nano1, nano2 int64) int {
	// 86400000000000 nanoseconds in a day
	const nanosInDay = 86400000000000
	return int((nano2 - nano1) / nanosInDay)
}
