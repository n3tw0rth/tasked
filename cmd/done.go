package cmd

import (
	"github.com/spf13/cobra"

	"github.com/n3tw0rth/tasked/internal/tui"
)

var doneCmd = &cobra.Command{
	Use:   "done",
	Short: "Mark tasks completed (inline multi-select)",
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
		ts, err := c.Tasks(ctx, listID, false)
		if err != nil {
			return err
		}
		items := make([]tui.PickerItem, 0, len(ts))
		for _, t := range ts {
			extra := ""
			if t.HasDue {
				extra = t.Due.Format("Jan 2")
			}
			items = append(items, tui.PickerItem{
				ID:    t.ID,
				Label: t.Title,
				Chip:  tui.PriorityChip(t.Priority),
				Extra: extra,
			})
		}
		res, err := tui.RunPicker("Mark as done — space to toggle", items, true)
		if err != nil {
			return err
		}
		if res.Cancel || len(res.Selected) == 0 {
			return nil
		}
		for _, it := range res.Selected {
			if err := c.Complete(ctx, listID, it.ID); err != nil {
				return err
			}
		}
		printf("%s Completed %d task(s)\n", tui.Ok.Render("✓"), len(res.Selected))
		return nil
	},
}

func init() { rootCmd.AddCommand(doneCmd) }
