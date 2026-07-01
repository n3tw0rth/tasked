package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/n3tw0rth/tasked/internal/auth"
	"github.com/n3tw0rth/tasked/internal/config"
	"github.com/n3tw0rth/tasked/internal/tui"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage local profiles (one per Google account)",
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show all profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if len(cfg.Profiles) == 0 {
			printf("%s\n", tui.Dim.Render("no profiles yet — run `tasked login`"))
			return nil
		}
		for name, p := range cfg.Profiles {
			marker := "  "
			if name == cfg.ActiveProfile {
				marker = tui.Selected.Render("* ")
			}
			list := p.ActiveListTitle
			if list == "" {
				list = tui.Dim.Render("(no list set)")
			}
			printf("%s%s  %s  %s\n", marker, name, tui.Dim.Render(p.Email), list)
		}
		return nil
	},
}

var profileUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set the active profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if _, ok := cfg.Profiles[args[0]]; !ok {
			return fmt.Errorf("profile %q not found", args[0])
		}
		cfg.ActiveProfile = args[0]
		if err := cfg.Save(); err != nil {
			return err
		}
		printf("%s Active profile: %s\n", tui.Ok.Render("✓"), args[0])
		return nil
	},
}

var profileAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new profile by signing in to a Google account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		name := args[0]
		email, err := auth.Login(cmd.Context(), name)
		if err != nil {
			return err
		}
		prof := &config.Profile{Name: name, Email: email}
		cfg.UpsertProfile(prof)
		if err := cfg.Save(); err != nil {
			return err
		}
		printf("%s Added profile %s (%s)\n", tui.Ok.Render("✓"), name, email)
		return nil
	},
}

var profileRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a profile and delete its token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if err := cfg.RemoveProfile(args[0]); err != nil {
			return err
		}
		if err := cfg.Save(); err != nil {
			return err
		}
		printf("%s Removed profile %s\n", tui.Ok.Render("✓"), args[0])
		return nil
	},
}

func init() {
	profileCmd.AddCommand(profileListCmd, profileUseCmd, profileAddCmd, profileRemoveCmd)
	rootCmd.AddCommand(profileCmd)
}
