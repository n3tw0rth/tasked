package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/n3tw0rth/tasked/internal/tui"
)

var (
	searchAll  bool
	searchJSON bool
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search tasks by title/notes in the active list",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, c, tl, err := openList(cmd)
		if err != nil {
			return err
		}
		query := strings.Join(args, " ")
		results, err := c.Search(ctx, tl.ID, query, searchAll)
		if err != nil {
			return err
		}
		if searchJSON {
			return printTasksJSON(results)
		}
		fmt.Print(tui.RenderTasks(fmt.Sprintf("%s — %q", tl.Title, query), results))
		return nil
	},
}

func init() {
	searchCmd.Flags().BoolVar(&searchAll, "all", false, "include completed tasks")
	searchCmd.Flags().BoolVar(&searchJSON, "json", false, "print tasks as JSON")
	addListFlag(searchCmd)
	rootCmd.AddCommand(searchCmd)
}
