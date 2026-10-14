package cmd

import (
	"fmt"
	"strings"

	"github.com/n3tw0rth/tasked/internal/tasks"
	"github.com/n3tw0rth/tasked/internal/tui"
)

// taskArgsHelp documents how task arguments are matched (see matchTask).
const taskArgsHelp = `Tasks can be given as arguments: a task ID, an exact title, or a unique
part of a title (case-insensitive). Quote titles that contain spaces.
Without arguments, an inline picker opens. ` + "`tasked ls --json`" + ` shows task IDs.`

// selectTasks resolves the tasks a command acts on: from the arguments if any
// were given, otherwise from an inline picker. An empty result with a nil
// error means there was nothing to pick or the user cancelled.
func selectTasks(ts []tasks.Task, args []string, title string, multi bool) ([]tasks.Task, error) {
	if len(args) > 0 {
		return matchTasks(ts, args)
	}
	if len(ts) == 0 {
		printf("%s\n", tui.Dim.Render("(no tasks)"))
		return nil, nil
	}
	if err := requireTTY(); err != nil {
		return nil, err
	}
	res, err := tui.RunPicker(title, taskPickerItems(ts), multi)
	if err != nil || res.Cancel {
		return nil, err
	}
	byID := make(map[string]tasks.Task, len(ts))
	for _, t := range ts {
		byID[t.ID] = t
	}
	out := make([]tasks.Task, 0, len(res.Selected))
	for _, it := range res.Selected {
		out = append(out, byID[it.ID])
	}
	return out, nil
}

// matchTasks resolves each selector with matchTask, dropping duplicates.
func matchTasks(ts []tasks.Task, selectors []string) ([]tasks.Task, error) {
	var out []tasks.Task
	seen := map[string]bool{}
	for _, sel := range selectors {
		t, err := matchTask(ts, sel)
		if err != nil {
			return nil, err
		}
		if !seen[t.ID] {
			seen[t.ID] = true
			out = append(out, t)
		}
	}
	return out, nil
}

// matchTask finds exactly one task by ID, else exact title, else unique title
// substring (both case-insensitive). Ambiguous or unmatched selectors are
// errors, so a script never acts on the wrong task.
func matchTask(ts []tasks.Task, sel string) (tasks.Task, error) {
	q := strings.TrimSpace(sel)
	if q == "" {
		return tasks.Task{}, fmt.Errorf("empty task argument")
	}
	for _, t := range ts {
		if t.ID == q {
			return t, nil
		}
	}
	lq := strings.ToLower(q)
	var exact, partial []tasks.Task
	for _, t := range ts {
		lt := strings.ToLower(t.Title)
		switch {
		case lt == lq:
			exact = append(exact, t)
		case strings.Contains(lt, lq):
			partial = append(partial, t)
		}
	}
	candidates := exact
	if len(candidates) == 0 {
		candidates = partial
	}
	switch len(candidates) {
	case 0:
		return tasks.Task{}, fmt.Errorf("no task matches %q", sel)
	case 1:
		return candidates[0], nil
	}
	var titles []string
	for i, t := range candidates {
		if i == 3 {
			titles = append(titles, "…")
			break
		}
		titles = append(titles, fmt.Sprintf("%q", t.Title))
	}
	return tasks.Task{}, fmt.Errorf("%q matches %d tasks (%s) — be more specific or use the task ID",
		sel, len(candidates), strings.Join(titles, ", "))
}

func taskPickerItems(ts []tasks.Task) []tui.PickerItem {
	items := make([]tui.PickerItem, 0, len(ts))
	for _, t := range ts {
		extra := ""
		if t.HasDue {
			extra = tui.FormatDue(t.Due)
		}
		items = append(items, tui.PickerItem{
			ID:    t.ID,
			Label: t.Title,
			Chip:  tui.PriorityChip(t.Priority),
			Extra: extra,
		})
	}
	return items
}
