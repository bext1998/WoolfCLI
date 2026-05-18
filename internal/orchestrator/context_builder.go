package orchestrator

import (
	"fmt"
	"sort"
	"strings"

	"woolf/internal/agents"
	"woolf/internal/openrouter"
	"woolf/internal/session"
)

type ContextBuilder struct{}

func (ContextBuilder) Build(role agents.Role, sess session.Session, roundIndex int) []openrouter.ChatMessage {
	system := strings.TrimSpace(role.SystemPrompt)
	if len(role.FocusAreas) > 0 {
		system += "\n\nFocus areas: " + strings.Join(role.FocusAreas, ", ") + "."
	}
	if strings.TrimSpace(role.ResponseTemplate) != "" {
		system += "\n\nUse this response template:\n" + strings.TrimSpace(role.ResponseTemplate)
	}

	var user strings.Builder
	user.WriteString("Review the following draft in your assigned role. Provide concrete, actionable feedback.\n")
	if hasPriorResponses(sess, roundIndex) {
		user.WriteString("When responding to prior agents, clearly use one stance: agree, disagree, extend, or neutral.\n")
	}
	user.WriteByte('\n')

	writeSource(&user, sess.Source)
	writeSummaries(&user, sess.Summaries)
	writeInterventions(&user, sess.Interventions)
	writePreviousDiscussion(&user, sess.Rounds, roundIndex)

	return []openrouter.ChatMessage{
		{Role: "system", Content: system},
		{Role: "user", Content: strings.TrimSpace(user.String())},
	}
}

func writeSource(user *strings.Builder, source *session.Source) {
	if source == nil {
		return
	}
	content := source.Content
	if content == "" {
		content = source.ContentPreview
	}
	if strings.TrimSpace(content) == "" {
		return
	}
	user.WriteString("## Draft\n")
	user.WriteString(content)
	user.WriteString("\n\n")
}

func writeSummaries(user *strings.Builder, summaries map[string]string) {
	if len(summaries) == 0 {
		return
	}
	keys := make([]string, 0, len(summaries))
	for key := range summaries {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	user.WriteString("## Session Summaries\n")
	for _, key := range keys {
		value := strings.TrimSpace(summaries[key])
		if value == "" {
			continue
		}
		user.WriteString("- ")
		user.WriteString(key)
		user.WriteString(": ")
		user.WriteString(value)
		user.WriteByte('\n')
	}
	user.WriteByte('\n')
}

func writeInterventions(user *strings.Builder, interventions []session.Intervention) {
	if len(interventions) == 0 {
		return
	}
	user.WriteString("## User Interventions\n")
	for _, intervention := range interventions {
		user.WriteString("- ")
		user.WriteString(intervention.Content)
		if intervention.FocusRange != nil {
			user.WriteString(fmt.Sprintf(" (focus lines %d-%d)", intervention.FocusRange.StartLine, intervention.FocusRange.EndLine))
		}
		user.WriteByte('\n')
	}
	user.WriteByte('\n')
}

func writePreviousDiscussion(user *strings.Builder, rounds []session.Round, roundIndex int) {
	user.WriteString("## Previous Discussion\n")
	for _, round := range rounds {
		if round.RoundIndex > roundIndex {
			continue
		}
		for _, response := range round.Responses {
			if strings.TrimSpace(response.Content) == "" {
				continue
			}
			user.WriteString(response.AgentName)
			if response.StanceTag != nil && *response.StanceTag != "" {
				user.WriteString(" [")
				user.WriteString(*response.StanceTag)
				user.WriteString("]")
			}
			user.WriteString(": ")
			user.WriteString(response.Content)
			user.WriteString("\n\n")
		}
	}
}

func hasPriorResponses(sess session.Session, roundIndex int) bool {
	for _, round := range sess.Rounds {
		if round.RoundIndex > roundIndex {
			continue
		}
		if len(round.Responses) > 0 {
			return true
		}
	}
	return false
}
