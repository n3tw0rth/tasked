package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func press(m *formModel, keys ...tea.KeyType) {
	for _, k := range keys {
		m.Update(tea.KeyMsg{Type: k})
	}
}

func TestFormBlocksInvalidDue(t *testing.T) {
	m := newForm("New task", FormResult{Title: "pay rent", Due: "2026-07-15 14:30"})
	press(m, tea.KeyEnter, tea.KeyEnter) // title -> notes -> due
	if m.focus != fieldDue {
		t.Fatalf("focus = %d, want due field", m.focus)
	}
	press(m, tea.KeyEnter)
	if m.focus != fieldDue || m.done {
		t.Errorf("enter on an invalid due date moved on (focus %d, done %v)", m.focus, m.done)
	}
	if m.dueHint() == "" {
		t.Error("no hint shown for the invalid due date")
	}
}

func TestFormSubmitAndNotesEdited(t *testing.T) {
	m := newForm("Edit task", FormResult{Title: "t", Notes: "line one\nline two", Due: "tmr", Priority: 2})
	press(m, tea.KeyEnter, tea.KeyEnter, tea.KeyEnter, tea.KeyEnter)
	if !m.done || m.cancel {
		t.Fatalf("form not submitted (done %v, cancel %v)", m.done, m.cancel)
	}
	// The single-line input flattens the newline; that alone must not count
	// as an edit, or saving would rewrite multi-line notes.
	if m.notes.Value() != m.notesInit {
		t.Error("untouched notes reported as edited")
	}

	m = newForm("Edit task", FormResult{Title: "t", Notes: "old"})
	press(m, tea.KeyTab)
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("!")})
	if m.notes.Value() == m.notesInit {
		t.Error("typed notes not reported as edited")
	}
}
