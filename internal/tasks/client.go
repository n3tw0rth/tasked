package tasks

import (
	"context"
	"sort"
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
	// Fetch (and if needed refresh) the token up front: once it is inside the
	// API client's transport, an expired or revoked sign-in surfaces as an
	// opaque wrapped error instead of auth.ReauthError.
	if _, err := src.Token(); err != nil {
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

// DefaultList returns the account's default task list ("My Tasks").
func (c *Client) DefaultList(ctx context.Context) (TaskList, error) {
	tl, err := c.svc.Tasklists.Get("@default").Context(ctx).Do()
	if err != nil {
		return TaskList{}, err
	}
	return TaskList{ID: tl.Id, Title: tl.Title}, nil
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
		apiTask.Due = formatAPIDue(in.Due)
	}
	res, err := c.svc.Tasks.Insert(listID, apiTask).Context(ctx).Do()
	if err != nil {
		return Task{}, err
	}
	return taskFromAPI(res), nil
}

// Update patches the fields set in u. It re-fetches the task first so the
// notes patch (which carries the priority tag) doesn't clobber edits made
// elsewhere since the task was listed. A task without a priority tag only
// gains one when u sets a priority.
func (c *Client) Update(ctx context.Context, listID, taskID string, u TaskUpdate) (Task, error) {
	cur, err := c.svc.Tasks.Get(listID, taskID).Context(ctx).Do()
	if err != nil {
		return Task{}, err
	}
	patch := &tasksapi.Task{}
	if u.Title != nil {
		patch.Title = *u.Title
	}
	if u.Notes != nil || u.Priority != nil {
		body := StripPriority(cur.Notes)
		if u.Notes != nil {
			body = strings.TrimSpace(*u.Notes)
		}
		switch {
		case u.Priority != nil:
			patch.Notes = UpsertPriority(body, *u.Priority)
		case HasPriority(cur.Notes):
			patch.Notes = UpsertPriority(body, ParsePriority(cur.Notes))
		default:
			patch.Notes = body
		}
		if patch.Notes == "" {
			patch.NullFields = append(patch.NullFields, "Notes")
		}
	}
	switch {
	case u.ClearDue:
		patch.NullFields = append(patch.NullFields, "Due")
	case u.Due != nil:
		patch.Due = formatAPIDue(*u.Due)
	}
	res, err := c.svc.Tasks.Patch(listID, taskID, patch).Context(ctx).Do()
	if err != nil {
		return Task{}, err
	}
	return taskFromAPI(res), nil
}

// SetPriority rewrites the [pN] tag in the task's notes, keeping the rest of
// the notes intact.
func (c *Client) SetPriority(ctx context.Context, listID, taskID string, p int) (Task, error) {
	return c.Update(ctx, listID, taskID, TaskUpdate{Priority: &p})
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

// Reopen marks a completed task as not done again.
func (c *Client) Reopen(ctx context.Context, listID, taskID string) error {
	patch := &tasksapi.Task{Status: "needsAction", NullFields: []string{"Completed"}}
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
	out := Task{
		ID:       t.Id,
		Parent:   t.Parent,
		Title:    t.Title,
		Notes:    StripPriority(t.Notes),
		Priority: ParsePriority(t.Notes),
		Done:     t.Status == "completed",
	}
	if t.Due != "" {
		out.Due, out.HasDue = parseAPIDue(t.Due, time.Local)
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
