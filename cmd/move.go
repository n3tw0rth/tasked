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
	Use:     "move [p1-p5 | list name]",
	Aliases: []string{"prio", "priority"},
	Short:   "Change task priorities or move tasks to another list",
	Long: `Pick tasks, then a destination: a priority (p1-p5) or another task list.
The destination can also be given as the argument to skip the second picker.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, _, prof, err := loadCtx(cmd)
		if err != nil {
			return err
		}
		listID, _, err := requireActiveList(prof)
		if err != nil {
			return err
		}
		c, err := newTasksClient(ctx, prof)
		if err != nil {
			return err
		}

		var target moveTarget
		if len(args) == 1 {
			target, err = resolveTargetArg(ctx, c, listID, args[0])
			if err != nil {
				return err
			}
		}

		ts, err := c.Tasks(ctx, listID, false)
		if err != nil {
			return err
		}
		items := make([]tui.PickerItem, 0, len(ts))
		for _, t := range ts {
			extra := ""
			if t.HasDue {
				extra = t.Due.Format("Jan 2")
			}
			items = append(items, tui.PickerItem{
				ID:    t.ID,
				Label: t.Title,
				Chip:  tui.PriorityChip(t.Priority),
				Extra: extra,
			})
		}
		res, err := tui.RunPicker("Move — space to toggle", items, true)
		if err != nil {
			return err
		}
		if res.Cancel || len(res.Selected) == 0 {
			return nil
		}

		if target == (moveTarget{}) {
			target, err = pickTarget(ctx, c, listID)
			if err != nil {
				return err
			}
			if target == (moveTarget{}) {
				return nil
			}
		}

		for _, it := range res.Selected {
			if target.priority != 0 {
				_, err = c.SetPriority(ctx, listID, it.ID, target.priority)
			} else {
				_, err = c.MoveToList(ctx, listID, it.ID, target.listID)
			}
			if err != nil {
				return err
			}
		}
		if target.priority != 0 {
			printf("%s Set %s on %d task(s)\n", tui.Ok.Render("✓"), tui.PriorityChip(target.priority), len(res.Selected))
		} else {
			printf("%s Moved %d task(s) to %s\n", tui.Ok.Render("✓"), len(res.Selected), target.listTitle)
		}
		return nil
	},
}

// resolveTargetArg interprets the CLI argument as a priority (1-5, p1-p5) or,
// failing that, as a task list title (case-insensitive).
func resolveTargetArg(ctx context.Context, c *tasks.Client, currentListID, arg string) (moveTarget, error) {
	if p, err := parsePriorityArg(arg); err == nil {
		return moveTarget{priority: p}, nil
	}
	lists, err := c.Lists(ctx)
	if err != nil {
		return moveTarget{}, err
	}
	for _, tl := range lists {
		if strings.EqualFold(tl.Title, strings.TrimSpace(arg)) {
			if tl.ID == currentListID {
				return moveTarget{}, fmt.Errorf("%q is already the active list", tl.Title)
			}
			return moveTarget{listID: tl.ID, listTitle: tl.Title}, nil
		}
	}
	return moveTarget{}, fmt.Errorf("no priority or task list matches %q (use p1-p5 or a list name)", arg)
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

func init() { rootCmd.AddCommand(moveCmd) }
