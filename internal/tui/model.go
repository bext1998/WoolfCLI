package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"woolf/internal/orchestrator"
	"woolf/internal/session"
	"woolf/internal/tui/components"
	"woolf/internal/tui/views"
)

type focusArea int

const (
	focusDiscussion focusArea = iota
	focusInput
)

type appState int

const (
	stateIdle appState = iota
	stateRunning
	stateDone
)

type model struct {
	pipeline orchestrator.Pipeline
	store    session.Store
	session  *session.Session
	opts     orchestrator.Options

	discussion views.DiscussionLog
	input      views.InputArea

	focus focusArea
	state appState
	err   error
	quit  bool

	events <-chan orchestrator.Event
	cancel context.CancelFunc

	width  int
	height int
	ready  bool
}

func (m model) Init() tea.Cmd {
	_, cmd := m.input.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	return tea.Batch(
		tea.SetWindowTitle("Woolf - AI Writing Salon"),
		cmd,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg)

	case tea.KeyMsg:
		if m.focus == focusInput {
			return m.handleInputKey(msg)
		}
		return m.handleDiscussionKey(msg)

	case pipelineEventMsg:
		return m.handlePipelineEvent(msg)

	case pipelineDoneMsg:
		return m.handlePipelineDone(msg)

	case startPipelineMsg:
		return m.startPipeline()
	}

	return m, nil
}

func (m model) View() string {
	if !m.ready {
		return "Woolf TUI initializing...\n"
	}
	if m.quit {
		return ""
	}

	minW, minH := 80, 24
	if m.width < minW || m.height < minH {
		return m.renderTooSmall(minW, minH)
	}

	headerHeight := 0
	if m.session != nil {
		headerHeight = 1
	}
	inputHeight := 5
	borderOverhead := 4
	availableHeight := m.height - headerHeight - 2 - borderOverhead
	discHeight := availableHeight - inputHeight
	if discHeight < 3 {
		discHeight = 3
	}
	statusHeight := 1

	var sb strings.Builder
	if m.session != nil {
		sb.WriteString(m.renderHeader())
		sb.WriteByte('\n')
	}

	sb.WriteString(m.renderDiscussion(discHeight))
	sb.WriteByte('\n')
	sb.WriteString(m.renderInput(inputHeight))
	sb.WriteByte('\n')
	sb.WriteString(m.renderStatus(statusHeight))
	return sb.String()
}

func (m model) renderHeader() string {
	title := m.session.Title
	if title == "" {
		title = m.session.SessionID
	}
	status := string(m.session.Status)
	roundInfo := fmt.Sprintf("Rounds: %d", len(m.session.Rounds))
	line := fmt.Sprintf("Woolf | %s | %s | %s", title, status, roundInfo)
	return HeaderBox().Width(m.width).MaxWidth(m.width).Render(line)
}

func (m model) renderDiscussion(height int) string {
	content := m.discussion.Render(m.width, height, borderStyle)
	if content == "" {
		placeholder := dimStyle.Render("Discussion area - type /help for commands")
		return borderStyle.Width(m.width).Height(height).MaxHeight(height).Render(placeholder)
	}
	return content
}

func (m model) renderInput(height int) string {
	h := height - 2
	if h < 1 {
		h = 1
	}
	m.input.SetWidth(m.width - 2)
	m.input.SetHeight(h)
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(0, 1)
	return box.Width(m.width).Height(height).MaxHeight(height).Render(m.input.View(box))
}

func (m model) renderStatus(height int) string {
	info := views.StatusInfo{
		SessionTitle: "",
		Tokens:       float64(m.session.Totals.TotalTokens),
		Cost:         m.session.Totals.TotalCostUSD,
		RoundCurrent: m.session.Totals.RoundsCompleted,
		Message:      "",
	}
	if m.session != nil {
		info.RoundCurrent = m.session.Totals.RoundsCompleted
		if m.opts.Rounds > 0 {
			info.RoundMax = m.opts.Rounds
		}
	}
	switch m.state {
	case stateRunning:
		info.State = "running"
	case stateDone:
		info.State = "done"
	default:
		info.State = "idle"
	}
	if m.err != nil {
		info.Message = errorStyle.Render(m.err.Error())
	}
	return statusBarStyle.Width(m.width).MaxWidth(m.width).Render(views.StatusView(info))
}

func (m model) renderTooSmall(minW, minH int) string {
	msg := fmt.Sprintf(
		"Terminal too small - need at least %dx%d (current: %dx%d)",
		minW, minH, m.width, m.height,
	)
	msg += "\n\nPlease resize your terminal window."
	return lipgloss.NewStyle().
		Foreground(ColorError).
		Bold(true).
		Padding(1).
		Render(msg)
}

func (m model) handleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	if !m.ready {
		m.ready = true
	}
	m.input.SetWidth(m.width - 4)
	return m, nil
}

func (m model) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.submitInput()

	case "esc":
		return m.cancelOrQuit()

	case "ctrl+c":
		return m.cancelOrQuit()

	case "tab":
		m.focus = focusDiscussion
		m.input.Blur()
		return m, nil

	case "?":
		m.showHelp()
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) handleDiscussionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m.discussion.ScrollDown(3)

	case "k", "up":
		m.discussion.ScrollUp(3)

	case "g":
		m.discussion.ScrollTop()

	case "G":
		m.discussion.ScrollBottom()

	case "tab":
		m.focus = focusInput
		m.input.Focus()
		return m, nil

	case "esc", "ctrl+c":
		return m.cancelOrQuit()

	case "?":
		m.showHelp()
		return m, nil
	}

	return m, nil
}

func (m model) submitInput() (tea.Model, tea.Cmd) {
	text := strings.TrimSpace(m.input.Value())
	if text == "" {
		m.input.Reset()
		return m, nil
	}
	m.input.Reset()

	cmd, isCmd := views.ParseSlashCommand(text)
	if isCmd {
		return m.executeCommand(cmd, text)
	}

	return m.saveInterventionAndStart(text)
}

func (m model) saveInterventionAndStart(text string) (tea.Model, tea.Cmd) {
	intervention := session.Intervention{
		AfterRound: len(m.session.Rounds),
		Type:       "chat",
		Content:    text,
		Timestamp:  time.Now().UTC(),
	}
	m.session.Interventions = append(m.session.Interventions, intervention)
	m.session.UpdatedAt = time.Now().UTC()
	if _, err := m.store.Save(*m.session); err != nil {
		m.discussion.AddSystem(fmt.Sprintf("Error saving intervention: %v", err))
		return m, nil
	}
	m.discussion.AddSystem(">" + text)
	return m.startPipeline()
}

func (m model) executeCommand(cmd views.SlashCommand, raw string) (tea.Model, tea.Cmd) {
	switch cmd.Command {
	case "start", "next":
		if cmd.Command == "next" {
			m.opts.Rounds = 1
		}
		if cmd.Args != "" {
			intervention := session.Intervention{
				AfterRound: len(m.session.Rounds),
				Type:       "chat",
				Content:    cmd.Args,
				Timestamp:  time.Now().UTC(),
			}
			m.session.Interventions = append(m.session.Interventions, intervention)
			if _, err := m.store.Save(*m.session); err != nil {
				m.discussion.AddSystem("Error saving: " + err.Error())
				return m, nil
			}
			m.discussion.AddSystem("> " + cmd.Args)
		}
		return m.startPipeline()

	case "end":
		m.session.Status = session.StatusCompleted
		m.session.UpdatedAt = time.Now().UTC()
		if _, err := m.store.Save(*m.session); err != nil {
			m.discussion.AddSystem("Error saving: " + err.Error())
			return m, nil
		}
		m.state = stateDone
		m.quit = true
		m.discussion.AddSystem("Session completed.")
		return m, tea.Quit

	case "pause":
		if m.cancel != nil {
			m.cancel()
			m.cancel = nil
		}
		m.state = stateIdle
		m.discussion.AddSystem("Pipeline paused.")
		return m, nil

	case "quit":
		if m.session.Status == session.StatusActive {
			m.session.Status = session.StatusPaused
		}
		m.session.UpdatedAt = time.Now().UTC()
		if _, err := m.store.Save(*m.session); err != nil {
			m.discussion.AddSystem("Error saving: " + err.Error())
			return m, nil
		}
		m.quit = true
		m.discussion.AddSystem("Goodbye.")
		return m, tea.Quit

	case "help":
		m.showHelp()
		return m, nil

	case "status":
		m.showStatus()
		return m, nil

	case "agents":
		m.showAgents()
		return m, nil

	case "cost":
		m.showCost()
		return m, nil

	default:
		m.discussion.AddSystem("Unknown command: /" + cmd.Command + " - type /help for available commands")
		return m, nil
	}
}

func (m model) startPipeline() (tea.Model, tea.Cmd) {
	if m.state == stateRunning {
		m.discussion.AddSystem("Pipeline already running.")
		return m, nil
	}
	m.state = stateRunning
	m.err = nil

	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel

	sess := *m.session
	opts := m.opts
	if opts.Rounds <= 0 {
		opts.Rounds = 1
	}

	events, err := m.pipeline.Run(ctx, &sess, opts)
	if err != nil {
		cancel()
		m.state = stateIdle
		m.discussion.AddSystem(fmt.Sprintf("Error starting pipeline: %v", err))
		return m, nil
	}

	m.events = events
	return m, listenEvents(events)
}

func (m model) handlePipelineEvent(msg pipelineEventMsg) (tea.Model, tea.Cmd) {
	event := msg.Event
	switch event.Type {
	case orchestrator.EventRoundStarted:
		m.discussion.AddSystem(fmt.Sprintf("Round %d started", event.RoundIndex))

	case orchestrator.EventAgentStarted:
		idx := m.discussion.AddAgent(event.AgentName, "", "", event.RoundIndex)
		_ = idx

	case orchestrator.EventAgentDelta:
		m.discussion.AppendContent(event.Content)

	case orchestrator.EventAgentFinished:
		m.session = &event.Session

	case orchestrator.EventError:
		m.session = &event.Session
		m.discussion.AddSystem(fmt.Sprintf("Error: %v", event.Error))
		if m.err == nil {
			m.err = event.Error
		}

	case orchestrator.EventDone:
		m.session = &event.Session
	}

	if m.events != nil {
		return m, listenEvents(m.events)
	}
	return m, nil
}

func (m model) handlePipelineDone(msg pipelineDoneMsg) (tea.Model, tea.Cmd) {
	m.state = stateIdle
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	if m.session == nil {
		m.discussion.AddSystem("Pipeline complete.")
		return m, nil
	}
	if m.err == nil && m.session.Status == session.StatusCompleted {
		m.state = stateDone
	}
	m.discussion.AddSystem(fmt.Sprintf(
		"Pipeline complete - %d rounds, %d tokens, $%.4f",
		m.session.Totals.RoundsCompleted,
		m.session.Totals.TotalTokens,
		m.session.Totals.TotalCostUSD,
	))
	return m, nil
}

func (m model) cancelOrQuit() (tea.Model, tea.Cmd) {
	if m.state == stateRunning {
		if m.cancel != nil {
			m.cancel()
			m.cancel = nil
		}
		m.state = stateIdle
		m.discussion.AddSystem("Cancelled.")
		return m, nil
	}
	if m.session.Status == session.StatusActive {
		m.session.Status = session.StatusPaused
		m.session.UpdatedAt = time.Now().UTC()
		m.store.Save(*m.session)
	}
	m.quit = true
	return m, tea.Quit
}

func (m *model) showHelp() {
	help := `Commands:
  /start [text]  Start the discussion pipeline
  /next [text]   Run one more round
  /end           Complete the session
  /pause         Pause running pipeline
  /quit          Save and quit
  /status        Show session details
  /agents        Show agent list
  /cost          Show cost breakdown
  /help          Show this help

Keys:
  j/k or arrows Scroll discussion
  g/G           Jump to top/bottom
  Tab           Switch focus (discussion/input)
  Enter         Submit input
  Esc           Cancel pipeline or quit
  Ctrl+C        Cancel pipeline or quit
  ?             Show help`
	m.discussion.AddSystem(help)
}

func (m *model) showStatus() {
	sess := m.session
	info := fmt.Sprintf(
		"Session: %s\nTitle: %s\nStatus: %s\nRounds: %d\n",
		sess.SessionID, sess.Title, sess.Status, len(sess.Rounds),
	)
	info += fmt.Sprintf(
		"Tokens: %d (prompt: %d, completion: %d)\nCost: $%.4f\n",
		sess.Totals.TotalTokens,
		sess.Totals.TotalPromptTokens,
		sess.Totals.TotalCompletionTokens,
		sess.Totals.TotalCostUSD,
	)
	if sess.Source != nil {
		info += fmt.Sprintf("Source: %s (%s)\n", sess.Source.Path, sess.Source.Type)
	}
	m.discussion.AddSystem(strings.TrimRight(info, "\n"))
}

func (m *model) showAgents() {
	if len(m.session.AgentsConfig) == 0 {
		m.discussion.AddSystem("No agents configured.")
		return
	}
	var lines []string
	for _, a := range m.session.AgentsConfig {
		badge := components.AgentBadge(a.DisplayName, "")
		lines = append(lines, fmt.Sprintf("- %s (%s) stance=%s", badge, a.Model, a.Stance))
	}
	m.discussion.AddSystem(strings.Join(lines, "\n"))
}

func (m *model) showCost() {
	sess := m.session
	info := fmt.Sprintf(
		"Prompt tokens: %d\nCompletion tokens: %d\nTotal tokens: %d\nTotal cost: $%.4f",
		sess.Totals.TotalPromptTokens,
		sess.Totals.TotalCompletionTokens,
		sess.Totals.TotalTokens,
		sess.Totals.TotalCostUSD,
	)
	m.discussion.AddSystem(info)
}

func listenEvents(ch <-chan orchestrator.Event) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-ch
		if !ok {
			return pipelineDoneMsg{}
		}
		return pipelineEventMsg{Event: event}
	}
}
