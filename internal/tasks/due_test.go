package tasks

import (
	"testing"
	"time"
)

var colombo = time.FixedZone("+0530", 5*3600+1800)

func TestParseDueAt(t *testing.T) {
	now := time.Date(2026, 7, 15, 10, 0, 0, 0, colombo) // a Wednesday
	tests := []struct {
		in   string
		want string
	}{
		{"today", "2026-07-15"},
		{"TMR", "2026-07-16"},
		{"tomorrow", "2026-07-16"},
		{"eow", "2026-07-19"},
		{"fri", "2026-07-17"},
		{"wed", "2026-07-22"}, // strictly in the future
		{"3d", "2026-07-18"},
		{"2w", "2026-07-29"},
		{"+1", "2026-07-16"},
		{"2026-08-01", "2026-08-01"},
	}
	for _, tt := range tests {
		got, err := parseDueAt(tt.in, now)
		if err != nil {
			t.Errorf("parseDueAt(%q): %v", tt.in, err)
			continue
		}
		if s := got.Format("2006-01-02"); s != tt.want {
			t.Errorf("parseDueAt(%q) = %s, want %s", tt.in, s, tt.want)
		}
		if got.Location() != colombo || got.Hour() != 0 || got.Minute() != 0 {
			t.Errorf("parseDueAt(%q) = %v, want local midnight", tt.in, got)
		}
	}
}

func TestParseDueAtRejects(t *testing.T) {
	now := time.Date(2026, 7, 15, 10, 0, 0, 0, colombo)
	for _, in := range []string{"", "  ", "someday", "2026-08-01 14:30", "2026-08-01T14:30", "2026-08-01T14:30:00Z"} {
		if got, err := parseDueAt(in, now); err == nil {
			t.Errorf("parseDueAt(%q) = %v, want error", in, got)
		}
	}
}

// Regression: dates used to be sent as local time converted to UTC, which
// moved them a day earlier east of UTC (and a day later west of it for the
// 23:59 shorthand forms), because Google keeps only the UTC date part.
func TestDueRoundTripAcrossTimezones(t *testing.T) {
	for _, off := range []int{14, 5, 0, -5, -11, -12} {
		loc := time.FixedZone("", off*3600+1800*sign(off))
		lateEvening := time.Date(2026, 7, 15, 23, 30, 0, 0, loc)
		for in, want := range map[string]string{
			"2026-07-15": "2026-07-15T00:00:00Z",
			"today":      "2026-07-15T00:00:00Z",
			"tmr":        "2026-07-16T00:00:00Z",
		} {
			d, err := parseDueAt(in, lateEvening)
			if err != nil {
				t.Fatalf("UTC%+d parseDueAt(%q): %v", off, in, err)
			}
			got := formatAPIDue(d)
			if got != want {
				t.Errorf("UTC%+d: %q sent as %s, want %s", off, in, got, want)
			}
			back, ok := parseAPIDue(got, loc)
			if !ok || !back.Equal(d) {
				t.Errorf("UTC%+d: %s read back as %v, want %v", off, got, back, d)
			}
		}
	}
}

func TestParseAPIDue(t *testing.T) {
	west := time.FixedZone("-0500", -5*3600)
	got, ok := parseAPIDue("2026-07-16T00:00:00.000Z", west)
	if !ok {
		t.Fatal("parseAPIDue failed")
	}
	if want := time.Date(2026, 7, 16, 0, 0, 0, 0, west); !got.Equal(want) {
		t.Errorf("parseAPIDue = %v, want %v", got, want)
	}
	if _, ok := parseAPIDue("not a date", west); ok {
		t.Error("parseAPIDue accepted garbage")
	}
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	}
	return 0
}
