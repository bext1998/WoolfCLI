package components

import "github.com/charmbracelet/lipgloss"

func AgentBadge(name, color string) string {
	c := lipgloss.Color(color)
	if c == "" {
		c = lipgloss.Color("#888888")
	}
	return lipgloss.NewStyle().
		Foreground(c).
		Bold(true).
		Render(name)
}
