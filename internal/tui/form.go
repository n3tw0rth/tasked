package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type FormResult struct {
	Title    string
	Due      string // raw user text; caller parses via tasks.ParseDueInput
	Priority int
	Cancel   bool
}

type formModel struct {
	title    textinput.Model
	due      textinput.Model
	priority int
	focus    int // 0=title, 1=due, 2=priority
	done     bool
	cancel   bool
}

func newForm() *formModel {
	t := textinput.New()
	t.Placeholder = "task title"
	t.CharLimit = 200
	t.Prompt = ""
	t.Focus()

	d := textinput.New()
	d.Placeholder = "today | tomorrow | 2026-07-15 | 2026-07-15 14:30 | (empty)"
	d.CharLimit = 40
	d.Prompt = ""

	return &formModel{title: t, due: d, priority: 3}
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
			if m.focus == 2 {
				if m.priority > 1 {
					m.priority--
				}
				return m, nil
			}
		case "right":
			if m.focus == 2 {
				if m.priority < 5 {
					m.priority++
				}
				return m, nil
			}
		case "enter":
			if m.focus < 2 {
				m.setFocus(m.focus + 1)
				return m, nil
			}
			if strings.TrimSpace(m.title.Value()) == "" {
				m.setFocus(0)
				return m, nil
			}
			m.done = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	switch m.focus {
	case 0:
		m.title, cmd = m.title.Update(msg)
	case 1:
		m.due, cmd = m.due.Update(msg)
	}
	return m, cmd
}

func (m *formModel) setFocus(i int) {
	if i < 0 {
		i = 2
	}
	if i > 2 {
		i = 0
	}
	m.focus = i
	m.title.Blur()
	m.due.Blur()
	switch i {
	case 0:
		m.title.Focus()
	case 1:
		m.due.Focus()
	}
}

func (m *formModel) View() string {
	if m.done {
		return ""
	}
	var b strings.Builder
	b.WriteString(Title.Render("New task"))
	b.WriteString("\n")

	b.WriteString(label("Title    ", m.focus == 0))
	b.WriteString(m.title.View())
	b.WriteString("\n")

	b.WriteString(label("Due      ", m.focus == 1))
	b.WriteString(m.due.View())
	b.WriteString("\n")

	b.WriteString(label("Priority ", m.focus == 2))
	b.WriteString(renderPrioRow(m.priority, m.focus == 2))
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
	for i := 1; i <= 5; i++ {
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

func RunForm() (FormResult, error) {
	m := newForm()
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return FormResult{}, err
	}
	if m.cancel {
		return FormResult{Cancel: true}, nil
	}
	return FormResult{
		Title:    strings.TrimSpace(m.title.Value()),
		Due:      strings.TrimSpace(m.due.Value()),
		Priority: m.priority,
	}, nil
}
