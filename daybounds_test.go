package time_test

import (
	"testing"

	"webtyp.com/time"
)

func TestZeroDayBoundsIsClosed(t *testing.T) {
	var b time.DayBounds
	if b.Open {
		t.Error("zero DayBounds should be closed (safe default)")
	}
	if ok := b.Contains(0); ok {
		t.Error("a closed day should not contain any minute")
	}
	if _, _, ok := b.Clamp(0, 1440); ok {
		t.Error("Clamp on a closed day should report nothing survives")
	}
}

func TestUnboundedCoversTheWholeDay(t *testing.T) {
	b := time.Unbounded()
	if !b.Open {
		t.Fatal("Unbounded should be open")
	}
	if b.OpenMin != 0 || b.CloseMin != 1440 {
		t.Fatalf("Unbounded = [%d, %d), want [0, 1440)", b.OpenMin, b.CloseMin)
	}
	for _, m := range []int{0, 1, 479, 1439} {
		if !b.Contains(m) {
			t.Errorf("Unbounded should contain minute %d", m)
		}
	}
	if b.Contains(1440) {
		t.Error("half-open window: 1440 is out")
	}
}

func TestContainsIsHalfOpen(t *testing.T) {
	b := time.DayBounds{Open: true, OpenMin: 480, CloseMin: 1200}
	if !b.Contains(480) {
		t.Error("OpenMin should be inside (half-open)")
	}
	if b.Contains(1200) {
		t.Error("CloseMin should be outside (half-open)")
	}
	if b.Contains(479) || b.Contains(1201) {
		t.Error("minutes outside the window should be out")
	}
}

func TestClampTrimsBothEdges(t *testing.T) {
	b := time.DayBounds{Open: true, OpenMin: 480, CloseMin: 1200}
	s, e, ok := b.Clamp(400, 1300)
	if !ok {
		t.Fatal("overlapping window should survive")
	}
	if s != 480 || e != 1200 {
		t.Errorf("Clamp(400,1300) = [%d,%d), want [480,1200)", s, e)
	}
}

func TestClampOnClosedDayReportsNothingSurvives(t *testing.T) {
	var b time.DayBounds
	if _, _, ok := b.Clamp(0, 1440); ok {
		t.Error("Clamp on a closed day must report nothing survives")
	}
}

func TestClampReportsFalseWhenTheWindowsDoNotOverlap(t *testing.T) {
	b := time.DayBounds{Open: true, OpenMin: 480, CloseMin: 1200}
	if _, _, ok := b.Clamp(0, 400); ok {
		t.Error("block entirely before opening must report false")
	}
	if _, _, ok := b.Clamp(1300, 1440); ok {
		t.Error("block entirely after closing must report false")
	}
	if _, _, ok := b.Clamp(1200, 1300); ok {
		t.Error("block starting exactly at close must report false (half-open)")
	}
}

func TestClampLeavesAnAlreadyInsideRangeUntouched(t *testing.T) {
	b := time.DayBounds{Open: true, OpenMin: 480, CloseMin: 1200}
	s, e, ok := b.Clamp(600, 900)
	if !ok {
		t.Fatal("inside range should survive")
	}
	if s != 600 || e != 900 {
		t.Errorf("Clamp(600,900) = [%d,%d), want [600,900) unchanged", s, e)
	}
}
