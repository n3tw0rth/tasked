package tasks

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Google Tasks stores due dates only: the API keeps the date part of the
// RFC 3339 timestamp it is sent, discards the time, and returns midnight UTC.
// Due dates are therefore handled as calendar dates (local midnight) and sent
// as midnight UTC of that same date, so no timezone offset can shift the day.

// formatAPIDue converts a due date to the API's date-only timestamp.
func formatAPIDue(d time.Time) string {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
}

// parseAPIDue reads the API's midnight-UTC timestamp back as the same
// calendar date at midnight in loc.
func parseAPIDue(s string, loc *time.Location) (time.Time, bool) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, false
	}
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc), true
}

var weekdays = map[string]time.Weekday{
	"sun": time.Sunday, "sunday": time.Sunday,
	"mon": time.Monday, "monday": time.Monday,
	"tue": time.Tuesday, "tues": time.Tuesday, "tuesday": time.Tuesday,
	"wed": time.Wednesday, "weds": time.Wednesday, "wednesday": time.Wednesday,
	"thu": time.Thursday, "thur": time.Thursday, "thurs": time.Thursday, "thursday": time.Thursday,
	"fri": time.Friday, "friday": time.Friday,
	"sat": time.Saturday, "saturday": time.Saturday,
}

// ParseDueInput accepts a range of shorthand and explicit date forms and
// returns the due date at local midnight:
//
//	today, tod                     today
//	tomorrow, tom, tmr             tomorrow
//	eow                            end of this week (upcoming Sunday)
//	mon..sun (or full names)       next occurrence of that weekday
//	+N                             N days from today
//	Nd, Nw                         N days / weeks from today (e.g. 3d, 2w)
//	YYYY-MM-DD                     explicit date
//
// Google Tasks can't store a time of day, so input with a time is rejected
// rather than silently truncated.
func ParseDueInput(s string) (time.Time, error) {
	return parseDueAt(s, time.Now())
}

func parseDueAt(s string, now time.Time) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, errors.New("empty due date")
	}
	loc := now.Location()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	lower := strings.ToLower(s)
	switch lower {
	case "today", "tod":
		return today, nil
	case "tomorrow", "tom", "tmr":
		return today.AddDate(0, 0, 1), nil
	case "eow": // end of week: upcoming Sunday (today if already Sunday)
		days := (int(time.Sunday) - int(now.Weekday()) + 7) % 7
		return today.AddDate(0, 0, days), nil
	}

	// Weekday names -> next occurrence, strictly in the future.
	if wd, ok := weekdays[lower]; ok {
		days := (int(wd) - int(now.Weekday()) + 7) % 7
		if days == 0 {
			days = 7
		}
		return today.AddDate(0, 0, days), nil
	}

	if days, ok := parseRelativeDays(lower); ok {
		return today.AddDate(0, 0, days), nil
	}

	if t, err := time.ParseInLocation("2006-01-02", s, loc); err == nil {
		return t, nil
	}
	for _, l := range []string{"2006-01-02 15:04", "2006-01-02T15:04", time.RFC3339} {
		if _, err := time.Parse(l, s); err == nil {
			return time.Time{}, fmt.Errorf("due times aren't supported (Google Tasks stores dates only) — use %q", s[:10])
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized due date %q (try tmr, fri, 3d, 2w, eow, or YYYY-MM-DD)", s)
}

// parseRelativeDays handles "+N", "Nd", and "Nw", returning the day offset.
func parseRelativeDays(s string) (int, bool) {
	if strings.HasPrefix(s, "+") {
		n, err := strconv.Atoi(s[1:])
		return n, err == nil
	}
	if len(s) >= 2 {
		unit := s[len(s)-1]
		if unit == 'd' || unit == 'w' {
			if n, err := strconv.Atoi(s[:len(s)-1]); err == nil {
				if unit == 'w' {
					n *= 7
				}
				return n, true
			}
		}
	}
	return 0, false
}
