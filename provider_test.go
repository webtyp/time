package time_test

import (
	"testing"
)

// RunAPITests runs all direct API tests.
func RunAPITests(t *testing.T) {
	t.Run("Clocks", func(t *testing.T) { ClocksShared(t) })
	t.Run("Format", func(t *testing.T) { FormatShared(t) })
	t.Run("ParseDate", func(t *testing.T) { ParseDateShared(t) })
	t.Run("ParseTime", func(t *testing.T) { ParseTimeShared(t) })
	t.Run("Relative", func(t *testing.T) { RelativeShared(t) })
	t.Run("Calendar", func(t *testing.T) { CalendarShared(t) })
	t.Run("Zone", func(t *testing.T) { ZoneShared(t) })
}
