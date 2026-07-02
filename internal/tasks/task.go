package tasks

import "time"

type Task struct {
	ID       string
	Title    string
	Notes    string // priority-stripped body for display
	RawNotes string // exactly what Google Tasks stores
	Priority int    // 1..5
	Due      time.Time
	HasDue   bool
	Done     bool
}

type TaskList struct {
	ID    string
	Title string
}
