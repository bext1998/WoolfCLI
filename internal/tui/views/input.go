package views

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type InputArea struct {
	textarea textarea.Model
	focused  bool
	width    int
	height   int
}

func NewInputArea() InputArea {
	ta := textarea.New()
	ta.Placeholder = "Type your message or /command..."
	ta.CharLimit = 10000
	ta.ShowLineNumbers = false
	ta.SetHeight(3)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.BlurredStyle.CursorLine = lipgloss.NewStyle()
	ta.KeyMap.InsertNewline.SetEnabled(false)
	return InputArea{
		textarea: ta,
		height:   3,
	}
}

func (i *InputArea) Focus() {
	i.focused = true
	i.textarea.Focus()
}

func (i *InputArea) Blur() {
	i.focused = false
	i.textarea.Blur()
}

func (i *InputArea) Focused() bool {
	return i.focused
}

func (i *InputArea) SetWidth(w int) {
	i.width = w
	i.textarea.SetWidth(w - 4)
}

func (i *InputArea) SetHeight(h int) {
	i.height = h
	i.textarea.SetHeight(h)
}

func (i *InputArea) Value() string {
	return i.textarea.Value()
}

func (i *InputArea) Reset() {
	i.textarea.Reset()
}

func (i *InputArea) Update(msg tea.Msg) (InputArea, tea.Cmd) {
	var cmd tea.Cmd
	i.textarea, cmd = i.textarea.Update(msg)
	return *i, cmd
}

func (i *InputArea) View(borderStyle lipgloss.Style) string {
	content := i.textarea.View()
	return borderStyle.Render(content)
}

type SlashCommand struct {
	Command string
	Args    string
}

var slashRegex = regexp.MustCompile(`^/(\w+)(?:\s+(.*))?$`)

func ParseSlashCommand(text string) (SlashCommand, bool) {
	trimmed := strings.TrimSpace(text)
	m := slashRegex.FindStringSubmatch(trimmed)
	if m == nil {
		return SlashCommand{}, false
	}
	return SlashCommand{
		Command: m[1],
		Args:    strings.TrimSpace(m[2]),
	}, true
}

type DiscussionViewport struct {
	viewport viewport.Model
	content  string
}

func NewDiscussionViewport(width, height int) DiscussionViewport {
	vp := viewport.New(width, height)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6C7086"))
	return DiscussionViewport{viewport: vp}
}

func (d *DiscussionViewport) SetContent(content string) {
	d.viewport.SetContent(content)
}

func (d *DiscussionViewport) Update(msg tea.Msg) (DiscussionViewport, tea.Cmd) {
	var cmd tea.Cmd
	d.viewport, cmd = d.viewport.Update(msg)
	return *d, cmd
}

func (d *DiscussionViewport) View() string {
	return d.viewport.View()
}

func (d *DiscussionViewport) GotoBottom() {
	d.viewport.GotoBottom()
}

func (d *DiscussionViewport) SetSize(w, h int) {
	d.viewport.Width = w
	d.viewport.Height = h
}
