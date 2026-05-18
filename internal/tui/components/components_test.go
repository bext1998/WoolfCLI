package components

import (
	"strings"
	"testing"
)

func TestAgentBadge(t *testing.T) {
	result := AgentBadge("Strict Editor", "#d14d41")
	if result == "" {
		t.Error("badge should not be empty")
	}
	if !strings.Contains(result, "Strict Editor") {
		t.Error("badge should contain the name")
	}
}

func TestAgentBadgeEmptyColor(t *testing.T) {
	result := AgentBadge("Name", "")
	if result == "" {
		t.Error("badge with empty color should not be empty")
	}
}

func TestStanceTag(t *testing.T) {
	tests := []struct {
		stance string
		want   string
	}{
		{"agree", "[agree]"},
		{"disagree", "[disagree]"},
		{"extend", "[extend]"},
		{"neutral", "[neutral]"},
		{"", ""},
	}

	for _, tt := range tests {
		result := StanceTag(tt.stance)
		if tt.stance == "" && result != "" {
			t.Errorf("StanceTag(%q) should be empty, got %q", tt.stance, result)
		}
		if tt.stance != "" && !strings.Contains(result, tt.want) {
			t.Errorf("StanceTag(%q) missing %q in %q", tt.stance, tt.want, result)
		}
	}
}

func TestCostMeter(t *testing.T) {
	result := CostMeter(1000, 0.035)
	if result == "" {
		t.Error("cost meter should not be empty")
	}
	if !strings.Contains(result, "1000") {
		t.Error("cost meter should contain token count")
	}
	if !strings.Contains(result, "$0.0350") {
		t.Error("cost meter should contain cost")
	}
}

func TestProgress(t *testing.T) {
	tests := []struct {
		state string
		want  string
	}{
		{"idle", "Idle"},
		{"running", "Live"},
		{"paused", "Paused"},
		{"done", "Done"},
		{"error", "Error"},
	}

	for _, tt := range tests {
		result := Progress(tt.state)
		if !strings.Contains(result, tt.want) {
			t.Errorf("Progress(%q) missing %q in %q", tt.state, tt.want, result)
		}
	}
}
