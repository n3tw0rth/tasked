package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/n3tw0rth/tasked/internal/tasks"
)

// RenderTasks returns a multi-line string suitable for printing directly.
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
	for _, t := range ts {
		b.WriteString(renderTaskRow(t))
		b.WriteString("\n")
	}
	return b.String()
}

func renderTaskRow(t tasks.Task) string {
	chip := PriorityChip(t.Priority)
	title := t.Title
	if t.Done {
		title = Dim.Render("✓ " + title)
	}
	right := ""
	if t.HasDue {
		right = "  " + Dim.Render(formatDue(t.Due))
	}
	notes := ""
	if t.Notes != "" {
		notes = "\n    " + Dim.Render(truncate(t.Notes, 80))
	}
	return fmt.Sprintf("%s  %s%s%s", chip, title, right, notes)
}

func formatDue(d time.Time) string {
	today := time.Now()
	t0 := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	d0 := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
	days := int(d0.Sub(t0).Hours() / 24)
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
	default:
		return d.Format("Jan 2")
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
