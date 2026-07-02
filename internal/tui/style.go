package tui

import "github.com/charmbracelet/lipgloss"

var (
	Title    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	Dim      = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	Ok       = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	Warn     = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	Danger   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	Selected = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	Hint     = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)

	prioStyles = map[int]lipgloss.Style{
		1: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("196")).Padding(0, 1),
		2: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("208")).Padding(0, 1),
		3: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("220")).Padding(0, 1),
		4: lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("39")).Padding(0, 1),
		5: lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("240")).Padding(0, 1),
	}
)

func PriorityChip(p int) string {
	if p < 1 {
		p = 1
	}
	if p > 5 {
		p = 5
	}
	s, ok := prioStyles[p]
	if !ok {
		s = prioStyles[5]
	}
	return s.Render("p" + string(rune('0'+p)))
}
