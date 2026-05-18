package orchestrator

import (
	"context"
	"fmt"
	"time"

	"woolf/internal/agents"
	"woolf/internal/openrouter"
	"woolf/internal/session"
)

type ChatClient interface {
	StreamChat(context.Context, openrouter.ChatRequest) (<-chan openrouter.StreamEvent, error)
}

type Pipeline struct {
	Client         ChatClient
	Store          session.Store
	ContextBuilder ContextBuilder
}

type Options struct {
	Rounds int
	Roles  []agents.Role
}

type EventType string

const (
	EventRoundStarted  EventType = "round_started"
	EventAgentStarted  EventType = "agent_started"
	EventAgentDelta    EventType = "agent_delta"
	EventAgentFinished EventType = "agent_finished"
	EventError         EventType = "error"
	EventDone          EventType = "done"
)

type Event struct {
	Type       EventType
	RoundIndex int
	AgentName  string
	Content    string
	Error      error
	Session    session.Session
}

func (p Pipeline) Run(ctx context.Context, sess *session.Session, opts Options) (<-chan Event, error) {
	if p.Client == nil {
		return nil, fmt.Errorf("orchestrator requires a chat client")
	}
	if p.Store == nil {
		return nil, fmt.Errorf("orchestrator requires a session store")
	}
	if sess == nil {
		return nil, fmt.Errorf("orchestrator requires a session")
	}
	if opts.Rounds <= 0 {
		opts.Rounds = 1
	}
	if len(opts.Roles) == 0 {
		return nil, fmt.Errorf("orchestrator requires at least one agent role")
	}

	events := make(chan Event)
	go func() {
		defer close(events)
		p.run(ctx, sess, opts, events)
	}()
	return events, nil
}

func (p Pipeline) run(ctx context.Context, sess *session.Session, opts Options, events chan<- Event) {
	ensureAgentConfig(sess, opts.Roles)
	hadError := false
	for roundOffset := 0; roundOffset < opts.Rounds; roundOffset++ {
		roundIndex := len(sess.Rounds) + 1
		round := session.Round{
			RoundIndex: roundIndex,
			StartedAt:  time.Now().UTC(),
			Responses:  []session.Response{},
		}
		sess.Rounds = append(sess.Rounds, round)
		events <- Event{Type: EventRoundStarted, RoundIndex: roundIndex, Session: *sess}

		for _, role := range opts.Roles {
			select {
			case <-ctx.Done():
				sess.Status = session.StatusPaused
				if _, err := p.Store.Save(*sess); err != nil {
					events <- Event{Type: EventError, RoundIndex: roundIndex, AgentName: role.Name, Error: err, Session: *sess}
					return
				}
				events <- Event{Type: EventError, RoundIndex: roundIndex, AgentName: role.Name, Error: ctx.Err(), Session: *sess}
				return
			default:
			}

			events <- Event{Type: EventAgentStarted, RoundIndex: roundIndex, AgentName: role.Name, Session: *sess}
			response := session.Response{
				AgentName: role.Name,
				Model:     role.Model,
				StanceTag: stanceFor(role, sess, roundIndex),
				Timestamp: time.Now().UTC(),
				Status:    "completed",
			}
			req := openrouter.ChatRequest{
				Model:       role.Model,
				Messages:    p.ContextBuilder.Build(role, *sess, roundIndex),
				Temperature: role.Temperature,
				MaxTokens:   role.MaxTokens,
			}
			stream, err := p.Client.StreamChat(ctx, req)
			if err != nil {
				hadError = true
				response.Status = "skipped"
				response.Content = err.Error()
				appendResponse(sess, roundIndex, response)
				recalculateTotals(sess)
				sess.Status = session.StatusError
				if _, saveErr := p.Store.Save(*sess); saveErr != nil {
					events <- Event{Type: EventError, RoundIndex: roundIndex, AgentName: role.Name, Error: saveErr, Session: *sess}
					return
				}
				events <- Event{Type: EventError, RoundIndex: roundIndex, AgentName: role.Name, Error: err, Session: *sess}
				continue
			}
			var streamErr error
			for event := range stream {
				if event.Error != nil {
					hadError = true
					response.Status = "skipped"
					response.Content = event.Error.Error()
					streamErr = event.Error
					break
				}
				if event.Content != "" {
					response.Content += event.Content
					events <- Event{Type: EventAgentDelta, RoundIndex: roundIndex, AgentName: role.Name, Content: event.Content, Session: *sess}
				}
				if event.Usage != nil {
					response.Tokens.Prompt = event.Usage.PromptTokens
					response.Tokens.Completion = event.Usage.CompletionTokens
				}
			}
			appendResponse(sess, roundIndex, response)
			recalculateTotals(sess)
			if streamErr != nil {
				sess.Status = session.StatusError
			}
			if _, err := p.Store.Save(*sess); err != nil {
				events <- Event{Type: EventError, RoundIndex: roundIndex, AgentName: role.Name, Error: err, Session: *sess}
				return
			}
			if streamErr != nil {
				events <- Event{Type: EventError, RoundIndex: roundIndex, AgentName: role.Name, Error: streamErr, Session: *sess}
				continue
			}
			events <- Event{Type: EventAgentFinished, RoundIndex: roundIndex, AgentName: role.Name, Session: *sess}
		}
		completeRound(sess, roundIndex)
		recalculateTotals(sess)
		if _, err := p.Store.Save(*sess); err != nil {
			events <- Event{Type: EventError, RoundIndex: roundIndex, Error: err, Session: *sess}
			return
		}
	}
	if hadError {
		sess.Status = session.StatusError
	} else {
		sess.Status = session.StatusCompleted
	}
	recalculateTotals(sess)
	if _, err := p.Store.Save(*sess); err != nil {
		events <- Event{Type: EventError, Error: err, Session: *sess}
		return
	}
	events <- Event{Type: EventDone, Session: *sess}
}

func ensureAgentConfig(sess *session.Session, roles []agents.Role) {
	if len(sess.AgentsConfig) > 0 {
		return
	}
	for i, role := range roles {
		sess.AgentsConfig = append(sess.AgentsConfig, session.AgentConfig{
			Name:        role.Name,
			DisplayName: role.DisplayName,
			Model:       role.Model,
			Stance:      role.Stance,
			Order:       i + 1,
			Color:       role.Color,
		})
	}
}

func stanceFor(role agents.Role, sess *session.Session, roundIndex int) *string {
	hasPriorResponse := false
	for _, round := range sess.Rounds {
		if round.RoundIndex > roundIndex {
			continue
		}
		if len(round.Responses) > 0 {
			hasPriorResponse = true
			break
		}
	}
	if !hasPriorResponse {
		return nil
	}
	stance := "neutral"
	switch role.Stance {
	case "critique":
		stance = "disagree"
	case "support":
		stance = "extend"
	}
	return &stance
}

func appendResponse(sess *session.Session, roundIndex int, response session.Response) {
	for i := range sess.Rounds {
		if sess.Rounds[i].RoundIndex == roundIndex {
			sess.Rounds[i].Responses = append(sess.Rounds[i].Responses, response)
			return
		}
	}
}

func completeRound(sess *session.Session, roundIndex int) {
	for i := range sess.Rounds {
		if sess.Rounds[i].RoundIndex == roundIndex {
			sess.Rounds[i].CompletedAt = time.Now().UTC()
			return
		}
	}
}

func recalculateTotals(sess *session.Session) {
	var totals session.Totals
	for _, round := range sess.Rounds {
		if !round.CompletedAt.IsZero() {
			totals.RoundsCompleted++
		}
		for _, response := range round.Responses {
			totals.TotalPromptTokens += response.Tokens.Prompt
			totals.TotalCompletionTokens += response.Tokens.Completion
			totals.TotalTokens += response.Tokens.Prompt + response.Tokens.Completion
			totals.TotalCostUSD += response.CostUSD
		}
	}
	sess.Totals = totals
}
