package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"woolf/internal/agents"
)

func newAgentsCommand(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Manage agent roles and presets",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List agent roles",
			RunE: func(cmd *cobra.Command, args []string) error {
				registry, err := agents.NewRegistry(app.loaded.Paths.AgentsDir)
				if err != nil {
					return err
				}
				for _, role := range registry.ListRoles() {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\n", role.Name, role.DisplayName, role.Model, role.Stance)
				}
				return nil
			},
		},
		&cobra.Command{
			Use:   "show <name>",
			Short: "Show agent role",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				registry, err := agents.NewRegistry(app.loaded.Paths.AgentsDir)
				if err != nil {
					return err
				}
				role, ok := registry.Role(args[0])
				if !ok {
					return fmt.Errorf("CFG-003: role %s not found", args[0])
				}
				printRole(cmd, role)
				return nil
			},
		},
		&cobra.Command{
			Use:   "add <path>",
			Short: "Add a custom agent role",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				role, err := agents.LoadRole(args[0])
				if err != nil {
					return err
				}
				data, err := os.ReadFile(args[0])
				if err != nil {
					return err
				}
				if err := os.MkdirAll(app.loaded.Paths.AgentsDir, 0o700); err != nil {
					return err
				}
				target := filepath.Join(app.loaded.Paths.AgentsDir, role.Name+".yaml")
				if err := os.WriteFile(target, data, 0o600); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "added: %s\n", role.Name)
				return nil
			},
		},
		&cobra.Command{
			Use:   "delete <name>",
			Short: "Delete a custom agent role",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				target := filepath.Join(app.loaded.Paths.AgentsDir, args[0]+".yaml")
				if err := os.Remove(target); err != nil {
					if os.IsNotExist(err) {
						return fmt.Errorf("CFG-003: custom role %s not found", args[0])
					}
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "deleted: %s\n", args[0])
				return nil
			},
		},
	)
	preset := &cobra.Command{
		Use:   "preset",
		Short: "Manage agent presets",
	}
	preset.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List presets",
			RunE: func(cmd *cobra.Command, args []string) error {
				registry, err := agents.NewRegistry(app.loaded.Paths.AgentsDir)
				if err != nil {
					return err
				}
				for _, item := range registry.ListPresets() {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", item.Name, item.DisplayName, strings.Join(item.Roles, ","))
				}
				return nil
			},
		},
		&cobra.Command{
			Use:   "show <name>",
			Short: "Show preset",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				registry, err := agents.NewRegistry(app.loaded.Paths.AgentsDir)
				if err != nil {
					return err
				}
				preset, ok := registry.Preset(args[0])
				if !ok {
					return fmt.Errorf("CFG-003: preset %s not found", args[0])
				}
				fmt.Fprintf(cmd.OutOrStdout(), "name: %s\n", preset.Name)
				fmt.Fprintf(cmd.OutOrStdout(), "display_name: %s\n", preset.DisplayName)
				fmt.Fprintf(cmd.OutOrStdout(), "roles: %s\n", strings.Join(preset.Roles, ", "))
				return nil
			},
		},
	)
	cmd.AddCommand(preset)
	return cmd
}

func printRole(cmd *cobra.Command, role agents.Role) {
	fmt.Fprintf(cmd.OutOrStdout(), "name: %s\n", role.Name)
	fmt.Fprintf(cmd.OutOrStdout(), "display_name: %s\n", role.DisplayName)
	fmt.Fprintf(cmd.OutOrStdout(), "model: %s\n", role.Model)
	fmt.Fprintf(cmd.OutOrStdout(), "stance: %s\n", role.Stance)
	if role.Temperature != 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "temperature: %g\n", role.Temperature)
	}
	if role.MaxTokens != 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "max_tokens: %d\n", role.MaxTokens)
	}
	if len(role.FocusAreas) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "focus_areas: %s\n", strings.Join(role.FocusAreas, ", "))
	}
	fmt.Fprintf(cmd.OutOrStdout(), "system_prompt: %s\n", role.SystemPrompt)
}
