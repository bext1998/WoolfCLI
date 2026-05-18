package agents

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuiltinRegistryResolvesPresets(t *testing.T) {
	registry, err := NewRegistry("")
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	presets := registry.ListPresets()
	if len(presets) < 3 {
		t.Fatalf("builtin presets = %d, want at least 3", len(presets))
	}
	roles, err := registry.ResolvePreset("editorial")
	if err != nil {
		t.Fatalf("ResolvePreset() error = %v", err)
	}
	if len(roles) != 3 {
		t.Fatalf("editorial role count = %d, want 3", len(roles))
	}
	if _, err := registry.ResolvePreset("critique"); err != nil {
		t.Fatalf("ResolvePreset(critique) error = %v", err)
	}
	if _, err := registry.ResolvePreset("review"); err != nil {
		t.Fatalf("ResolvePreset(review) compatibility alias error = %v", err)
	}
	listed := registry.ListRoles()
	if len(listed) < 6 {
		t.Fatalf("builtin roles = %d, want at least 6", len(listed))
	}
	for i := 1; i < len(listed); i++ {
		if listed[i-1].Name > listed[i].Name {
			t.Fatalf("roles are not sorted: %s before %s", listed[i-1].Name, listed[i].Name)
		}
	}
}

func TestLoadUserRoleValidation(t *testing.T) {
	dir := t.TempDir()
	rolePath := filepath.Join(dir, "custom.yaml")
	if err := os.WriteFile(rolePath, []byte(`
name: custom
display_name: Custom
model: openai/gpt-4o-mini
stance: neutral
system_prompt: Test prompt.
`), 0o600); err != nil {
		t.Fatal(err)
	}
	registry, err := NewRegistry(dir)
	if err != nil {
		t.Fatalf("NewRegistry(user roles) error = %v", err)
	}
	if _, ok := registry.Role("custom"); !ok {
		t.Fatalf("custom role was not loaded")
	}
}

func TestRoleValidationRejectsBadStance(t *testing.T) {
	role := Role{Name: "bad", DisplayName: "Bad", Model: "model", SystemPrompt: "prompt", Stance: "maybe"}
	if err := role.Validate(); err == nil {
		t.Fatalf("Validate() error = nil, want invalid stance")
	}
}

func TestRoleValidationRejectsUnsafeNames(t *testing.T) {
	tests := []string{
		"../evil",
		"bad/name",
		"BadName",
		"bad_name",
	}
	for _, name := range tests {
		role := Role{Name: name, DisplayName: "Bad", Model: "model", SystemPrompt: "prompt", Stance: "neutral"}
		if err := role.Validate(); err == nil {
			t.Fatalf("Validate() error = nil for unsafe name %q", name)
		}
	}
}
