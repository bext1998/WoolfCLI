package orchestrator

import (
	"strings"
	"testing"

	"woolf/internal/agents"
	"woolf/internal/session"
)

func TestContextBuilderIncludesDraftInterventionsAndPriorResponses(t *testing.T) {
	stance := "agree"
	messages := (ContextBuilder{}).Build(agents.Role{
		Name:             "strict-editor",
		DisplayName:      "Strict Editor",
		Model:            "model",
		FocusAreas:       []string{"clarity", "structure"},
		SystemPrompt:     "You edit strictly.",
		ResponseTemplate: "Stance: <tag>",
	}, session.Session{
		Source: &session.Source{Type: "md", Content: "draft body"},
		Summaries: map[string]string{
			"round-1": "earlier summary",
		},
		Interventions: []session.Intervention{
			{
				Type:       "focus",
				Content:    "tighten the opening",
				FocusRange: &session.FocusRange{StartLine: 2, EndLine: 5},
			},
		},
		Rounds: []session.Round{
			{
				RoundIndex: 1,
				Responses: []session.Response{
					{AgentName: "casual-reader", StanceTag: &stance, Content: "the hook works"},
				},
			},
		},
	}, 1)

	if len(messages) != 2 {
		t.Fatalf("messages = %d, want 2", len(messages))
	}
	system := messages[0].Content
	user := messages[1].Content
	assertContains(t, system, "You edit strictly.")
	assertContains(t, system, "Focus areas: clarity, structure.")
	assertContains(t, system, "Stance: <tag>")
	assertContains(t, user, "## Draft")
	assertContains(t, user, "draft body")
	assertContains(t, user, "## Session Summaries")
	assertContains(t, user, "round-1: earlier summary")
	assertContains(t, user, "## User Interventions")
	assertContains(t, user, "tighten the opening (focus lines 2-5)")
	assertContains(t, user, "## Previous Discussion")
	assertContains(t, user, "casual-reader [agree]: the hook works")
	assertContains(t, user, "clearly use one stance")
}

func TestContextBuilderUsesSourcePreviewWhenContentIsEmpty(t *testing.T) {
	messages := (ContextBuilder{}).Build(agents.Role{
		Name:         "reader",
		DisplayName:  "Reader",
		Model:        "model",
		SystemPrompt: "Read.",
	}, session.Session{
		Source: &session.Source{Type: "txt", ContentPreview: "preview only"},
	}, 1)

	assertContains(t, messages[1].Content, "preview only")
}

func assertContains(t *testing.T, value, want string) {
	t.Helper()
	if !strings.Contains(value, want) {
		t.Fatalf("%q does not contain %q", value, want)
	}
}
