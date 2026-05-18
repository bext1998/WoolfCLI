package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentsListAndShowRoles(t *testing.T) {
	dir := t.TempDir()
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", filepath.Join(dir, "config.toml"), "agents", "list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("agents list error = %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "strict-editor") {
		t.Fatalf("agents list output missing strict-editor:\n%s", out.String())
	}

	cmd = NewRootCommand()
	out.Reset()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", filepath.Join(dir, "config.toml"), "agents", "show", "strict-editor"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("agents show error = %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "name: strict-editor") || !strings.Contains(out.String(), "model: openai/gpt-4o-mini") {
		t.Fatalf("agents show output missing role details:\n%s", out.String())
	}
}

func TestAgentsAddAndDeleteCustomRole(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	configPath := filepath.Join(dir, "config.toml")
	config := "[paths]\nagents_dir = " + quotedTomlString(agentsDir) + "\n"
	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	rolePath := filepath.Join(dir, "custom.yaml")
	if err := os.WriteFile(rolePath, []byte(`
note: keep-me
name: custom
display_name: Custom
model: openai/gpt-4o-mini
stance: neutral
system_prompt: Test prompt.
`), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", configPath, "agents", "add", rolePath})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("agents add error = %v\n%s", err, out.String())
	}
	if _, err := os.Stat(filepath.Join(agentsDir, "custom.yaml")); err != nil {
		t.Fatalf("custom role was not written: %v", err)
	}
	written, err := os.ReadFile(filepath.Join(agentsDir, "custom.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(written), "note: keep-me") {
		t.Fatalf("custom role file did not preserve source YAML:\n%s", string(written))
	}

	cmd = NewRootCommand()
	out.Reset()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", configPath, "agents", "show", "custom"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("agents show custom error = %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "name: custom") {
		t.Fatalf("agents show custom output missing role:\n%s", out.String())
	}

	cmd = NewRootCommand()
	out.Reset()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", configPath, "agents", "delete", "custom"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("agents delete error = %v\n%s", err, out.String())
	}
	if _, err := os.Stat(filepath.Join(agentsDir, "custom.yaml")); !os.IsNotExist(err) {
		t.Fatalf("custom role still exists, err = %v", err)
	}
}

func TestAgentsDeleteRejectsUnsafeName(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--config", configPath, "agents", "delete", "../custom"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "CFG-003") {
		t.Fatalf("agents delete error = %v, want CFG-003", err)
	}
}

func quotedTomlString(value string) string {
	return `"` + strings.ReplaceAll(value, `\`, `\\`) + `"`
}
