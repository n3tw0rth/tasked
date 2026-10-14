package cmd

import (
	"github.com/spf13/cobra"

	"github.com/n3tw0rth/tasked/internal/tasks"
	"github.com/n3tw0rth/tasked/internal/tui"
)

var reopenCmd = &cobra.Command{
	Use:     "reopen [task...]",
	Aliases: []string{"undone"},
	Short:   "Mark completed tasks as not done (inline multi-select, or by title/ID)",
	Long:    "Mark completed tasks as not done.\n\n" + taskArgsHelp,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, c, tl, err := openList(cmd)
		if err != nil {
			return err
		}
		all, err := c.Tasks(ctx, tl.ID, true)
		if err != nil {
			return err
		}
		var completed []tasks.Task
		for _, t := range all {
			if t.Done {
				completed = append(completed, t)
			}
		}
		sel, err := selectTasks(completed, args, "Reopen — space to toggle", true)
		if err != nil || len(sel) == 0 {
			return err
		}
		for _, t := range sel {
			if err := c.Reopen(ctx, tl.ID, t.ID); err != nil {
				return err
			}
		}
		printf("%s Reopened %d task(s)\n", tui.Ok.Render("✓"), len(sel))
		return nil
	},
}

func init() {
	addListFlag(reopenCmd)
	rootCmd.AddCommand(reopenCmd)
}
