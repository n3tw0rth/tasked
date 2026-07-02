package cmd

import (
	"github.com/spf13/cobra"

	"github.com/n3tw0rth/tasked/internal/auth"
	"github.com/n3tw0rth/tasked/internal/tui"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove the saved token for a profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, _, prof, err := loadCtx(cmd)
		if err != nil {
			return err
		}
		if err := auth.Logout(prof.Name); err != nil {
			return err
		}
		printf("%s Logged out profile %s\n", tui.Ok.Render("✓"), prof.Name)
		return nil
	},
}

func init() { rootCmd.AddCommand(logoutCmd) }
