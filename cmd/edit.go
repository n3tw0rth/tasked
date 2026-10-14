package cmd

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"

	"github.com/n3tw0rth/tasked/internal/tasks"
	"github.com/n3tw0rth/tasked/internal/tui"
)

var (
	editTitle    string
	editNotes    string
	editDue      string
	editNoDue    bool
	editPriority string
)

var editCmd = &cobra.Command{
	Use:   "edit [task]",
	Short: "Edit a task's title, notes, due date, or priority",
	Long: `Edit one task. With flags, only those fields change; without flags, an
inline form opens prefilled with the task.

` + taskArgsHelp,
	Example: `  tasked edit
  tasked edit "Pay rent" --due 2026-08-01
  tasked edit milk --title "Buy oat milk" --no-due`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		u, err := editFlagsUpdate(cmd)
		if err != nil {
			return err
		}
		ctx, c, tl, err := openList(cmd)
		if err != nil {
			return err
		}
		ts, err := c.Tasks(ctx, tl.ID, false)
		if err != nil {
			return err
		}
		sel, err := selectTasks(ts, args, "Edit — pick a task", false)
		if err != nil || len(sel) == 0 {
			return err
		}
		t := sel[0]
		if u.IsZero() {
			if err := requireTTY(); err != nil {
				return err
			}
			var ok bool
			u, ok, err = editFormUpdate(t)
			if err != nil || !ok {
				return err
			}
		}
		if u.IsZero() {
			printf("%s\n", tui.Dim.Render("nothing changed"))
			return nil
		}
		updated, err := c.Update(ctx, tl.ID, t.ID, u)
		if err != nil {
			return err
		}
		printf("%s Updated %s  %s\n", tui.Ok.Render("✓"), tui.PriorityChip(updated.Priority), updated.Title)
		return nil
	},
}

// editFlagsUpdate builds an update from the flags that were set, validating
// them before any network calls.
func editFlagsUpdate(cmd *cobra.Command) (tasks.TaskUpdate, error) {
	var u tasks.TaskUpdate
	f := cmd.Flags()
	if f.Changed("title") {
		t := strings.TrimSpace(editTitle)
		if t == "" {
			return u, errors.New("--title can't be empty")
		}
		u.Title = &t
	}
	if f.Changed("notes") {
		u.Notes = &editNotes
	}
	if f.Changed("priority") {
		p, err := parsePriorityArg(editPriority)
		if err != nil {
			return u, err
		}
		u.Priority = &p
	}
	if f.Changed("due") && editNoDue {
		return u, errors.New("--due and --no-due can't be used together")
	}
	if f.Changed("due") {
		d, err := tasks.ParseDueInput(editDue)
		if err != nil {
			return u, err
		}
		u.Due = &d
	}
	u.ClearDue = editNoDue
	return u, nil
}

// editFormUpdate opens the task form prefilled with t and returns only the
// fields the user changed. ok is false if the form was cancelled.
func editFormUpdate(t tasks.Task) (u tasks.TaskUpdate, ok bool, err error) {
	init := tui.FormResult{Title: t.Title, Notes: t.Notes, Priority: t.Priority}
	if t.HasDue {
		init.Due = t.Due.Format("2006-01-02")
	}
	res, err := tui.RunForm("Edit task", init)
	if err != nil || res.Cancel {
		return u, false, err
	}
	if res.Title != t.Title {
		u.Title = &res.Title
	}
	if res.NotesEdited {
		u.Notes = &res.Notes
	}
	if res.Priority != t.Priority {
		u.Priority = &res.Priority
	}
	if res.Due != init.Due {
		if res.Due == "" {
			u.ClearDue = true
		} else {
			d, err := tasks.ParseDueInput(res.Due)
			if err != nil {
				return u, false, err
			}
			u.Due = &d
		}
	}
	return u, true, nil
}

func init() {
	editCmd.Flags().StringVar(&editTitle, "title", "", "new title")
	editCmd.Flags().StringVarP(&editNotes, "notes", "n", "", "new notes (empty to clear)")
	editCmd.Flags().StringVarP(&editDue, "due", "d", "", "new due date: tmr, fri, 3d, 2w, eow, YYYY-MM-DD, ...")
	editCmd.Flags().BoolVar(&editNoDue, "no-due", false, "remove the due date")
	editCmd.Flags().StringVarP(&editPriority, "priority", "p", "", "new priority p1..p5")
	addListFlag(editCmd)
	rootCmd.AddCommand(editCmd)
}
