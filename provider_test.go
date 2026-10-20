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

	t.Run("Clocks", func(t *testing.T) { ClocksShared(t) })
	t.Run("Format", func(t *testing.T) { FormatShared(t) })
	t.Run("FormatOffset", func(t *testing.T) { FormatOffsetShared(t) })
	t.Run("ParseDate", func(t *testing.T) { ParseDateShared(t) })
	t.Run("ParseTime", func(t *testing.T) { ParseTimeShared(t) })
	t.Run("ParseDateTime", func(t *testing.T) { ParseDateTimeShared(t) })
	t.Run("Relative", func(t *testing.T) { RelativeShared(t) })
	t.Run("Calendar", func(t *testing.T) { CalendarShared(t) })
	t.Run("LocalMinutesToUnixUTC", func(t *testing.T) { LocalMinutesToUnixUTCShared(t) })
}
