package views

import "testing"

func TestParseSlashCommand(t *testing.T) {
	tests := []struct {
		input    string
		wantCmd  string
		wantArgs string
		wantOK   bool
	}{
		{"/start", "start", "", true},
		{"/start hello world", "start", "hello world", true},
		{"/next", "next", "", true},
		{"/end", "end", "", true},
		{"/pause", "pause", "", true},
		{"/quit", "quit", "", true},
		{"/help", "help", "", true},
		{"/status", "status", "", true},
		{"/agents", "agents", "", true},
		{"/cost", "cost", "", true},
		{"/focus 12-18", "focus", "12-18", true},
		{"/skip strict-editor", "skip", "strict-editor", true},
		{"/summarize", "summarize", "", true},
		{"/export md", "export", "md", true},
		{"   /start   ", "start", "", true},
		{"", "", "", false},
		{"hello", "", "", false},
		{"not a command", "", "", false},
		{"/", "", "", false},
		{"/with spaces", "with", "spaces", true},
	}

	for _, tt := range tests {
		cmd, ok := ParseSlashCommand(tt.input)
		if ok != tt.wantOK {
			t.Errorf("ParseSlashCommand(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			continue
		}
		if !ok {
			continue
		}
		if cmd.Command != tt.wantCmd {
			t.Errorf("ParseSlashCommand(%q).Command = %q, want %q", tt.input, cmd.Command, tt.wantCmd)
		}
		if cmd.Args != tt.wantArgs {
			t.Errorf("ParseSlashCommand(%q).Args = %q, want %q", tt.input, cmd.Args, tt.wantArgs)
		}
	}
}
