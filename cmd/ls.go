package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/n3tw0rth/tasked/internal/tasks"
	"github.com/n3tw0rth/tasked/internal/tui"
)

var (
	lsShowAll     bool
	lsPriority    string
	lsOverrideID  string
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List tasks sorted by priority",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, _, prof, err := loadCtx(cmd)
		if err != nil {
			return err
		}
		listID := lsOverrideID
		listTitle := ""
		if listID == "" {
			id, title, err := requireActiveList(prof)
			if err != nil {
				return err
			}
			listID = id
			listTitle = title
		}
		c, err := newTasksClient(ctx, prof)
		if err != nil {
			return err
		}
		items, err := c.Tasks(ctx, listID, lsShowAll)
		if err != nil {
			return err
		}
		if lsPriority != "" {
			p, err := parsePriorityFlag(lsPriority)
			if err != nil {
				return err
			}
			items = filterPriority(items, p)
		}
		header := listTitle
		if header == "" {
			header = listID
		}
		fmt.Print(tui.RenderTasks(header, items))
		return nil
	},
}

func filterPriority(ts []tasks.Task, p int) []tasks.Task {
	var out []tasks.Task
	for _, t := range ts {
		if t.Priority == p {
			out = append(out, t)
		}
	}
	return out
}

func parsePriorityFlag(s string) (int, error) {
	switch s {
	case "1", "p1":
		return 1, nil
	case "2", "p2":
		return 2, nil
	case "3", "p3":
		return 3, nil
	case "4", "p4":
		return 4, nil
	case "5", "p5":
		return 5, nil
	}
	return 0, fmt.Errorf("invalid --priority %q (expected p1..p5)", s)
}

func init() {
	lsCmd.Flags().BoolVar(&lsShowAll, "all", false, "include completed tasks")
	lsCmd.Flags().StringVar(&lsPriority, "priority", "", "filter by priority (p1..p5)")
	lsCmd.Flags().StringVar(&lsOverrideID, "list", "", "override the active list ID")
	rootCmd.AddCommand(lsCmd)
}
