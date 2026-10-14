package tasks

import (
	"testing"
	"time"
)

func TestParsePriority(t *testing.T) {
	tests := map[string]int{
		"":              Default,
		"no tag":        Default,
		"[p1] foo":      1,
		"foo [P2]":      2,
		"[p5] last":     5,
		"[p9] invalid":  Default,
		"[p3] x [p1] y": 3,
	}
	for in, want := range tests {
		if got := ParsePriority(in); got != want {
			t.Errorf("ParsePriority(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestStripPriority(t *testing.T) {
	tests := map[string]string{
		"[p1] foo":     "foo",
		"foo [p2] bar": "foo bar",
		"foo [p2]":     "foo",
		"[p1]":         "",
		"  plain  ":    "plain",
		"[p1] a [p2]":  "a [p2]",
		"[x] [p3] y":   "[x] y",
	}
	for in, want := range tests {
		if got := StripPriority(in); got != want {
			t.Errorf("StripPriority(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestUpsertPriority(t *testing.T) {
	tests := []struct {
		notes string
		p     int
		want  string
	}{
		{"", 2, "[p2]"},
		{"[p1] foo", 4, "[p4] foo"},
		{"foo [p3] bar", 1, "[p1] foo bar"},
		{"x", 0, "[p1] x"},
		{"x", 9, "[p5] x"},
	}
	for _, tt := range tests {
		if got := UpsertPriority(tt.notes, tt.p); got != tt.want {
			t.Errorf("UpsertPriority(%q, %d) = %q, want %q", tt.notes, tt.p, got, tt.want)
		}
	}
}

func TestHasPriority(t *testing.T) {
	if !HasPriority("todo [P4]") || HasPriority("todo") || HasPriority("[p0]") {
		t.Error("HasPriority misclassified notes")
	}
}

func TestSortByPriority(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2026, 7, d, 0, 0, 0, 0, time.UTC) }
	ts := []Task{
		{Title: "b no due", Priority: 2},
		{Title: "late", Priority: 2, Due: day(20), HasDue: true},
		{Title: "low", Priority: 5},
		{Title: "early", Priority: 2, Due: day(10), HasDue: true},
		{Title: "A no due", Priority: 2},
		{Title: "top", Priority: 1},
	}
	SortByPriority(ts)
	want := []string{"top", "early", "late", "A no due", "b no due", "low"}
	for i, w := range want {
		if ts[i].Title != w {
			t.Fatalf("position %d = %q, want %q (full order: %v)", i, ts[i].Title, w, titles(ts))
		}
	}
}

func TestTaskUpdateIsZero(t *testing.T) {
	if !(TaskUpdate{}).IsZero() {
		t.Error("empty update not zero")
	}
	title := "x"
	if (TaskUpdate{Title: &title}).IsZero() || (TaskUpdate{ClearDue: true}).IsZero() {
		t.Error("non-empty update reported zero")
	}
}

func titles(ts []Task) []string {
	out := make([]string, len(ts))
	for i, t := range ts {
		out[i] = t.Title
	}
	return out
}
