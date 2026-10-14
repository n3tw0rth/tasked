package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/n3tw0rth/tasked/internal/tasks"
	"github.com/n3tw0rth/tasked/internal/tui"
)

var (
	lsShowAll  bool
	lsPriority string
	lsJSON     bool
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List tasks sorted by priority",
	RunE: func(cmd *cobra.Command, args []string) error {
		var prio int
		if lsPriority != "" {
			p, err := parsePriorityArg(lsPriority)
			if err != nil {
				return err
			}
			prio = p
		}
		ctx, c, tl, err := openList(cmd)
		if err != nil {
			return err
		}
		items, err := c.Tasks(ctx, tl.ID, lsShowAll)
		if err != nil {
			return err
		}
		if prio != 0 {
			items = filterPriority(items, prio)
		}
		if lsJSON {
			return printTasksJSON(items)
		}
		fmt.Print(tui.RenderTasks(tl.Title, items))
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

type taskJSON struct {
	ID       string `json:"id"`
	Parent   string `json:"parent,omitempty"` // parent task ID for subtasks
	Title    string `json:"title"`
	Notes    string `json:"notes,omitempty"`
	Priority int    `json:"priority"`
	Due      string `json:"due,omitempty"` // YYYY-MM-DD
	Done     bool   `json:"done"`
}

func printTasksJSON(ts []tasks.Task) error {
	out := make([]taskJSON, 0, len(ts))
	for _, t := range ts {
		j := taskJSON{ID: t.ID, Parent: t.Parent, Title: t.Title, Notes: t.Notes, Priority: t.Priority, Done: t.Done}
		if t.HasDue {
			j.Due = t.Due.Format("2006-01-02")
		}
		out = append(out, j)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func init() {
	lsCmd.Flags().BoolVar(&lsShowAll, "all", false, "include completed tasks")
	lsCmd.Flags().StringVarP(&lsPriority, "priority", "p", "", "filter by priority (p1..p5)")
	lsCmd.Flags().BoolVar(&lsJSON, "json", false, "print tasks as JSON")
	addListFlag(lsCmd)
	rootCmd.AddCommand(lsCmd)
}
