package tests

import (
	"testing"

	"webtyp.com/time"
)

func TestDateAndCivilDate(t *testing.T) {
	cases := []struct {
		y, m, d, h, min, s int
		wantSec            int64
	}{
		{1970, 1, 1, 0, 0, 0, 0},
		{1970, 1, 1, 1, 30, 15, 5415},
		{2025, 1, 6, 0, 0, 0, 1736121600},
		{2025, 1, 6, 10, 0, 0, 1736157600},
		{2026, 3, 2, 9, 30, 0, 1772443800},
	}

	for _, c := range cases {
		got := time.Date(c.y, c.m, c.d, c.h, c.min, c.s)
		if got != c.wantSec {
			t.Errorf("Date(%d,%d,%d,%d,%d,%d) = %d, want %d", c.y, c.m, c.d, c.h, c.min, c.s, got, c.wantSec)
		}

		ry, rm, rd, rh, rmin, rs := time.CivilDate(got)
		if ry != c.y || rm != c.m || rd != c.d || rh != c.h || rmin != c.min || rs != c.s {
			t.Errorf("CivilDate(%d) = (%d,%d,%d,%d,%d,%d), want (%d,%d,%d,%d,%d,%d)",
				got, ry, rm, rd, rh, rmin, rs, c.y, c.m, c.d, c.h, c.min, c.s)
		}
	}
}

func TestDateKeyToUnix(t *testing.T) {
	if got := time.DateKeyToUnix("1970-01-01"); got != 0 {
		t.Errorf("DateKeyToUnix(1970-01-01) = %d, want 0", got)
	}
	if got := time.DateKeyToUnix("2025-01-06"); got != 1736121600 {
		t.Errorf("DateKeyToUnix(2025-01-06) = %d, want 1736121600", got)
	}
	if got := time.DateKeyToUnix("invalid-date"); got != 0 {
		t.Errorf("DateKeyToUnix(invalid) = %d, want 0", got)
	}
}
