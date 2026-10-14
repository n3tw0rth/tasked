package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/n3tw0rth/tasked/internal/tui"
)

var (
	rmShowAll bool
	rmYes     bool
)

var rmCmd = &cobra.Command{
	Use:   "rm [task...]",
	Short: "Delete tasks (inline multi-select, or by title/ID)",
	Long:  "Delete tasks. Asks for confirmation unless --yes is given.\n\n" + taskArgsHelp,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, c, tl, err := openList(cmd)
		if err != nil {
			return err
		}
		ts, err := c.Tasks(ctx, tl.ID, rmShowAll)
		if err != nil {
			return err
		}
		sel, err := selectTasks(ts, args, "Delete tasks — space to toggle", true)
		if err != nil || len(sel) == 0 {
			return err
		}
		if !rmYes {
			if requireTTY() != nil {
				return errors.New("refusing to delete without confirmation — pass --yes")
			}
			for _, t := range sel {
				printf("  %s  %s\n", tui.PriorityChip(t.Priority), t.Title)
			}
			fmt.Printf("%s Delete %d task(s)? type 'yes' to confirm: ", tui.Warn.Render("!"), len(sel))
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "yes" {
				printf("%s\n", tui.Dim.Render("aborted"))
				return nil
			}
		}
		for _, t := range sel {
			if err := c.Delete(ctx, tl.ID, t.ID); err != nil {
				return err
			}
		}
		printf("%s Deleted %d task(s)\n", tui.Ok.Render("✓"), len(sel))
		return nil
	},
}

func init() {
	rmCmd.Flags().BoolVar(&rmShowAll, "all", false, "include completed tasks")
	rmCmd.Flags().BoolVarP(&rmYes, "yes", "y", false, "skip the confirmation prompt")
	addListFlag(rmCmd)
	rootCmd.AddCommand(rmCmd)
}
