package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/n3tw0rth/tasked/internal/tasks"
	"github.com/n3tw0rth/tasked/internal/tui"
)

var (
	addDue      string
	addPriority string
	addNotes    string
)

var addCmd = &cobra.Command{
	Use:   "add [title...]",
	Short: "Create a task (inline form, or from arguments and flags)",
	Long: `Create a task. With a title as arguments the task is created directly;
without one, an inline form opens, prefilled from any flags given.`,
	Example: `  tasked add
  tasked add Pay rent --due fri -p 1
  tasked add "Call the bank" -d tmr -n "ask about the card"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		in := tasks.Task{
			Title:    strings.TrimSpace(strings.Join(args, " ")),
			Notes:    addNotes,
			Priority: tasks.Default,
		}
		if addPriority != "" {
			p, err := parsePriorityArg(addPriority)
			if err != nil {
				return err
			}
			in.Priority = p
		}
		due := addDue
		if in.Title != "" && due != "" {
			// Validate before any network calls.
			if _, err := tasks.ParseDueInput(due); err != nil {
				return err
			}
		}
		if in.Title == "" {
			if err := requireTTY(); err != nil {
				return err
			}
		}
		ctx, c, tl, err := openList(cmd)
		if err != nil {
			return err
		}
		if in.Title == "" {
			res, err := tui.RunForm("New task", tui.FormResult{Notes: in.Notes, Due: due, Priority: in.Priority})
			if err != nil {
				return err
			}
			if res.Cancel {
				return nil
			}
			in.Title, in.Notes, in.Priority, due = res.Title, res.Notes, res.Priority, res.Due
		}
		if due != "" {
			d, err := tasks.ParseDueInput(due)
			if err != nil {
				return err
			}
			in.Due, in.HasDue = d, true
		}
		created, err := c.Create(ctx, tl.ID, in)
		if err != nil {
			return err
		}
		printf("%s Created %s  %s %s\n", tui.Ok.Render("✓"), tui.PriorityChip(created.Priority), created.Title, tui.Dim.Render("in "+tl.Title))
		return nil
	},
}

func init() {
	addCmd.Flags().StringVarP(&addDue, "due", "d", "", "due date: tmr, fri, 3d, 2w, eow, YYYY-MM-DD, ...")
	addCmd.Flags().StringVarP(&addPriority, "priority", "p", "", "priority p1..p5 (default p3)")
	addCmd.Flags().StringVarP(&addNotes, "notes", "n", "", "task notes")
	addListFlag(addCmd)
	rootCmd.AddCommand(addCmd)
}
