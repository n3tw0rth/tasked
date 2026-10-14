package cmd

import (
	"github.com/spf13/cobra"

	"github.com/n3tw0rth/tasked/internal/auth"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Inspect the OAuth client used to sign in",
	Long: `tasked signs in with its built-in OAuth client. To use your own Desktop app
client instead, set both TASKED_CLIENT_ID and TASKED_CLIENT_SECRET.`,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show which OAuth client is in use",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := auth.LoadClient()
		if err != nil {
			return err
		}
		printf("client  %s\nsource  %s\n", c.ID, c.Source)
		return nil
	},
}

func init() {
	authCmd.AddCommand(authStatusCmd)
	rootCmd.AddCommand(authCmd)
}
