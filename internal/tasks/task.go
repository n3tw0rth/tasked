package tasks

import "time"

type Task struct {
	ID       string
	Parent   string // parent task ID for subtasks, empty for top-level tasks
	Title    string
	Notes    string    // priority-stripped body for display
	Priority int       // 1..5
	Due      time.Time // due date at local midnight (Google Tasks has no due times)
	HasDue   bool
	Done     bool
}

type TaskList struct {
	ID    string
	Title string
}

// TaskUpdate lists the fields to change on a task; nil fields are left as-is.
type TaskUpdate struct {
	Title    *string
	Notes    *string // body without the priority tag
	Priority *int
	Due      *time.Time
	ClearDue bool
}

func (u TaskUpdate) IsZero() bool {
	return u.Title == nil && u.Notes == nil && u.Priority == nil && u.Due == nil && !u.ClearDue
}
