package cmd

import (
	"strings"
	"testing"

	"github.com/n3tw0rth/tasked/internal/tasks"
)

var sample = []tasks.Task{
	{ID: "id1", Title: "Pay rent"},
	{ID: "id2", Title: "Buy milk"},
	{ID: "id3", Title: "Buy oat milk"},
	{ID: "id4", Title: "Rent car"},
	{ID: "id5", Title: "Milk"},
}

func TestMatchTask(t *testing.T) {
	tests := map[string]string{
		"id2":      "id2",
		"pay rent": "id1", // exact, case-insensitive
		"milk":     "id5", // exact title beats substring matches
		"oat":      "id3", // unique substring
		" id4 ":    "id4",
	}
	for sel, want := range tests {
		got, err := matchTask(sample, sel)
		if err != nil || got.ID != want {
			t.Errorf("matchTask(%q) = %q, %v; want %q", sel, got.ID, err, want)
		}
	}
}

func TestMatchTaskErrors(t *testing.T) {
	tests := map[string]string{
		"rent": "matches 2 tasks",
		"buy":  "matches 2 tasks",
		"nope": "no task matches",
		"  ":   "empty",
	}
	for sel, want := range tests {
		_, err := matchTask(sample, sel)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("matchTask(%q) error = %v, want it to mention %q", sel, err, want)
		}
	}
}

func TestMatchTasksDedupes(t *testing.T) {
	got, err := matchTasks(sample, []string{"id1", "Pay rent", "oat"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ID != "id1" || got[1].ID != "id3" {
		t.Errorf("matchTasks = %+v, want id1, id3", got)
	}
	if _, err := matchTasks(sample, []string{"id1", "nope"}); err == nil {
		t.Error("matchTasks ignored an unmatched selector")
	}
}

func TestFindList(t *testing.T) {
	lists := []tasks.TaskList{{ID: "L1", Title: "My Tasks"}, {ID: "L2", Title: "Work"}}
	for q, want := range map[string]string{"L2": "L2", "work": "L2", " my tasks ": "L1"} {
		if tl, ok := findList(lists, q); !ok || tl.ID != want {
			t.Errorf("findList(%q) = %+v, %v; want %s", q, tl, ok, want)
		}
	}
	if _, ok := findList(lists, "Home"); ok {
		t.Error("findList matched a missing list")
	}
}

func TestParsePriorityArg(t *testing.T) {
	for in, want := range map[string]int{"p1": 1, "P2": 2, " 3 ": 3, "5": 5} {
		if got, err := parsePriorityArg(in); err != nil || got != want {
			t.Errorf("parsePriorityArg(%q) = %d, %v; want %d", in, got, err, want)
		}
	}
	for _, in := range []string{"p0", "p6", "x", ""} {
		if _, err := parsePriorityArg(in); err == nil {
			t.Errorf("parsePriorityArg(%q) accepted", in)
		}
	}
}
