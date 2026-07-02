package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/n3tw0rth/tasked/internal/config"
	"github.com/n3tw0rth/tasked/internal/tasks"
)

var profileFlag string

var rootCmd = &cobra.Command{
	Use:           "tasked",
	Short:         "Manage Google Tasks with priorities from your terminal",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&profileFlag, "profile", "", "profile to use (default: active profile)")
}

// loadCtx returns config + resolved profile + a context for use in the command.
func loadCtx(cmd *cobra.Command) (context.Context, *config.Config, *config.Profile, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, nil, err
	}
	prof, err := cfg.ResolveProfile(profileFlag)
	if err != nil {
		return nil, nil, nil, err
	}
	return cmd.Context(), cfg, prof, nil
}

// requireActiveList returns the list ID and title for the resolved profile.
// If none is set, it errors with a helpful message.
func requireActiveList(prof *config.Profile) (string, string, error) {
	if prof.ActiveListID == "" {
		return "", "", errors.New("no active task list — run `tasked lists switch` to pick one")
	}
	return prof.ActiveListID, prof.ActiveListTitle, nil
}

func newTasksClient(ctx context.Context, prof *config.Profile) (*tasks.Client, error) {
	return tasks.New(ctx, prof.Name)
}

func printf(format string, a ...any) { fmt.Printf(format, a...) }
