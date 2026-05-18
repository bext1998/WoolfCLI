package tui

import "github.com/charmbracelet/lipgloss"

var (
	ColorBackground    = lipgloss.Color("#1E1E2E")
	ColorForeground    = lipgloss.Color("#CDD6F4")
	ColorBorder        = lipgloss.Color("#6C7086")
	ColorHeaderBG      = lipgloss.Color("#313244")
	ColorStatusBG      = lipgloss.Color("#181825")
	ColorHighlight     = lipgloss.Color("#F5C2E7")
	ColorDim           = lipgloss.Color("#585B70")
	ColorError         = lipgloss.Color("#F38BA8")
	ColorWarning       = lipgloss.Color("#FAB387")
	ColorAgentDefault  = lipgloss.Color("#888888")
)

var StanceTagColors = map[string]lipgloss.Color{
	"agree":    lipgloss.Color("#A6E3A1"),
	"disagree": lipgloss.Color("#F38BA8"),
	"extend":   lipgloss.Color("#89DCEB"),
	"neutral":  lipgloss.Color("#9399B2"),
}

type Theme struct {
	Name string
}

var DarkTheme = Theme{Name: "dark"}

var baseStyle = lipgloss.NewStyle().
	Foreground(ColorForeground)

var borderStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(ColorBorder).
	Padding(0, 1)

var headerStyle = lipgloss.NewStyle().
	Background(ColorHeaderBG).
	Foreground(ColorForeground).
	Bold(true).
	Padding(0, 1)

var statusBarStyle = lipgloss.NewStyle().
	Background(ColorStatusBG).
	Foreground(ColorForeground).
	Padding(0, 1)

var errorStyle = lipgloss.NewStyle().
	Foreground(ColorError)

var dimStyle = lipgloss.NewStyle().
	Foreground(ColorDim)

func AppStyle() lipgloss.Style {
	return baseStyle
}

func BorderBox(title string) lipgloss.Style {
	if title != "" {
		return borderStyle.BorderTopForeground(ColorBorder).
			BorderBottomForeground(ColorBorder).
			BorderTop(true).
			BorderBottom(true).
			BorderLeft(true).
			BorderRight(true)
	}
	return borderStyle
}

func HeaderBox() lipgloss.Style {
	return headerStyle
}

func StatusBox() lipgloss.Style {
	return statusBarStyle
}

func ErrorText(s string) string {
	return errorStyle.Render(s)
}

func DimText(s string) string {
	return dimStyle.Render(s)
}
