package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/n3tw0rth/tasked/internal/tui"
)

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
		title := args[0]
		for _, extra := range args[1:] {
			title += " " + extra
		}
		l, err := c.CreateList(ctx, title)
		if err != nil {
			return err
		}
		printf("%s Created list %q (%s)\n", tui.Ok.Render("✓"), l.Title, l.ID)
		return nil
	},
}

var listsSwitchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Pick the active task list (inline picker)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cfg, prof, err := loadCtx(cmd)
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
			return errors.New("no task lists to switch to")
		}
		items := make([]tui.PickerItem, 0, len(ls))
		for _, l := range ls {
			items = append(items, tui.PickerItem{ID: l.ID, Label: l.Title})
		}
		res, err := tui.RunPicker("Select active list", items, false)
		if err != nil {
			return err
		}
		if res.Cancel || len(res.Selected) == 0 {
			return nil
		}
		chosen := res.Selected[0]
		prof.ActiveListID = chosen.ID
		prof.ActiveListTitle = chosen.Label
		cfg.UpsertProfile(prof)
		if err := cfg.Save(); err != nil {
			return err
		}
		printf("%s Active list: %s\n", tui.Ok.Render("✓"), chosen.Label)
		return nil
	},
}

var listsRmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Delete a task list (inline picker)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cfg, prof, err := loadCtx(cmd)
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
			return errors.New("no task lists to remove")
		}
		items := make([]tui.PickerItem, 0, len(ls))
		for _, l := range ls {
			items = append(items, tui.PickerItem{ID: l.ID, Label: l.Title})
		}
		res, err := tui.RunPicker("Delete list — pick one (esc to cancel)", items, false)
		if err != nil {
			return err
		}
		if res.Cancel || len(res.Selected) == 0 {
			return nil
		}
		chosen := res.Selected[0]
		fmt.Printf("%s about to delete list %q — type 'yes' to confirm: ", tui.Warn.Render("!"), chosen.Label)
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "yes" {
			printf("%s\n", tui.Dim.Render("aborted"))
			return nil
		}
		if err := c.DeleteList(ctx, chosen.ID); err != nil {
			return err
		}
		if prof.ActiveListID == chosen.ID {
			prof.ActiveListID = ""
			prof.ActiveListTitle = ""
			cfg.UpsertProfile(prof)
			_ = cfg.Save()
		}
		printf("%s Deleted list %q\n", tui.Ok.Render("✓"), chosen.Label)
		return nil
	},
}

func init() {
	listsCmd.AddCommand(listsLsCmd, listsCreateCmd, listsSwitchCmd, listsRmCmd)
	rootCmd.AddCommand(listsCmd)
}
