package time_test

import (
	"testing"

	"webtyp.com/time"
)

func TestTimeZoneOffset(t *testing.T) {
	// Default should be detected or 0
	initial := time.GetTimeZoneOffset()
	t.Logf("Initial detected offset: %d", initial)

	// Set manual offset
	time.SetTimeZoneOffset(-3)
	if time.GetTimeZoneOffset() != -3 {
		t.Errorf("SetTimeZoneOffset(-3) failed, got %d", time.GetTimeZoneOffset())
	}

	time.SetTimeZoneOffset(5)
	if time.GetTimeZoneOffset() != 5 {
		t.Errorf("SetTimeZoneOffset(5) failed, got %d", time.GetTimeZoneOffset())
	}

	// Restore something reasonable
	time.SetTimeZoneOffset(0)
}

func TestFormatWithOffset(t *testing.T) {
	// 2021-01-01 00:00:00 UTC
	nano := int64(1609459200 * 1000000000)

	// In UTC
	time.SetTimeZoneOffset(0)
	if time.FormatTime(nano) != "00:00:00" {
		t.Errorf("FormatTime(UTC) = %s, want 00:00:00", time.FormatTime(nano))
	}

	// In UTC-3
	time.SetTimeZoneOffset(-3)
	if time.FormatTime(nano) != "21:00:00" {
		t.Errorf("FormatTime(UTC-3) = %s, want 21:00:00", time.FormatTime(nano))
	}

	// In UTC+1
	time.SetTimeZoneOffset(1)
	if time.FormatTime(nano) != "01:00:00" {
		t.Errorf("FormatTime(UTC+1) = %s, want 01:00:00", time.FormatTime(nano))
	}
}

func TestIsTodayWithOffset(t *testing.T) {
	// Near midnight UTC
	// 2021-01-02 01:00:00 UTC
	nano := int64(1609549200 * 1000000000)

	// If it's 2021-01-02 01:00:00 UTC
	// In UTC-3 it's 2021-01-01 22:00:00 (Previous day)

	// We need a reference "now" to test IsToday, but Now() changes.
	// So we'll just verify FormatDate behaves correctly for today's logic.

	time.SetTimeZoneOffset(0)
	dateUTC := time.FormatDate(nano)

	time.SetTimeZoneOffset(-3)
	dateLocal := time.FormatDate(nano)

	if dateUTC == dateLocal {
		t.Errorf("FormatDate should differ near midnight UTC with offset: UTC=%s, UTC-3=%s", dateUTC, dateLocal)
	}
}
