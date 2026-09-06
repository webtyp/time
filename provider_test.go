package time_test

import (
	"testing"

	"webtyp.com/time"
)

// RunAPITests runs all direct API tests.
func RunAPITests(t *testing.T) {
	// Set offset to 0 for consistent test results in shared tests (expecting UTC)
	initialOffset := time.GetTimeZoneOffset()
	time.SetTimeZoneOffset(0)
	defer time.SetTimeZoneOffset(initialOffset)

	t.Run("UnixNano", func(t *testing.T) { UnixNanoShared(t) })
	t.Run("FormatDate", func(t *testing.T) { FormatDateShared(t) })
	t.Run("FormatTime", func(t *testing.T) { FormatTimeShared(t) })
	t.Run("FormatDateTime", func(t *testing.T) { FormatDateTimeShared(t) })
	t.Run("FormatDateTimeShort", func(t *testing.T) { FormatDateTimeShortShared(t) })
	t.Run("FormatISO8601", func(t *testing.T) { FormatISO8601Shared(t) })
	t.Run("FormatCompact", func(t *testing.T) { FormatCompactShared(t) })
	t.Run("ParseDate", func(t *testing.T) { ParseDateShared(t) })
	t.Run("ParseTime", func(t *testing.T) { ParseTimeShared(t) })
	t.Run("ParseDateTime", func(t *testing.T) { ParseDateTimeShared(t) })
	t.Run("IsToday", func(t *testing.T) { IsTodayShared(t) })
	t.Run("IsPast", func(t *testing.T) { IsPastShared(t) })
	t.Run("IsFuture", func(t *testing.T) { IsFutureShared(t) })
	t.Run("DaysBetween", func(t *testing.T) { DaysBetweenSharedTest(t) })
	t.Run("Weekday", func(t *testing.T) { WeekdayShared(t) })
	t.Run("MidnightUTC", func(t *testing.T) { MidnightUTCShared(t) })
	t.Run("LocalMinutesToUnixUTC", func(t *testing.T) { LocalMinutesToUnixUTCShared(t) })
}
