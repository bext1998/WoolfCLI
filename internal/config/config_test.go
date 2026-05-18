package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAppliesEnvironmentOverrides(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	sessionsDir := filepath.Join(dir, "sessions")
	if err := os.WriteFile(configPath, []byte(`
[api]
openrouter_key = "file-key"

[paths]
sessions_dir = "file-sessions"
`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OPENROUTER_API_KEY", "env-key")
	t.Setenv("WOOLF_SESSIONS_DIR", sessionsDir)

	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Config.API.OpenRouterKey != "env-key" {
		t.Fatalf("OpenRouterKey = %q, want env override", loaded.Config.API.OpenRouterKey)
	}
	if loaded.Paths.SessionsDir != sessionsDir {
		t.Fatalf("SessionsDir = %q, want %q", loaded.Paths.SessionsDir, sessionsDir)
	}
}

func TestMaskSecret(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{value: "", want: ""},
		{value: "short", want: "****"},
		{value: "sk-or-secret-value", want: "sk-or-se****"},
	}
	for _, tt := range tests {
		if got := MaskSecret(tt.value); got != tt.want {
			t.Fatalf("MaskSecret(%q) = %q, want %q", tt.value, got, tt.want)
		}
	}
}
