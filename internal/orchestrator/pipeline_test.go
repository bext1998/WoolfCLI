package orchestrator

import (
	"context"
	"strings"
	"testing"
	"time"

	"woolf/internal/agents"
	"woolf/internal/openrouter"
	"woolf/internal/session"
)

type fakeClient struct {
	requests []openrouter.ChatRequest
}

func (f *fakeClient) StreamChat(ctx context.Context, req openrouter.ChatRequest) (<-chan openrouter.StreamEvent, error) {
	f.requests = append(f.requests, req)
	ch := make(chan openrouter.StreamEvent, 2)
	ch <- openrouter.StreamEvent{Content: req.Model + " response", Usage: &openrouter.Usage{PromptTokens: 1, CompletionTokens: 2, TotalTokens: 3}}
	close(ch)
	return ch, nil
}

func TestPipelineRunsAgentsAndPersistsSession(t *testing.T) {
	dir := t.TempDir()
	store := session.NewStore(dir)
	sess, _, err := store.Create("draft", "")
	if err != nil {
		t.Fatal(err)
	}
	sess.Source = &session.Source{Type: "md", Content: "hello"}
	client := &fakeClient{}
	roles := []agents.Role{
		{Name: "a", DisplayName: "A", Model: "model-a", SystemPrompt: "prompt"},
		{Name: "b", DisplayName: "B", Model: "model-b", SystemPrompt: "prompt"},
		{Name: "c", DisplayName: "C", Model: "model-c", SystemPrompt: "prompt"},
	}
	events, err := (Pipeline{Client: client, Store: store}).Run(context.Background(), &sess, Options{Rounds: 2, Roles: roles})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	for range events {
	}
	loaded, _, err := store.Load(sess.SessionID)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Status != session.StatusCompleted {
		t.Fatalf("status = %s", loaded.Status)
	}
	if len(loaded.Rounds) != 2 {
		t.Fatalf("rounds = %d", len(loaded.Rounds))
	}
	if len(loaded.Rounds[0].Responses) != 3 {
		t.Fatalf("responses = %d", len(loaded.Rounds[0].Responses))
	}
	if len(client.requests) != 6 {
		t.Fatalf("requests = %d", len(client.requests))
	}
	if loaded.Totals.TotalTokens != 18 {
		t.Fatalf("total tokens = %d", loaded.Totals.TotalTokens)
	}
	if !strings.Contains(client.requests[1].Messages[1].Content, "a: model-a response") {
		t.Fatalf("second agent context does not include first agent response: %q", client.requests[1].Messages[1].Content)
	}
	if !strings.Contains(client.requests[2].Messages[1].Content, "model-b response") {
		t.Fatalf("third agent context does not include second agent response: %q", client.requests[2].Messages[1].Content)
	}
}

func TestPipelineContextCancelPausesSession(t *testing.T) {
	store := session.NewStore(t.TempDir())
	sess := session.Session{
		SessionID:     "20260508-091500-cancel",
		Version:       session.Version,
		Status:        session.StatusActive,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
		AgentsConfig:  []session.AgentConfig{},
		Rounds:        []session.Round{},
		Interventions: []session.Intervention{},
		Summaries:     map[string]string{},
	}
	if _, err := store.Save(sess); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	events, err := (Pipeline{Client: &fakeClient{}, Store: store}).Run(ctx, &sess, Options{
		Rounds: 1,
		Roles:  []agents.Role{{Name: "a", DisplayName: "A", Model: "m", SystemPrompt: "p"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for range events {
	}
	loaded, _, err := store.Load(sess.SessionID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Status != session.StatusPaused {
		t.Fatalf("status = %s, want paused", loaded.Status)
	}
}
