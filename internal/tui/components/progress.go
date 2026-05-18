package components

import "github.com/charmbracelet/lipgloss"

var (
	stateIdle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#9399B2")).Render("○ Idle")
	stateRunning = lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")).Render("● Live")
	statePaused  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAB387")).Render("⏸ Paused")
	stateDone    = lipgloss.NewStyle().Foreground(lipgloss.Color("#89DCEB")).Render("✓ Done")
	stateError   = lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8")).Render("✗ Error")
)

func Progress(state string) string {
	switch state {
	case "running":
		return stateRunning
	case "paused":
		return statePaused
	case "done":
		return stateDone
	case "error":
		return stateError
	default:
		return stateIdle
	}
}
