package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/n3tw0rth/tasked/internal/tui"
)

var searchAll bool

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search tasks by title/notes in the active list",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, _, prof, err := loadCtx(cmd)
		if err != nil {
			return err
		}
		listID, listTitle, err := requireActiveList(prof)
		if err != nil {
			return err
		}
		c, err := newTasksClient(ctx, prof)
		if err != nil {
			return err
		}
		query := args[0]
		for _, extra := range args[1:] {
			query += " " + extra
		}
		results, err := c.Search(ctx, listID, query, searchAll)
		if err != nil {
			return err
		}
		fmt.Print(tui.RenderTasks(fmt.Sprintf("%s — %q", listTitle, query), results))
		return nil
	},
}

func init() {
	searchCmd.Flags().BoolVar(&searchAll, "all", false, "include completed tasks")
	rootCmd.AddCommand(searchCmd)
}
