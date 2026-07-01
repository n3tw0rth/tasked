package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type PickerItem struct {
	ID    string
	Label string
	Extra string // right-aligned dim text (e.g., due date, list size)
	Chip  string // pre-rendered chip prefix (e.g., priority)
}

type PickerResult struct {
	Selected []PickerItem
	Cancel   bool
}

type pickerModel struct {
	title    string
	items    []PickerItem
	cursor   int
	multi    bool
	chosen   map[int]bool
	done     bool
	cancel   bool
	maxShown int
}

func newPicker(title string, items []PickerItem, multi bool) *pickerModel {
	return &pickerModel{
		title:    title,
		items:    items,
		multi:    multi,
		chosen:   map[int]bool{},
		maxShown: 12,
	}
}

func (m *pickerModel) Init() tea.Cmd { return nil }

func (m *pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.cancel = true
			m.done = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case " ":
			if m.multi && len(m.items) > 0 {
				m.chosen[m.cursor] = !m.chosen[m.cursor]
			}
		case "enter":
			if !m.multi && len(m.items) > 0 {
				m.chosen[m.cursor] = true
			}
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *pickerModel) View() string {
	if m.done {
		return ""
	}
	var b strings.Builder
	if m.title != "" {
		b.WriteString(Title.Render(m.title))
		b.WriteString("\n")
	}
	if len(m.items) == 0 {
		b.WriteString(Dim.Render("  (nothing to show)"))
		b.WriteString("\n")
	}

	start := 0
	end := len(m.items)
	if end > m.maxShown {
		half := m.maxShown / 2
		start = m.cursor - half
		if start < 0 {
			start = 0
		}
		end = start + m.maxShown
		if end > len(m.items) {
			end = len(m.items)
			start = end - m.maxShown
			if start < 0 {
				start = 0
			}
		}
	}

	for i := start; i < end; i++ {
		it := m.items[i]
		cursor := "  "
		if i == m.cursor {
			cursor = Selected.Render("> ")
		}
		mark := "  "
		if m.multi {
			if m.chosen[i] {
				mark = Ok.Render("[x] ")
			} else {
				mark = "[ ] "
			}
		}
		chip := ""
		if it.Chip != "" {
			chip = it.Chip + " "
		}
		label := it.Label
		if i == m.cursor {
			label = Selected.Render(label)
		}
		extra := ""
		if it.Extra != "" {
			extra = "  " + Dim.Render(it.Extra)
		}
		b.WriteString(fmt.Sprintf("%s%s%s%s%s\n", cursor, mark, chip, label, extra))
	}
	if len(m.items) > m.maxShown {
		b.WriteString(Dim.Render(fmt.Sprintf("  (%d of %d)\n", m.cursor+1, len(m.items))))
	}
	hint := "↑/↓ move  enter select  esc cancel"
	if m.multi {
		hint = "↑/↓ move  space toggle  enter confirm  esc cancel"
	}
	b.WriteString(Hint.Render(hint))
	return b.String()
}

// RunPicker shows an inline picker. If multi is false, at most one item is
// returned. Result.Cancel is true if the user hit esc/ctrl-c.
func RunPicker(title string, items []PickerItem, multi bool) (PickerResult, error) {
	m := newPicker(title, items, multi)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return PickerResult{}, err
	}
	res := PickerResult{Cancel: m.cancel}
	if !m.cancel {
		for i, ok := range m.chosen {
			if ok && i >= 0 && i < len(items) {
				res.Selected = append(res.Selected, items[i])
			}
		}
	}
	return res, nil
}
