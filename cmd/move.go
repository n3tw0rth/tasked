package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/n3tw0rth/tasked/internal/tasks"
	"github.com/n3tw0rth/tasked/internal/tui"
)

// moveTarget is either a priority (1..5) or a destination task list.
type moveTarget struct {
	priority  int
	listID    string
	listTitle string
}

var moveCmd = &cobra.Command{
	Use:     "move [p1-p5 | list] [task...]",
	Aliases: []string{"prio", "priority"},
	Short:   "Change task priorities or move tasks to another list",
	Long: `Pick tasks, then a destination: a priority (p1-p5) or another task list.
The destination can be given as the first argument to skip the second
picker, followed by tasks to skip the first.

` + taskArgsHelp,
	Example: `  tasked move
  tasked move p1
  tasked move Work "Pay rent" milk`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, c, tl, err := openList(cmd)
		if err != nil {
			return err
		}

		var target moveTarget
		if len(args) > 0 {
			target, err = resolveTargetArg(ctx, c, tl.ID, args[0])
			if err != nil {
				return err
			}
			args = args[1:]
		}

		ts, err := c.Tasks(ctx, tl.ID, false)
		if err != nil {
			return err
		}
		sel, err := selectTasks(ts, args, "Move — space to toggle", true)
		if err != nil || len(sel) == 0 {
			return err
		}

		if target == (moveTarget{}) {
			target, err = pickTarget(ctx, c, tl.ID)
			if err != nil {
				return err
			}
			if target == (moveTarget{}) {
				return nil
			}
		}

		for _, t := range sel {
			if target.priority != 0 {
				_, err = c.SetPriority(ctx, tl.ID, t.ID, target.priority)
			} else {
				_, err = c.MoveToList(ctx, tl.ID, t.ID, target.listID)
			}
			if err != nil {
				return err
			}
		}
		if target.priority != 0 {
			printf("%s Set %s on %d task(s)\n", tui.Ok.Render("✓"), tui.PriorityChip(target.priority), len(sel))
		} else {
			printf("%s Moved %d task(s) to %s\n", tui.Ok.Render("✓"), len(sel), target.listTitle)
		}
		return nil
	},
}

// resolveTargetArg interprets the CLI argument as a priority (1-5, p1-p5) or,
// failing that, as a task list title or ID.
func resolveTargetArg(ctx context.Context, c *tasks.Client, currentListID, arg string) (moveTarget, error) {
	if p, err := parsePriorityArg(arg); err == nil {
		return moveTarget{priority: p}, nil
	}
	lists, err := c.Lists(ctx)
	if err != nil {
		return moveTarget{}, err
	}
	tl, ok := findList(lists, arg)
	if !ok {
		return moveTarget{}, fmt.Errorf("no priority or task list matches %q (use p1-p5 or a list name)", arg)
	}
	if tl.ID == currentListID {
		return moveTarget{}, fmt.Errorf("the tasks are already in %q", tl.Title)
	}
	return moveTarget{listID: tl.ID, listTitle: tl.Title}, nil
}

// pickTarget shows a single-select picker over priorities and other task
// lists. It returns the zero moveTarget if the user cancelled.
func pickTarget(ctx context.Context, c *tasks.Client, currentListID string) (moveTarget, error) {
	labels := map[int]string{1: "Highest", 2: "High", 3: "Medium", 4: "Low", 5: "Lowest"}
	var items []tui.PickerItem
	for p := tasks.Highest; p <= tasks.Lowest; p++ {
		items = append(items, tui.PickerItem{
			ID:    "p" + strconv.Itoa(p),
			Label: labels[p],
			Chip:  tui.PriorityChip(p),
		})
	}
	lists, err := c.Lists(ctx)
	if err != nil {
		return moveTarget{}, err
	}
	byID := map[string]tasks.TaskList{}
	for _, tl := range lists {
		if tl.ID == currentListID {
			continue
		}
		byID[tl.ID] = tl
		items = append(items, tui.PickerItem{
			ID:    tl.ID,
			Label: tl.Title,
			Extra: "list",
		})
	}
	res, err := tui.RunPicker("Move to — priority or list", items, false)
	if err != nil {
		return moveTarget{}, err
	}
	if res.Cancel || len(res.Selected) == 0 {
		return moveTarget{}, nil
	}
	id := res.Selected[0].ID
	if strings.HasPrefix(id, "p") && len(id) == 2 {
		p, err := strconv.Atoi(id[1:])
		if err == nil {
			return moveTarget{priority: p}, nil
		}
	}
	tl := byID[id]
	return moveTarget{listID: tl.ID, listTitle: tl.Title}, nil
}

func parsePriorityArg(s string) (int, error) {
	v := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(s)), "p")
	p, err := strconv.Atoi(v)
	if err != nil || p < tasks.Highest || p > tasks.Lowest {
		return 0, fmt.Errorf("invalid priority %q (use 1-5 or p1-p5)", s)
	}
	return p, nil
}

func init() {
	addListFlag(moveCmd)
	rootCmd.AddCommand(moveCmd)
}
