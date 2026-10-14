package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/n3tw0rth/tasked/internal/tasks"
)

func TestFormatDueAt(t *testing.T) {
	west := time.FixedZone("-0500", -5*3600)
	now := time.Date(2026, 7, 15, 15, 0, 0, 0, west) // a Wednesday afternoon
	day := func(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, west) }
	tests := []struct {
		due  time.Time
		want string
	}{
		{day(2026, 7, 15), "today"},
		{day(2026, 7, 16), "tomorrow"}, // used to show "today" west of UTC
		{day(2026, 7, 14), "yesterday"},
		{day(2026, 7, 12), "3d overdue"},
		{day(2026, 7, 18), "Sat"},
		{day(2026, 8, 14), "Aug 14"},
		{day(2027, 1, 5), "Jan 5 2027"},
	}
	for _, tt := range tests {
		if got := formatDueAt(tt.due, now); got != tt.want {
			t.Errorf("formatDueAt(%s) = %q, want %q", tt.due.Format("2006-01-02"), got, tt.want)
		}
	}
}

func TestFormatDueAtAcrossDST(t *testing.T) {
	ny, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("tzdata unavailable:", err)
	}
	// DST starts 2026-03-08, so the two midnights are 47 hours apart.
	now := time.Date(2026, 3, 7, 12, 0, 0, 0, ny)
	due := time.Date(2026, 3, 9, 0, 0, 0, 0, ny)
	if got := formatDueAt(due, now); got != "Mon" {
		t.Errorf("formatDueAt across DST = %q, want %q", got, "Mon")
	}
}

func TestRenderTasksNestsSubtasks(t *testing.T) {
	ts := []tasks.Task{ // already priority-sorted, as Client.Tasks returns them
		{ID: "a", Title: "Groceries", Priority: 1},
		{ID: "x", Parent: "gone", Title: "Orphan", Priority: 1},
		{ID: "a2", Parent: "a", Title: "Eggs", Priority: 2},
		{ID: "b", Title: "Taxes", Priority: 3, Notes: "receipts"},
		{ID: "a1", Parent: "a", Title: "Milk", Priority: 4},
		{ID: "b1", Parent: "b", Title: "Find forms", Priority: 5, Notes: "W-2"},
	}
	// lipgloss renders plain text when stdout isn't a terminal, as under
	// go test; chips keep their one-space padding.
	want := strings.Join([]string{
		"Home",
		" p1   Groceries",
		"     p2   Eggs",
		"     p4   Milk",
		" p1   Orphan",
		" p3   Taxes",
		"    receipts",
		"     p5   Find forms",
		"        W-2",
		"",
	}, "\n")
	if got := RenderTasks("Home", ts); got != want {
		t.Errorf("RenderTasks =\n%s\nwant\n%s", got, want)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		in   string
		n    int
		want string
	}{
		{"short", 10, "short"},
		{"héllo wörld", 5, "héll…"}, // used to cut mid-rune
		{"සිංහල අකුරු", 4, "සිං…"},
		{"line one\nline two", 80, "line one …"},
		{"crlf\r\nnext", 80, "crlf …"},
	}
	for _, tt := range tests {
		if got := truncate(tt.in, tt.n); got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.in, tt.n, got, tt.want)
		}
	}
}
