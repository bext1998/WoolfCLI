package cli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"woolf/internal/agents"
	"woolf/internal/ingestion"
	"woolf/internal/orchestrator"
	"woolf/internal/session"
	"woolf/internal/tui"
)

func newStartCommand(app *App) *cobra.Command {
	var draft string
	var preset string
	var agentNames string
	var rounds int
	var useTUI bool
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Create a new session",
		RunE: func(cmd *cobra.Command, args []string) error {
			if preset == "" {
				preset = app.loaded.Config.Defaults.DefaultPreset
			}
			if rounds == 0 {
				rounds = app.loaded.Config.Defaults.MaxRounds
			}
			title := "untitled"
			if draft != "" {
				title = strings.TrimSuffix(filepath.Base(draft), filepath.Ext(draft))
			}
			if app.client == nil && strings.TrimSpace(app.loaded.Config.API.OpenRouterKey) == "" {
				return fmt.Errorf("CFG-001: OpenRouter API key is required")
			}
			var source *session.Source
			if draft != "" {
				src, err := sourceFromDraft(draft)
				if err != nil {
					return err
				}
				source = &src
			}
			registry, err := agents.NewRegistry(app.loaded.Paths.AgentsDir)
			if err != nil {
				return err
			}
			roles, err := rolesForStart(registry, preset, agentNames)
			if err != nil {
				return err
			}
			sess, path, err := app.store.Create(title, "")
			if err != nil {
				return err
			}
			if source != nil {
				sess.Source = source
				if path, err = app.store.Save(sess); err != nil {
					return err
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created session: %s\n", sess.SessionID)
			fmt.Fprintf(cmd.OutOrStdout(), "path: %s\n", path)
			fmt.Fprintf(cmd.OutOrStdout(), "preset: %s\n", preset)
			fmt.Fprintf(cmd.OutOrStdout(), "rounds: %d\n", rounds)
			if agentNames != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "agents: %s\n", agentNames)
			}
			pipeline := orchestrator.Pipeline{Client: app.chatClient(), Store: app.store}
			if useTUI {
				sessCopy := sess
				return tui.Run(pipeline, app.store, &sessCopy, orchestrator.Options{Rounds: rounds, Roles: roles})
			}
			events, err := pipeline.Run(context.Background(), &sess, orchestrator.Options{Rounds: rounds, Roles: roles})
			if err != nil {
				return err
			}
			var runErr error
			for event := range events {
				switch event.Type {
				case orchestrator.EventAgentStarted:
					fmt.Fprintf(cmd.OutOrStdout(), "round %d agent %s started\n", event.RoundIndex, event.AgentName)
				case orchestrator.EventAgentDelta:
					if app.verbose {
						fmt.Fprint(cmd.OutOrStdout(), event.Content)
					}
				case orchestrator.EventAgentFinished:
					fmt.Fprintf(cmd.OutOrStdout(), "round %d agent %s finished\n", event.RoundIndex, event.AgentName)
				case orchestrator.EventError:
					fmt.Fprintf(cmd.ErrOrStderr(), "%s\n", event.Error)
					if runErr == nil {
						runErr = event.Error
					}
				case orchestrator.EventDone:
					fmt.Fprintf(cmd.OutOrStdout(), "finished session: %s status=%s (%s rounds)\n", event.Session.SessionID, event.Session.Status, strconv.Itoa(event.Session.Totals.RoundsCompleted))
				}
			}
			return runErr
		},
	}
	cmd.Flags().StringVar(&draft, "draft", "", "draft file")
	cmd.Flags().StringVar(&preset, "preset", "", "agent preset")
	cmd.Flags().StringVar(&agentNames, "agents", "", "comma-separated agent names")
	cmd.Flags().IntVar(&rounds, "rounds", 0, "discussion rounds")
	cmd.Flags().BoolVar(&useTUI, "tui", false, "launch interactive TUI")
	return cmd
}

func rolesForStart(registry *agents.Registry, preset string, names string) ([]agents.Role, error) {
	if names != "" {
		return registry.ResolveRoles(strings.Split(names, ","))
	}
	return registry.ResolvePreset(preset)
}

func sourceFromDraft(path string) (session.Source, error) {
	doc, err := ingestion.Ingest(path)
	if err != nil {
		return session.Source{}, err
	}
	sum := sha256.Sum256([]byte(doc.Content))
	return session.Source{
		Type:           doc.Format,
		Path:           doc.Path,
		Content:        doc.Content,
		ContentHash:    hex.EncodeToString(sum[:]),
		ContentPreview: preview(doc.Content, 200),
	}, nil
}

func preview(value string, maxRunes int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= maxRunes {
		return string(runes)
	}
	return string(runes[:maxRunes])
}
