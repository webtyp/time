package time_test

import (
	"testing"

	"webtyp.com/time"
)

func TestZeroDayLimitsIsClosed(t *testing.T) {
	var l time.DayLimits
	if l.Open {
		t.Error("zero DayLimits should be closed (safe default)")
	}
	if ok := l.Contains(0); ok {
		t.Error("a closed day should not contain any minute")
	}
	if _, _, ok := l.Clamp(0, 1440); ok {
		t.Error("Clamp on a closed day should report nothing survives")
	}
}

func TestUnlimitedCoversTheWholeDay(t *testing.T) {
	l := time.Unlimited()
	if !l.Open {
		t.Fatal("Unlimited should be open")
	}
	if l.OpenMin != 0 || l.CloseMin != 1440 {
		t.Fatalf("Unlimited = [%d, %d), want [0, 1440)", l.OpenMin, l.CloseMin)
	}
	for _, m := range []int{0, 1, 479, 1439} {
		if !l.Contains(m) {
			t.Errorf("Unlimited should contain minute %d", m)
		}
	}
	if l.Contains(1440) {
		t.Error("half-open window: 1440 is out")
	}
}

func TestContainsIsHalfOpen(t *testing.T) {
	l := time.DayLimits{Open: true, OpenMin: 480, CloseMin: 1200}
	if !l.Contains(480) {
		t.Error("OpenMin should be inside (half-open)")
	}
	if l.Contains(1200) {
		t.Error("CloseMin should be outside (half-open)")
	}
	if l.Contains(479) || l.Contains(1201) {
		t.Error("minutes outside the window should be out")
	}
}

func TestClampTrimsBothEdges(t *testing.T) {
	l := time.DayLimits{Open: true, OpenMin: 480, CloseMin: 1200}
	s, e, ok := l.Clamp(400, 1300)
	if !ok {
		t.Fatal("overlapping window should survive")
	}
	if s != 480 || e != 1200 {
		t.Errorf("Clamp(400,1300) = [%d,%d), want [480,1200)", s, e)
	}
}

func TestClampOnClosedDayReportsNothingSurvives(t *testing.T) {
	var l time.DayLimits
	if _, _, ok := l.Clamp(0, 1440); ok {
		t.Error("Clamp on a closed day must report nothing survives")
	}
}

func TestClampReportsFalseWhenTheWindowsDoNotOverlap(t *testing.T) {
	l := time.DayLimits{Open: true, OpenMin: 480, CloseMin: 1200}
	if _, _, ok := l.Clamp(0, 400); ok {
		t.Error("block entirely before opening must report false")
	}
	if _, _, ok := l.Clamp(1300, 1440); ok {
		t.Error("block entirely after closing must report false")
	}
	if _, _, ok := l.Clamp(1200, 1300); ok {
		t.Error("block starting exactly at close must report false (half-open)")
	}
}

func TestClampLeavesAnAlreadyInsideRangeUntouched(t *testing.T) {
	l := time.DayLimits{Open: true, OpenMin: 480, CloseMin: 1200}
	s, e, ok := l.Clamp(600, 900)
	if !ok {
		t.Fatal("inside range should survive")
	}
	if s != 600 || e != 900 {
		t.Errorf("Clamp(600,900) = [%d,%d), want [600,900) unchanged", s, e)
	}
}
