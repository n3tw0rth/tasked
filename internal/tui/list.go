package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/n3tw0rth/tasked/internal/tasks"
)

// RenderTasks returns a multi-line string suitable for printing directly.
// Subtasks are nested under their parent when the parent is in ts; the order
// of ts (normally tasks.SortByPriority) is kept within each level.
func RenderTasks(header string, ts []tasks.Task) string {
	var b strings.Builder
	if header != "" {
		b.WriteString(Title.Render(header))
		b.WriteString("\n")
	}
	if len(ts) == 0 {
		b.WriteString(Dim.Render("  (no tasks)\n"))
		return b.String()
	}
	present := make(map[string]bool, len(ts))
	for _, t := range ts {
		present[t.ID] = true
	}
	children := map[string][]tasks.Task{}
	var roots []tasks.Task
	for _, t := range ts {
		// A subtask whose parent isn't shown (completed, or not a search
		// match) is listed at the top level rather than dropped.
		if t.Parent != "" && t.Parent != t.ID && present[t.Parent] {
			children[t.Parent] = append(children[t.Parent], t)
		} else {
			roots = append(roots, t)
		}
	}
	var walk func(t tasks.Task, depth int)
	walk = func(t tasks.Task, depth int) {
		b.WriteString(renderTaskRow(t, depth))
		b.WriteString("\n")
		for _, c := range children[t.ID] {
			walk(c, depth+1)
		}
	}
	for _, t := range roots {
		walk(t, 0)
	}
	return b.String()
}

func renderTaskRow(t tasks.Task, depth int) string {
	indent := strings.Repeat("    ", depth)
	chip := PriorityChip(t.Priority)
	title := t.Title
	if t.Done {
		title = Dim.Render("✓ " + title)
	}
	right := ""
	if t.HasDue {
		right = "  " + Dim.Render(FormatDue(t.Due))
	}
	notes := ""
	if t.Notes != "" {
		notes = "\n" + indent + "    " + Dim.Render(truncate(t.Notes, 80))
	}
	return fmt.Sprintf("%s%s  %s%s%s", indent, chip, title, right, notes)
}

// FormatDue renders a due date relative to today ("tomorrow", "Fri",
// "2d overdue", "Jul 15").
func FormatDue(d time.Time) string {
	return formatDueAt(d, time.Now())
}

func formatDueAt(d, now time.Time) string {
	days := daysBetween(now, d)
	switch {
	case days == 0:
		return "today"
	case days == 1:
		return "tomorrow"
	case days == -1:
		return "yesterday"
	case days < 0:
		return fmt.Sprintf("%dd overdue", -days)
	case days < 7:
		return d.Format("Mon")
	case d.Year() != now.Year():
		return d.Format("Jan 2 2006")
	default:
		return d.Format("Jan 2")
	}
}

// daysBetween counts calendar days from a to b, reading each date in its own
// location; going through UTC keeps DST changes out of the count.
func daysBetween(a, b time.Time) int {
	ua := time.Date(a.Year(), a.Month(), a.Day(), 0, 0, 0, 0, time.UTC)
	ub := time.Date(b.Year(), b.Month(), b.Day(), 0, 0, 0, 0, time.UTC)
	return int(ub.Sub(ua).Hours() / 24)
}

// truncate shortens s to its first line and at most n runes, marking
// anything cut off with an ellipsis.
func truncate(s string, n int) string {
	line, _, more := strings.Cut(s, "\n")
	line = strings.TrimRight(line, "\r")
	r := []rune(line)
	if len(r) > n {
		return string(r[:n-1]) + "…"
	}
	if more {
		return line + " …"
	}
	return line
}
