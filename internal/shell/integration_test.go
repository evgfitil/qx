package shell

import (
	"strings"
	"testing"
)

func TestScript(t *testing.T) {
	tests := []struct {
		name    string
		shell   string
		wantErr bool
	}{
		{name: "bash returns script", shell: "bash"},
		{name: "zsh returns script", shell: "zsh"},
		{name: "fish returns script", shell: "fish"},
		{name: "unsupported shell returns error", shell: "powershell", wantErr: true},
		{name: "empty shell returns error", shell: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, err := Script(tt.shell)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), "unsupported shell") {
					t.Errorf("error should mention 'unsupported shell', got: %s", err.Error())
				}
				if !strings.Contains(err.Error(), "fish") {
					t.Errorf("error should list fish as supported, got: %s", err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if script == "" {
				t.Fatal("expected non-empty script")
			}
		})
	}
}

func TestScriptBashContent(t *testing.T) {
	script, err := Script("bash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedParts := []string{"READLINE_LINE", `bind -x`}
	for _, part := range expectedParts {
		if !strings.Contains(script, part) {
			t.Errorf("bash script should contain %q", part)
		}
	}
}

func TestScriptZshContent(t *testing.T) {
	script, err := Script("zsh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedParts := []string{"LBUFFER", "bindkey"}
	for _, part := range expectedParts {
		if !strings.Contains(script, part) {
			t.Errorf("zsh script should contain %q", part)
		}
	}
}

func TestScriptFishContent(t *testing.T) {
	script, err := Script("fish")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedParts := []string{"bind", "commandline", "function", "end"}
	for _, part := range expectedParts {
		if !strings.Contains(script, part) {
			t.Errorf("fish script should contain %q", part)
		}
	}
}
