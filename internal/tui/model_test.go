package tui

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"woolf/internal/agents"
	"woolf/internal/openrouter"
	"woolf/internal/orchestrator"
	"woolf/internal/session"
	"woolf/internal/tui/views"
)

type fakeTUIClient struct{}

func (fakeTUIClient) StreamChat(ctx context.Context, req openrouter.ChatRequest) (<-chan openrouter.StreamEvent, error) {
	ch := make(chan openrouter.StreamEvent, 1)
	ch <- openrouter.StreamEvent{
		Content: "delta",
		Usage: &openrouter.Usage{
			PromptTokens:     2,
			CompletionTokens: 3,
			TotalTokens:      5,
		},
	}
	close(ch)
	return ch, nil
}

func TestStartPipelineDoesNotCancelImmediately(t *testing.T) {
	store := session.NewStore(t.TempDir())
	sess := testSession()
	m := model{
		pipeline: orchestrator.Pipeline{Client: fakeTUIClient{}, Store: store},
		store:    store,
		session:  &sess,
		opts: orchestrator.Options{
			Rounds: 1,
			Roles: []agents.Role{{
				Name:         "strict-editor",
				DisplayName:  "Strict Editor",
				Model:        "openai/gpt-4o-mini",
				Stance:       "critique",
				SystemPrompt: "Review the draft.",
			}},
		},
		discussion: views.DiscussionLog{},
		input:      views.NewInputArea(),
		state:      stateIdle,
	}

	nextModel, cmd := m.startPipeline()
	if cmd == nil {
		t.Fatal("startPipeline returned nil command")
	}
	m = nextModel.(model)
	if m.state != stateRunning {
		t.Fatalf("state = %v, want running", m.state)
	}

	sawRoundStart := false
	for step := 0; step < 10 && cmd != nil; step++ {
		msg := cmd()
		switch msg := msg.(type) {
		case pipelineEventMsg:
			if msg.Event.Type == orchestrator.EventRoundStarted {
				sawRoundStart = true
			}
			nextModel, nextCmd := m.handlePipelineEvent(msg)
			m = nextModel.(model)
			cmd = nextCmd
		case pipelineDoneMsg:
			nextModel, nextCmd := m.handlePipelineDone(msg)
			m = nextModel.(model)
			cmd = nextCmd
		default:
			t.Fatalf("command returned %T, want pipeline event or done", msg)
		}
		if m.state == stateDone {
			break
		}
	}
	if !sawRoundStart {
		t.Fatal("pipeline did not emit a round_started event")
	}
	if m.session == nil || m.session.Totals.TotalTokens != 5 {
		t.Fatalf("session totals were not preserved: %#v", m.session)
	}
}

func TestHandlePipelineDonePreservesCurrentSession(t *testing.T) {
	sess := testSession()
	sess.Status = session.StatusCompleted
	sess.Totals.RoundsCompleted = 1
	sess.Totals.TotalTokens = 5

	m := model{
		session:    &sess,
		discussion: views.DiscussionLog{},
		state:      stateRunning,
	}

	nextModel, _ := m.handlePipelineDone(pipelineDoneMsg{})
	m = nextModel.(model)
	if m.session == nil {
		t.Fatal("session was cleared")
	}
	if m.session.SessionID != sess.SessionID {
		t.Fatalf("session id = %q, want %q", m.session.SessionID, sess.SessionID)
	}
	if m.state != stateDone {
		t.Fatalf("state = %v, want done", m.state)
	}
}

func testSession() session.Session {
	now := time.Now().UTC()
	return session.Session{
		SessionID:     "20260519-120000-test",
		Version:       session.Version,
		Title:         "test",
		Status:        session.StatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
		AgentsConfig:  []session.AgentConfig{},
		Rounds:        []session.Round{},
		Interventions: []session.Intervention{},
		Summaries:     map[string]string{},
	}
}

var _ tea.Model = model{}
