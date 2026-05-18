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
		newAgentsListCommand(app),
		newAgentsShowCommand(app),
		newAgentsAddCommand(app),
		newAgentsDeleteCommand(app),
		newAgentsValidateCommand(app),
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

func newAgentsListCommand(app *App) *cobra.Command {
	return &cobra.Command{
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
	}
}

func newAgentsShowCommand(app *App) *cobra.Command {
	return &cobra.Command{
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
	}
}

func newAgentsAddCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "add <role-yaml>",
		Short: "Add a custom agent role",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			role, err := agents.LoadRole(args[0])
			if err != nil {
				return err
			}
			if err := os.MkdirAll(app.loaded.Paths.AgentsDir, 0o700); err != nil {
				return err
			}
			data, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			path := customRolePath(app.loaded.Paths.AgentsDir, role.Name)
			if err := os.WriteFile(path, data, 0o600); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "agent added: %s\npath: %s\n", role.Name, path)
			return nil
		},
	}
}

func newAgentsDeleteCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a custom agent role",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := agents.ValidateRoleName(args[0]); err != nil {
				return err
			}
			path := customRolePath(app.loaded.Paths.AgentsDir, args[0])
			if err := os.Remove(path); err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("CFG-003: custom role %s not found", args[0])
				}
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "agent deleted: %s\npath: %s\n", args[0], path)
			return nil
		},
	}
}

func newAgentsValidateCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "validate [role-yaml...]",
		Short: "Validate custom agent role files",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := args
			if len(paths) == 0 {
				matches, err := filepath.Glob(filepath.Join(app.loaded.Paths.AgentsDir, "*.yaml"))
				if err != nil {
					return err
				}
				paths = matches
			}
			for _, path := range paths {
				role, err := agents.LoadRole(path)
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "valid: %s (%s)\n", role.Name, path)
			}
			if len(paths) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no custom roles found")
			}
			return nil
		},
	}
}

func printRole(cmd *cobra.Command, role agents.Role) {
	fmt.Fprintf(cmd.OutOrStdout(), "name: %s\n", role.Name)
	fmt.Fprintf(cmd.OutOrStdout(), "display_name: %s\n", role.DisplayName)
	fmt.Fprintf(cmd.OutOrStdout(), "model: %s\n", role.Model)
	fmt.Fprintf(cmd.OutOrStdout(), "stance: %s\n", role.Stance)
	fmt.Fprintf(cmd.OutOrStdout(), "temperature: %.2f\n", role.Temperature)
	fmt.Fprintf(cmd.OutOrStdout(), "max_tokens: %d\n", role.MaxTokens)
	if len(role.FocusAreas) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "focus_areas: %s\n", strings.Join(role.FocusAreas, ", "))
	}
	fmt.Fprintf(cmd.OutOrStdout(), "system_prompt: %s\n", role.SystemPrompt)
	if role.ResponseTemplate != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "response_template: %s\n", role.ResponseTemplate)
	}
	if role.Color != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "color: %s\n", role.Color)
	}
	if role.FallbackModel != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "fallback_model: %s\n", role.FallbackModel)
	}
}

func customRolePath(dir, name string) string {
	return filepath.Join(dir, name+".yaml")
}
