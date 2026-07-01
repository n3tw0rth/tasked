package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/n3tw0rth/tasked/internal/tui"
)

var rmShowAll bool

var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Delete tasks (inline multi-select)",
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
		ts, err := c.Tasks(ctx, listID, rmShowAll)
		if err != nil {
			return err
		}
		items := make([]tui.PickerItem, 0, len(ts))
		for _, t := range ts {
			items = append(items, tui.PickerItem{
				ID:    t.ID,
				Label: t.Title,
				Chip:  tui.PriorityChip(t.Priority),
			})
		}
		res, err := tui.RunPicker("Delete tasks — space to toggle", items, true)
		if err != nil {
			return err
		}
		if res.Cancel || len(res.Selected) == 0 {
			return nil
		}
		fmt.Printf("%s Delete %d task(s)? type 'yes' to confirm: ", tui.Warn.Render("!"), len(res.Selected))
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "yes" {
			printf("%s\n", tui.Dim.Render("aborted"))
			return nil
		}
		for _, it := range res.Selected {
			if err := c.Delete(ctx, listID, it.ID); err != nil {
				return err
			}
		}
		printf("%s Deleted %d task(s)\n", tui.Ok.Render("✓"), len(res.Selected))
		return nil
	},
}

func init() {
	rmCmd.Flags().BoolVar(&rmShowAll, "all", false, "include completed tasks")
	rootCmd.AddCommand(rmCmd)
}
