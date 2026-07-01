package cmd

import (
	"github.com/spf13/cobra"

	"github.com/n3tw0rth/tasked/internal/tasks"
	"github.com/n3tw0rth/tasked/internal/tui"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Create a task via an inline form (title, due, priority)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, _, prof, err := loadCtx(cmd)
		if err != nil {
			return err
		}
		listID, _, err := requireActiveList(prof)
		if err != nil {
			return err
		}
		c, err := newTasksClient(ctx, prof)
		if err != nil {
			return err
		}
		res, err := tui.RunForm()
		if err != nil {
			return err
		}
		if res.Cancel {
			return nil
		}
		in := tasks.Task{
			Title:    res.Title,
			Priority: res.Priority,
		}
		if res.Due != "" {
			due, err := tasks.ParseDueInput(res.Due)
			if err != nil {
				return err
			}
			in.Due = due
			in.HasDue = true
		}
		created, err := c.Create(ctx, listID, in)
		if err != nil {
			return err
		}
		printf("%s Created %s  %s\n", tui.Ok.Render("✓"), tui.PriorityChip(created.Priority), created.Title)
		return nil
	},
}

func init() { rootCmd.AddCommand(addCmd) }
