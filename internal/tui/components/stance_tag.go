package components

import "github.com/charmbracelet/lipgloss"

var stanceColors = map[string]lipgloss.Color{
	"agree":    lipgloss.Color("#A6E3A1"),
	"disagree": lipgloss.Color("#F38BA8"),
	"extend":   lipgloss.Color("#89DCEB"),
	"neutral":  lipgloss.Color("#9399B2"),
}

func StanceTag(stance string) string {
	if stance == "" {
		return ""
	}
	c, ok := stanceColors[stance]
	if !ok {
		c = lipgloss.Color("#9399B2")
	}
	return lipgloss.NewStyle().
		Foreground(c).
		Render("[" + stance + "]")
}
