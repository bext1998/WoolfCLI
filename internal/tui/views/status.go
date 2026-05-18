package views

import (
	"fmt"

	"woolf/internal/tui/components"
)

type StatusInfo struct {
	SessionTitle string
	RoundCurrent int
	RoundMax     int
	Tokens       float64
	Cost         float64
	State        string
	Message      string
}

func StatusView(info StatusInfo) string {
	progress := components.Progress(info.State)
	cost := components.CostMeter(info.Tokens, info.Cost)
	rounds := ""
	if info.RoundMax > 0 {
		rounds = fmt.Sprintf("R: %d/%d", info.RoundCurrent, info.RoundMax)
	}
	line := fmt.Sprintf("%s  %s  %s", progress, cost, rounds)
	if info.SessionTitle != "" {
		line = fmt.Sprintf("Session: %s  %s", info.SessionTitle, line)
	}
	if info.Message != "" {
		line += "  " + info.Message
	}
	return line
}
