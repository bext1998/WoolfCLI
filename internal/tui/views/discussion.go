package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"woolf/internal/tui/components"
)

type EntryType int

const (
	EntrySystem EntryType = iota
	EntryAgent
)

type DiscussionEntry struct {
	Type     EntryType
	Content  string
	Agent    string
	Model    string
	Color    string
	Stance   string
	Round    int
	System   bool
}

type DiscussionLog struct {
	Entries      []DiscussionEntry
	ScrollOffset int
}

func (d *DiscussionLog) AddSystem(text string) {
	d.Entries = append(d.Entries, DiscussionEntry{
		Type:    EntrySystem,
		Content: text,
		System:  true,
	})
	d.ScrollOffset = 0
}

func (d *DiscussionLog) AddAgent(agent, model, color string, round int) int {
	idx := len(d.Entries)
	d.Entries = append(d.Entries, DiscussionEntry{
		Type:  EntryAgent,
		Agent: agent,
		Model: model,
		Color: color,
		Round: round,
	})
	d.ScrollOffset = 0
	return idx
}

func (d *DiscussionLog) AppendContent(delta string) {
	if len(d.Entries) == 0 {
		return
	}
	d.Entries[len(d.Entries)-1].Content += delta
	d.ScrollOffset = 0
}

func (d *DiscussionLog) SetStance(idx int, stance string) {
	if idx >= 0 && idx < len(d.Entries) {
		d.Entries[idx].Stance = stance
	}
}

func (d *DiscussionLog) ScrollUp(lines int) {
	maxScroll := len(d.Entries) - 1
	if maxScroll < 0 {
		maxScroll = 0
	}
	d.ScrollOffset += lines
	if d.ScrollOffset > maxScroll {
		d.ScrollOffset = maxScroll
	}
}

func (d *DiscussionLog) ScrollDown(lines int) {
	d.ScrollOffset -= lines
	if d.ScrollOffset < 0 {
		d.ScrollOffset = 0
	}
}

func (d *DiscussionLog) ScrollTop() {
	if len(d.Entries) == 0 {
		d.ScrollOffset = 0
		return
	}
	d.ScrollOffset = len(d.Entries) - 1
}

func (d *DiscussionLog) ScrollBottom() {
	d.ScrollOffset = 0
}

func (d *DiscussionLog) Render(width, height int, borderStyle lipgloss.Style) string {
	if height <= 0 {
		return ""
	}
	availableHeight := height - 2
	if availableHeight <= 0 {
		return ""
	}
	visible := d.visibleEntries(availableHeight)
	var sb strings.Builder
	for _, entry := range visible {
		sb.WriteString(d.renderEntry(entry, width-4))
		sb.WriteByte('\n')
	}
	content := strings.TrimRight(sb.String(), "\n")
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		MaxHeight(height).
		Render(borderStyle.Render(content))
}

func (d *DiscussionLog) visibleEntries(max int) []DiscussionEntry {
	if len(d.Entries) == 0 {
		return nil
	}
	start := len(d.Entries) - 1 - d.ScrollOffset
	if start < 0 {
		start = 0
	}
	end := start - max + 1
	if end < 0 {
		end = 0
	}
	if end > start {
		end = start
	}
	result := make([]DiscussionEntry, 0, start-end+1)
	for i := start; i >= end; i-- {
		result = append(result, d.Entries[i])
	}
	return result
}

func (d *DiscussionLog) renderEntry(entry DiscussionEntry, maxWidth int) string {
	switch entry.Type {
	case EntrySystem:
		return renderSystemEntry(entry, maxWidth)
	case EntryAgent:
		return renderAgentEntry(entry, maxWidth)
	default:
		return ""
	}
}

func renderSystemEntry(entry DiscussionEntry, maxWidth int) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#585B70"))
	return style.Render(entry.Content)
}

func renderAgentEntry(entry DiscussionEntry, maxWidth int) string {
	var sb strings.Builder
	badge := components.AgentBadge(entry.Agent, entry.Color)
	modelInfo := lipgloss.NewStyle().Foreground(lipgloss.Color("#585B70")).Render(" (" + entry.Model + ")")
	roundInfo := lipgloss.NewStyle().Foreground(lipgloss.Color("#585B70")).Render(" Round " + itoa(entry.Round))
	header := badge + modelInfo + roundInfo
	if entry.Stance != "" {
		header += " " + components.StanceTag(entry.Stance)
	}
	sb.WriteString(header)
	sb.WriteByte('\n')
	contentStyle := lipgloss.NewStyle().Width(maxWidth)
	sb.WriteString(contentStyle.Render(entry.Content))
	return sb.String()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
