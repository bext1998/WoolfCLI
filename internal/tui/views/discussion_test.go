package views

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestDiscussionLogAddSystem(t *testing.T) {
	log := DiscussionLog{}
	log.AddSystem("hello")
	if len(log.Entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(log.Entries))
	}
	if log.Entries[0].Content != "hello" {
		t.Errorf("content = %q, want %q", log.Entries[0].Content, "hello")
	}
	if log.Entries[0].Type != EntrySystem {
		t.Errorf("type = %v, want %v", log.Entries[0].Type, EntrySystem)
	}
}

func TestDiscussionLogAddAgent(t *testing.T) {
	log := DiscussionLog{}
	idx := log.AddAgent("Strict Editor", "gpt-4o", "#d14d41", 1)
	if idx != 0 {
		t.Errorf("idx = %d, want 0", idx)
	}
	if len(log.Entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(log.Entries))
	}
	e := log.Entries[0]
	if e.Agent != "Strict Editor" {
		t.Errorf("agent = %q", e.Agent)
	}
	if e.Round != 1 {
		t.Errorf("round = %d", e.Round)
	}
}

func TestDiscussionLogAppendContent(t *testing.T) {
	log := DiscussionLog{}
	log.AddAgent("A", "m", "#fff", 1)
	log.AppendContent("hello ")
	log.AppendContent("world")
	if log.Entries[0].Content != "hello world" {
		t.Errorf("content = %q", log.Entries[0].Content)
	}
}

func TestDiscussionLogSetStance(t *testing.T) {
	log := DiscussionLog{}
	log.AddAgent("A", "m", "#fff", 1)
	log.SetStance(0, "agree")
	if log.Entries[0].Stance != "agree" {
		t.Errorf("stance = %q", log.Entries[0].Stance)
	}
}

func TestDiscussionLogScroll(t *testing.T) {
	log := DiscussionLog{}
	for i := 0; i < 20; i++ {
		log.AddSystem("entry")
	}
	if log.ScrollOffset != 0 {
		t.Errorf("initial offset = %d, want 0", log.ScrollOffset)
	}
	log.ScrollUp(5)
	if log.ScrollOffset != 5 {
		t.Errorf("offset after up 5 = %d, want 5", log.ScrollOffset)
	}
	log.ScrollDown(2)
	if log.ScrollOffset != 3 {
		t.Errorf("offset after down 2 = %d, want 3", log.ScrollOffset)
	}
	log.ScrollUp(100)
	if log.ScrollOffset > 19 {
		t.Errorf("offset after large up = %d, want <= 19", log.ScrollOffset)
	}
	log.ScrollTop()
	if log.ScrollOffset != 19 {
		t.Errorf("offset after top = %d, want 19", log.ScrollOffset)
	}
	log.ScrollBottom()
	if log.ScrollOffset != 0 {
		t.Errorf("offset after bottom = %d, want 0", log.ScrollOffset)
	}
}

func TestDiscussionLogRender(t *testing.T) {
	log := DiscussionLog{}
	log.AddSystem("System message")
	log.AddAgent("Editor", "gpt-4o", "#d14d41", 1)
	log.AppendContent("This is agent content.")

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)

	rendered := log.Render(80, 20, borderStyle)
	if rendered == "" {
		t.Error("render returned empty string")
	}
	if !strings.Contains(rendered, "Editor") {
		t.Error("render missing agent name")
	}
	if !strings.Contains(rendered, "agent content") {
		t.Error("render missing agent content")
	}
	if !strings.Contains(rendered, "System message") {
		t.Error("render missing system message")
	}
}

func TestDiscussionLogEmptyRender(t *testing.T) {
	log := DiscussionLog{}
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)
	rendered := log.Render(80, 20, borderStyle)
	if rendered == "" {
		t.Error("empty render should not return empty string")
	}
}
