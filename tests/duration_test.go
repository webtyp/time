package tests

import (
	"testing"

	"webtyp.com/time"
)

func TestDuration_UnitsAreExact(t *testing.T) {
	if time.Second != 1000000000 {
		t.Fatalf("Second = %d, want 1000000000 ns", int64(time.Second))
	}
	if 1000*time.Millisecond != time.Second || 1000*time.Microsecond != time.Millisecond || 1000*time.Nanosecond != time.Microsecond {
		t.Fatalf("each unit must be 1000 of the one below: %d %d %d %d",
			int64(time.Second), int64(time.Millisecond), int64(time.Microsecond), int64(time.Nanosecond))
	}
}

func TestDuration_MillisecondsDropsTheFraction(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want int64
	}{
		{1500 * time.Millisecond, 1500},
		{2 * time.Second, 2000},
		{999 * time.Microsecond, 0},
		{0, 0},
	}
	for _, c := range cases {
		if got := c.d.Milliseconds(); got != c.want {
			t.Errorf("(%d ns).Milliseconds() = %d, want %d", int64(c.d), got, c.want)
		}
	}
}

func TestNano_AddMovesTheInstantByTheDuration(t *testing.T) {
	start := time.Unix(1700000000).Nano()
	due := start.Add(30 * time.Second)
	if due.Unix() != time.Unix(1700000030) {
		t.Fatalf("Add(30s).Unix() = %d, want 1700000030", int64(due.Unix()))
	}
	if due-start != time.Nano(30*time.Second) {
		t.Fatalf("Add(30s) moved the instant by %d ns, want 30000000000", int64(due-start))
	}
}
