package tasks

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	Highest = 1
	Lowest  = 5
)

var priorityRE = regexp.MustCompile(`(?i)\[p([1-5])\]`)

// ParsePriority returns the priority number 1..5 found in notes; if none is
// present, it returns Lowest (5).
func ParsePriority(notes string) int {
	m := priorityRE.FindStringSubmatch(notes)
	if len(m) < 2 {
		return Lowest
	}
	switch m[1] {
	case "1":
		return 1
	case "2":
		return 2
	case "3":
		return 3
	case "4":
		return 4
	default:
		return 5
	}
}

// StripPriority removes the [pN] token (case-insensitive) and any single
// leading/trailing whitespace that surrounded it.
func StripPriority(notes string) string {
	stripped := priorityRE.ReplaceAllString(notes, "")
	return strings.TrimSpace(stripped)
}

// UpsertPriority strips any existing [pN] token and prepends a fresh one.
// If p is outside 1..5, it clamps to the allowed range.
func UpsertPriority(notes string, p int) string {
	if p < Highest {
		p = Highest
	}
	if p > Lowest {
		p = Lowest
	}
	rest := StripPriority(notes)
	if rest == "" {
		return fmt.Sprintf("[p%d]", p)
	}
	return fmt.Sprintf("[p%d] %s", p, rest)
}
