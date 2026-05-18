package tui

import "woolf/internal/orchestrator"

type pipelineEventMsg struct {
	Event orchestrator.Event
}

type pipelineDoneMsg struct {
}

type windowSizeMsg struct {
	Width  int
	Height int
}

type startPipelineMsg struct{}
