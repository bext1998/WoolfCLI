package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"woolf/internal/orchestrator"
	"woolf/internal/session"
	"woolf/internal/tui/views"
)

func Run(pipeline orchestrator.Pipeline, store session.Store, sess *session.Session, opts orchestrator.Options) error {
	m := model{
		pipeline: pipeline,
		store:    store,
		session:  sess,
		opts:     opts,
		discussion: views.DiscussionLog{
			Entries:      make([]views.DiscussionEntry, 0),
			ScrollOffset: 0,
		},
		input: views.NewInputArea(),
		focus: focusInput,
		state: stateIdle,
		width: 80,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
