package tasks

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	Highest = 1
	Lowest  = 5
	// Default is the priority of new tasks and of untagged tasks (e.g. ones
	// created in the Google Tasks apps), so those aren't buried at the bottom.
	Default = 3
)

var priorityRE = regexp.MustCompile(`(?i)\[p([1-5])\]`)

// HasPriority reports whether notes carry a [pN] tag.
func HasPriority(notes string) bool {
	return priorityRE.MatchString(notes)
}

// ParsePriority returns the priority number 1..5 from the first [pN] tag in
// notes; if none is present, it returns Default (3).
func ParsePriority(notes string) int {
	m := priorityRE.FindStringSubmatch(notes)
	if len(m) < 2 {
		return Default
	}
	return int(m[1][0] - '0')
}

// StripPriority removes the first [pN] tag (the one ParsePriority reads) and
// a single space after it, then trims the result. Any other text, including
// later [pN] tags, is left alone.
func StripPriority(notes string) string {
	loc := priorityRE.FindStringIndex(notes)
	if loc == nil {
		return strings.TrimSpace(notes)
	}
	end := loc[1]
	if end < len(notes) && notes[end] == ' ' {
		end++
	}
	return strings.TrimSpace(notes[:loc[0]] + notes[end:])
}

// UpsertPriority strips the existing [pN] tag and prepends a fresh one.
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
