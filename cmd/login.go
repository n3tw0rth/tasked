package cmd

import (
	"context"

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
		prof, err := signIn(cmd.Context(), cfg, name)
		if err != nil {
			return err
		}
		printf("%s Signed in as %s (profile: %s)\n", tui.Ok.Render("✓"), prof.Email, name)
		return nil
	},
}

// signIn runs the browser sign-in for the named profile and saves it,
// keeping the existing profile's settings (such as its active list).
func signIn(ctx context.Context, cfg *config.Config, name string) (*config.Profile, error) {
	email, err := auth.Login(ctx, name)
	if err != nil {
		return nil, err
	}
	prof := cfg.Profiles[name]
	if prof == nil {
		prof = &config.Profile{Name: name}
	}
	prof.Email = email
	cfg.UpsertProfile(prof)
	return prof, cfg.Save()
}

func init() { rootCmd.AddCommand(loginCmd) }
