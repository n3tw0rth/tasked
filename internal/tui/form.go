package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/n3tw0rth/tasked/internal/tasks"
)

// FormResult holds the form's initial values and what the user submitted.
type FormResult struct {
	Title    string
	Notes    string
	Due      string // raw user text; caller parses via tasks.ParseDueInput
	Priority int
	// NotesEdited reports whether the notes field was changed. The field is a
	// single line, so multi-line notes are shown flattened; callers should
	// only save Notes when this is true.
	NotesEdited bool
	Cancel      bool
}

const (
	fieldTitle = iota
	fieldNotes
	fieldDue
	fieldPriority
	numFields
)

type formModel struct {
	header    string
	title     textinput.Model
	notes     textinput.Model
	notesInit string
	due       textinput.Model
	priority  int
	focus     int
	done      bool
	cancel    bool
}

func newForm(header string, init FormResult) *formModel {
	t := textinput.New()
	t.Placeholder = "task title"
	t.CharLimit = 1024
	t.Width = 60
	t.Prompt = ""
	t.SetValue(init.Title)
	t.Focus()

	n := textinput.New()
	n.Placeholder = "(optional)"
	n.CharLimit = 8192
	n.Width = 60
	n.Prompt = ""
	n.SetValue(init.Notes)

	d := textinput.New()
	d.Placeholder = "tmr | fri | 3d | 2w | eow | 2026-07-15 | (empty)"
	d.CharLimit = 40
	d.Prompt = ""
	d.SetValue(init.Due)

	prio := init.Priority
	if prio < tasks.Highest || prio > tasks.Lowest {
		prio = tasks.Default
	}
	return &formModel{header: header, title: t, notes: n, notesInit: n.Value(), due: d, priority: prio}
}

func (m *formModel) Init() tea.Cmd { return textinput.Blink }

func (m *formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancel = true
			m.done = true
			return m, tea.Quit
		case "tab", "shift+tab", "up", "down":
			delta := 1
			if msg.String() == "shift+tab" || msg.String() == "up" {
				delta = -1
			}
			m.setFocus(m.focus + delta)
			return m, nil
		case "left":
			if m.focus == fieldPriority {
				if m.priority > tasks.Highest {
					m.priority--
				}
				return m, nil
			}
		case "right":
			if m.focus == fieldPriority {
				if m.priority < tasks.Lowest {
					m.priority++
				}
				return m, nil
			}
		case "enter":
			if m.focus == fieldDue && m.dueErr() != nil {
				return m, nil
			}
			if m.focus < fieldPriority {
				m.setFocus(m.focus + 1)
				return m, nil
			}
			if strings.TrimSpace(m.title.Value()) == "" {
				m.setFocus(fieldTitle)
				return m, nil
			}
			if m.dueErr() != nil {
				m.setFocus(fieldDue)
				return m, nil
			}
			m.done = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	switch m.focus {
	case fieldTitle:
		m.title, cmd = m.title.Update(msg)
	case fieldNotes:
		m.notes, cmd = m.notes.Update(msg)
	case fieldDue:
		m.due, cmd = m.due.Update(msg)
	}
	return m, cmd
}

func (m *formModel) setFocus(i int) {
	m.focus = (i + numFields) % numFields
	m.title.Blur()
	m.notes.Blur()
	m.due.Blur()
	switch m.focus {
	case fieldTitle:
		m.title.Focus()
	case fieldNotes:
		m.notes.Focus()
	case fieldDue:
		m.due.Focus()
	}
}

// dueErr validates the due field; empty means no due date.
func (m *formModel) dueErr() error {
	v := strings.TrimSpace(m.due.Value())
	if v == "" {
		return nil
	}
	_, err := tasks.ParseDueInput(v)
	return err
}

// dueHint previews the date the due field resolves to, or why it doesn't.
func (m *formModel) dueHint() string {
	v := strings.TrimSpace(m.due.Value())
	if v == "" {
		return ""
	}
	d, err := tasks.ParseDueInput(v)
	if err != nil {
		return "  " + Danger.Render("✗ "+err.Error())
	}
	return "  " + Dim.Render("→ "+d.Format("Mon Jan 2"))
}

func (m *formModel) View() string {
	if m.done {
		return ""
	}
	var b strings.Builder
	b.WriteString(Title.Render(m.header))
	b.WriteString("\n")

	b.WriteString(label("Title    ", m.focus == fieldTitle))
	b.WriteString(m.title.View())
	b.WriteString("\n")

	b.WriteString(label("Notes    ", m.focus == fieldNotes))
	b.WriteString(m.notes.View())
	b.WriteString("\n")

	b.WriteString(label("Due      ", m.focus == fieldDue))
	b.WriteString(m.due.View())
	b.WriteString(m.dueHint())
	b.WriteString("\n")

	b.WriteString(label("Priority ", m.focus == fieldPriority))
	b.WriteString(renderPrioRow(m.priority, m.focus == fieldPriority))
	b.WriteString("\n")

	b.WriteString(Hint.Render("tab/↑↓ move  ←/→ priority  enter next/submit  esc cancel"))
	return b.String()
}

func label(name string, active bool) string {
	if active {
		return Selected.Render(name)
	}
	return Dim.Render(name)
}

func renderPrioRow(current int, active bool) string {
	var parts []string
	for i := tasks.Highest; i <= tasks.Lowest; i++ {
		s := fmt.Sprintf("p%d", i)
		if i == current {
			if active {
				s = Selected.Render("[" + s + "]")
			} else {
				s = "[" + s + "]"
			}
		} else {
			s = Dim.Render(" " + s + " ")
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, " ")
}

// RunForm shows an inline task form under the given header, prefilled from
// init.
func RunForm(header string, init FormResult) (FormResult, error) {
	m := newForm(header, init)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return FormResult{}, err
	}
	if m.cancel {
		return FormResult{Cancel: true}, nil
	}
	return FormResult{
		Title:       strings.TrimSpace(m.title.Value()),
		Notes:       strings.TrimSpace(m.notes.Value()),
		Due:         strings.TrimSpace(m.due.Value()),
		Priority:    m.priority,
		NotesEdited: m.notes.Value() != m.notesInit,
	}, nil
}
