package cmd

import (
	"github.com/spf13/cobra"

	"github.com/n3tw0rth/tasked/internal/auth"
	"github.com/n3tw0rth/tasked/internal/config"
	"github.com/n3tw0rth/tasked/internal/tui"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Sign in to Google (creates or refreshes the current profile)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		name := profileFlag
		if name == "" {
			name = cfg.ActiveProfile
		}
		if name == "" {
			name = "default"
		}
		email, err := auth.Login(cmd.Context(), name)
		if err != nil {
			return err
		}
		prof := cfg.Profiles[name]
		if prof == nil {
			prof = &config.Profile{Name: name}
		}
		prof.Email = email
		cfg.UpsertProfile(prof)
		if err := cfg.Save(); err != nil {
			return err
		}
		printf("%s Signed in as %s (profile: %s)\n", tui.Ok.Render("✓"), email, name)
		return nil
	},
}

func init() { rootCmd.AddCommand(loginCmd) }
