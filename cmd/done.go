package cmd

import (
	"github.com/spf13/cobra"

	"github.com/n3tw0rth/tasked/internal/tui"
)

var doneCmd = &cobra.Command{
	Use:   "done [task...]",
	Short: "Mark tasks completed (inline multi-select, or by title/ID)",
	Long:  "Mark tasks completed.\n\n" + taskArgsHelp,
	Example: `  tasked done
  tasked done "Pay rent" milk`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, c, tl, err := openList(cmd)
		if err != nil {
			return err
		}
		ts, err := c.Tasks(ctx, tl.ID, false)
		if err != nil {
			return err
		}
		sel, err := selectTasks(ts, args, "Mark as done — space to toggle", true)
		if err != nil || len(sel) == 0 {
			return err
		}
		for _, t := range sel {
			if err := c.Complete(ctx, tl.ID, t.ID); err != nil {
				return err
			}
		}
		printf("%s Completed %d task(s)\n", tui.Ok.Render("✓"), len(sel))
		return nil
	},
}

func init() {
	addListFlag(doneCmd)
	rootCmd.AddCommand(doneCmd)
}
