package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"

	"github.com/n3tw0rth/tasked/internal/config"
	"github.com/n3tw0rth/tasked/internal/tasks"
)

// version is set at build time with:
//
//	-ldflags "-X github.com/n3tw0rth/tasked/cmd.version=..."
var version = ""

var (
	profileFlag string
	listFlag    string
)

var rootCmd = &cobra.Command{
	Use:           "tasked",
	Short:         "Manage Google Tasks with priorities from your terminal",
	Version:       buildVersion(),
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&profileFlag, "profile", "", "profile to use (default: active profile)")
}

// buildVersion prefers the ldflags version, then the module version recorded
// by `go install module@version`.
func buildVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
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

func newTasksClient(ctx context.Context, prof *config.Profile) (*tasks.Client, error) {
	return tasks.New(ctx, prof.Name)
}

// openList sets up a task command: the resolved profile's client and the
// task list to work on (see resolveList).
func openList(cmd *cobra.Command) (context.Context, *tasks.Client, tasks.TaskList, error) {
	ctx, _, prof, err := loadCtx(cmd)
	if err != nil {
		return nil, nil, tasks.TaskList{}, err
	}
	c, err := newTasksClient(ctx, prof)
	if err != nil {
		return nil, nil, tasks.TaskList{}, err
	}
	tl, err := resolveList(ctx, c, prof)
	if err != nil {
		return nil, nil, tasks.TaskList{}, err
	}
	return ctx, c, tl, nil
}

// addListFlag registers --list on task commands.
func addListFlag(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().StringVarP(&listFlag, "list", "l", "", "task list name or ID (default: active list)")
	}
}

// resolveList picks the task list a command works on: --list (name or ID),
// else the profile's active list, else the account's default list, so a
// fresh login works without running `lists switch` first.
func resolveList(ctx context.Context, c *tasks.Client, prof *config.Profile) (tasks.TaskList, error) {
	if listFlag != "" {
		lists, err := c.Lists(ctx)
		if err != nil {
			return tasks.TaskList{}, err
		}
		if tl, ok := findList(lists, listFlag); ok {
			return tl, nil
		}
		return tasks.TaskList{}, fmt.Errorf("no task list matches %q — see `tasked lists ls`", listFlag)
	}
	if prof.ActiveListID != "" {
		return tasks.TaskList{ID: prof.ActiveListID, Title: prof.ActiveListTitle}, nil
	}
	return c.DefaultList(ctx)
}

// findList matches a task list by exact ID, then by title (case-insensitive).
func findList(lists []tasks.TaskList, q string) (tasks.TaskList, bool) {
	q = strings.TrimSpace(q)
	for _, tl := range lists {
		if tl.ID == q {
			return tl, true
		}
	}
	for _, tl := range lists {
		if strings.EqualFold(tl.Title, q) {
			return tl, true
		}
	}
	return tasks.TaskList{}, false
}

// requireTTY fails fast when stdin isn't a terminal, so interactive pickers
// and prompts error out in scripts instead of misbehaving.
func requireTTY() error {
	fi, err := os.Stdin.Stat()
	if err != nil || fi.Mode()&os.ModeCharDevice == 0 {
		return errors.New("stdin is not a terminal — pass tasks as arguments instead (see --help)")
	}
	return nil
}

func printf(format string, a ...any) { fmt.Printf(format, a...) }
