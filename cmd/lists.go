package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/n3tw0rth/tasked/internal/tasks"
	"github.com/n3tw0rth/tasked/internal/tui"
)

var listsRmYes bool

var listsCmd = &cobra.Command{
	Use:   "lists",
	Short: "Manage Google task lists for the active profile",
}

var listsLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Show task lists",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, _, prof, err := loadCtx(cmd)
		if err != nil {
			return err
		}
		c, err := newTasksClient(ctx, prof)
		if err != nil {
			return err
		}
		ls, err := c.Lists(ctx)
		if err != nil {
			return err
		}
		if len(ls) == 0 {
			printf("%s\n", tui.Dim.Render("(no task lists — create one with `tasked lists create <name>`)"))
			return nil
		}
		for _, l := range ls {
			marker := "  "
			if l.ID == prof.ActiveListID {
				marker = tui.Selected.Render("* ")
			}
			printf("%s%s  %s\n", marker, l.Title, tui.Dim.Render(l.ID))
		}
		return nil
	},
}

var listsCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new task list",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, _, prof, err := loadCtx(cmd)
		if err != nil {
			return err
		}
		c, err := newTasksClient(ctx, prof)
		if err != nil {
			return err
		}
		l, err := c.CreateList(ctx, strings.Join(args, " "))
		if err != nil {
			return err
		}
		printf("%s Created list %q (%s)\n", tui.Ok.Render("✓"), l.Title, l.ID)
		return nil
	},
}

var listsSwitchCmd = &cobra.Command{
	Use:   "switch [list]",
	Short: "Set the active task list (inline picker, or by name/ID)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cfg, prof, err := loadCtx(cmd)
		if err != nil {
			return err
		}
		c, err := newTasksClient(ctx, prof)
		if err != nil {
			return err
		}
		chosen, ok, err := chooseList(ctx, c, args, "Select active list")
		if err != nil || !ok {
			return err
		}
		prof.ActiveListID = chosen.ID
		prof.ActiveListTitle = chosen.Title
		cfg.UpsertProfile(prof)
		if err := cfg.Save(); err != nil {
			return err
		}
		printf("%s Active list: %s\n", tui.Ok.Render("✓"), chosen.Title)
		return nil
	},
}

var listsRmCmd = &cobra.Command{
	Use:   "rm [list]",
	Short: "Delete a task list (inline picker, or by name/ID)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cfg, prof, err := loadCtx(cmd)
		if err != nil {
			return err
		}
		c, err := newTasksClient(ctx, prof)
		if err != nil {
			return err
		}
		chosen, ok, err := chooseList(ctx, c, args, "Delete list — pick one (esc to cancel)")
		if err != nil || !ok {
			return err
		}
		if !listsRmYes {
			if requireTTY() != nil {
				return errors.New("refusing to delete without confirmation — pass --yes")
			}
			fmt.Printf("%s about to delete list %q and all its tasks — type 'yes' to confirm: ", tui.Warn.Render("!"), chosen.Title)
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "yes" {
				printf("%s\n", tui.Dim.Render("aborted"))
				return nil
			}
		}
		if err := c.DeleteList(ctx, chosen.ID); err != nil {
			return err
		}
		if prof.ActiveListID == chosen.ID {
			prof.ActiveListID = ""
			prof.ActiveListTitle = ""
			cfg.UpsertProfile(prof)
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("list deleted, but saving config failed: %w", err)
			}
		}
		printf("%s Deleted list %q\n", tui.Ok.Render("✓"), chosen.Title)
		return nil
	},
}

// chooseList resolves a list from the argument (name or ID) if given,
// otherwise from an inline picker. ok is false if the user cancelled.
func chooseList(ctx context.Context, c *tasks.Client, args []string, title string) (tasks.TaskList, bool, error) {
	ls, err := c.Lists(ctx)
	if err != nil {
		return tasks.TaskList{}, false, err
	}
	if len(args) == 1 {
		tl, ok := findList(ls, args[0])
		if !ok {
			return tasks.TaskList{}, false, fmt.Errorf("no task list matches %q", args[0])
		}
		return tl, true, nil
	}
	if len(ls) == 0 {
		return tasks.TaskList{}, false, errors.New("no task lists")
	}
	if err := requireTTY(); err != nil {
		return tasks.TaskList{}, false, err
	}
	items := make([]tui.PickerItem, 0, len(ls))
	for _, l := range ls {
		items = append(items, tui.PickerItem{ID: l.ID, Label: l.Title})
	}
	res, err := tui.RunPicker(title, items, false)
	if err != nil || res.Cancel || len(res.Selected) == 0 {
		return tasks.TaskList{}, false, err
	}
	return tasks.TaskList{ID: res.Selected[0].ID, Title: res.Selected[0].Label}, true, nil
}

func init() {
	listsRmCmd.Flags().BoolVarP(&listsRmYes, "yes", "y", false, "skip the confirmation prompt")
	listsCmd.AddCommand(listsLsCmd, listsCreateCmd, listsSwitchCmd, listsRmCmd)
	rootCmd.AddCommand(listsCmd)
}
