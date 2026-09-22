package tasks

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/option"
	tasksapi "google.golang.org/api/tasks/v1"

	"github.com/n3tw0rth/tasked/internal/auth"
)

type Client struct {
	svc *tasksapi.Service
}

func New(ctx context.Context, profile string) (*Client, error) {
	src, err := auth.TokenSource(ctx, profile)
	if err != nil {
		return nil, err
	}
	svc, err := tasksapi.NewService(ctx, option.WithTokenSource(src))
	if err != nil {
		return nil, err
	}
	return &Client{svc: svc}, nil
}

func (c *Client) Lists(ctx context.Context) ([]TaskList, error) {
	var out []TaskList
	pageToken := ""
	for {
		call := c.svc.Tasklists.List().MaxResults(100).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		res, err := call.Do()
		if err != nil {
			return nil, err
		}
		for _, tl := range res.Items {
			out = append(out, TaskList{ID: tl.Id, Title: tl.Title})
		}
		if res.NextPageToken == "" {
			break
		}
		pageToken = res.NextPageToken
	}
	return out, nil
}

func (c *Client) CreateList(ctx context.Context, title string) (TaskList, error) {
	tl, err := c.svc.Tasklists.Insert(&tasksapi.TaskList{Title: title}).Context(ctx).Do()
	if err != nil {
		return TaskList{}, err
	}
	return TaskList{ID: tl.Id, Title: tl.Title}, nil
}

func (c *Client) DeleteList(ctx context.Context, id string) error {
	return c.svc.Tasklists.Delete(id).Context(ctx).Do()
}

func (c *Client) Tasks(ctx context.Context, listID string, includeCompleted bool) ([]Task, error) {
	var out []Task
	pageToken := ""
	for {
		call := c.svc.Tasks.List(listID).MaxResults(100).ShowCompleted(includeCompleted).ShowHidden(includeCompleted).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		res, err := call.Do()
		if err != nil {
			return nil, err
		}
		for _, t := range res.Items {
			out = append(out, taskFromAPI(t))
		}
		if res.NextPageToken == "" {
			break
		}
		pageToken = res.NextPageToken
	}
	SortByPriority(out)
	return out, nil
}

func (c *Client) Create(ctx context.Context, listID string, in Task) (Task, error) {
	notes := UpsertPriority(in.Notes, in.Priority)
	apiTask := &tasksapi.Task{
		Title: in.Title,
		Notes: notes,
	}
	if in.HasDue {
		apiTask.Due = in.Due.UTC().Format(time.RFC3339)
	}
	res, err := c.svc.Tasks.Insert(listID, apiTask).Context(ctx).Do()
	if err != nil {
		return Task{}, err
	}
	return taskFromAPI(res), nil
}

// SetPriority rewrites the [pN] token in the task's notes, keeping the rest
// of the notes intact. It re-fetches the task first so edits made elsewhere
// since the task was listed aren't clobbered by the notes patch.
func (c *Client) SetPriority(ctx context.Context, listID, taskID string, p int) (Task, error) {
	cur, err := c.svc.Tasks.Get(listID, taskID).Context(ctx).Do()
	if err != nil {
		return Task{}, err
	}
	patch := &tasksapi.Task{Notes: UpsertPriority(cur.Notes, p)}
	res, err := c.svc.Tasks.Patch(listID, taskID, patch).Context(ctx).Do()
	if err != nil {
		return Task{}, err
	}
	return taskFromAPI(res), nil
}

// MoveToList moves a task to another task list. Notes (including the
// priority tag), due date, and status travel with the task.
func (c *Client) MoveToList(ctx context.Context, listID, taskID, destListID string) (Task, error) {
	res, err := c.svc.Tasks.Move(listID, taskID).DestinationTasklist(destListID).Context(ctx).Do()
	if err != nil {
		return Task{}, err
	}
	return taskFromAPI(res), nil
}

func (c *Client) Complete(ctx context.Context, listID, taskID string) error {
	patch := &tasksapi.Task{Status: "completed"}
	_, err := c.svc.Tasks.Patch(listID, taskID, patch).Context(ctx).Do()
	return err
}

func (c *Client) Delete(ctx context.Context, listID, taskID string) error {
	return c.svc.Tasks.Delete(listID, taskID).Context(ctx).Do()
}

func (c *Client) Search(ctx context.Context, listID, query string, includeCompleted bool) ([]Task, error) {
	all, err := c.Tasks(ctx, listID, includeCompleted)
	if err != nil {
		return nil, err
	}
	q := strings.ToLower(query)
	var out []Task
	for _, t := range all {
		if strings.Contains(strings.ToLower(t.Title), q) || strings.Contains(strings.ToLower(t.Notes), q) {
			out = append(out, t)
		}
	}
	return out, nil
}

func taskFromAPI(t *tasksapi.Task) Task {
	raw := t.Notes
	prio := ParsePriority(raw)
	body := StripPriority(raw)
	out := Task{
		ID:       t.Id,
		Title:    t.Title,
		Notes:    body,
		RawNotes: raw,
		Priority: prio,
		Done:     t.Status == "completed",
	}
	if t.Due != "" {
		if due, err := time.Parse(time.RFC3339, t.Due); err == nil {
			out.Due = due
			out.HasDue = true
		}
	}
	return out
}

// SortByPriority sorts in place: p1 first, then earliest due, then title.
func SortByPriority(ts []Task) {
	sort.SliceStable(ts, func(i, j int) bool {
		if ts[i].Priority != ts[j].Priority {
			return ts[i].Priority < ts[j].Priority
		}
		if ts[i].HasDue != ts[j].HasDue {
			return ts[i].HasDue
		}
		if ts[i].HasDue && !ts[i].Due.Equal(ts[j].Due) {
			return ts[i].Due.Before(ts[j].Due)
		}
		return strings.ToLower(ts[i].Title) < strings.ToLower(ts[j].Title)
	})
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
// returns the resulting time in the local timezone:
//
//	today, tod, eod                due end of today
//	tomorrow, tom, tmr             due end of tomorrow
//	eow                            end of this week (upcoming Sunday)
//	mon..sun (or full names)       next occurrence of that weekday
//	+N                             N days from now
//	Nd, Nw                         N days / weeks from now (e.g. 3d, 2w)
//	RFC3339, YYYY-MM-DD,           explicit date / datetime
//	YYYY-MM-DD HH:MM
//
// Shorthand day forms are pinned to 23:59 local time.
func ParseDueInput(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty due")
	}
	now := time.Now()
	loc := now.Location()

	endOfDay := func(t time.Time) time.Time {
		return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 0, 0, loc)
	}

	lower := strings.ToLower(s)
	switch lower {
	case "today", "tod", "eod":
		return endOfDay(now), nil
	case "tomorrow", "tom", "tmr":
		return endOfDay(now.AddDate(0, 0, 1)), nil
	case "eow": // end of week: upcoming Sunday (today if already Sunday)
		days := (int(time.Sunday) - int(now.Weekday()) + 7) % 7
		return endOfDay(now.AddDate(0, 0, days)), nil
	}

	// Weekday names -> next occurrence, strictly in the future.
	if wd, ok := weekdays[lower]; ok {
		days := (int(wd) - int(now.Weekday()) + 7) % 7
		if days == 0 {
			days = 7
		}
		return endOfDay(now.AddDate(0, 0, days)), nil
	}

	// Offset forms: +N (days), Nd (days), Nw (weeks).
	if t, ok := parseRelativeDue(lower, now, endOfDay); ok {
		return t, nil
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02",
	}
	for _, l := range layouts {
		if t, err := time.ParseInLocation(l, s, loc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized due date: %q", s)
}

// parseRelativeDue handles "+N", "Nd", and "Nw". The bool reports whether s
// matched a relative form (regardless of a leading zero-day offset).
func parseRelativeDue(s string, now time.Time, endOfDay func(time.Time) time.Time) (time.Time, bool) {
	if strings.HasPrefix(s, "+") {
		if n, err := strconv.Atoi(s[1:]); err == nil {
			return endOfDay(now.AddDate(0, 0, n)), true
		}
		return time.Time{}, false
	}
	if len(s) >= 2 {
		unit := s[len(s)-1]
		if unit == 'd' || unit == 'w' {
			if n, err := strconv.Atoi(s[:len(s)-1]); err == nil {
				days := n
				if unit == 'w' {
					days = 7 * n
				}
				return endOfDay(now.AddDate(0, 0, days)), true
			}
		}
	}
	return time.Time{}, false
}
