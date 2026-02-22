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
		{name: "case sensitive Fish", shell: "Fish", wantErr: true},
		{name: "case sensitive BASH", shell: "BASH", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, err := Script(tt.shell)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if script != "" {
					t.Errorf("expected empty script on error, got: %q", script)
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

func TestScriptContent(t *testing.T) {
	tests := []struct {
		shell         string
		expectedParts []string
	}{
		{"bash", []string{"READLINE_LINE", "bind -x", "QX_PATH", "--query"}},
		{"zsh", []string{"LBUFFER", "bindkey", "QX_PATH", "--query"}},
		{"fish", []string{"__qx_widget", "commandline -r", "\\cg", "QX_PATH", "--query", "string collect", "pipestatus"}},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			script, err := Script(tt.shell)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, part := range tt.expectedParts {
				if !strings.Contains(script, part) {
					t.Errorf("%s script should contain %q", tt.shell, part)
				}
			}
		})
	}
}

// TestBashSplitCondition verifies bash script uses split exit code handling:
// exit 0 always updates buffer (even empty result), exit 130 only if non-empty.
func TestBashSplitCondition(t *testing.T) {
	script, err := Script("bash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// exit 0 branch must not require -n check
	if !strings.Contains(script, `$exit_code -eq 0`) {
		t.Error("bash script should contain separate exit 0 check")
	}

	// exit 0 condition line must not include -n (the whole point of the fix)
	for _, line := range strings.Split(script, "\n") {
		if strings.Contains(line, "$exit_code -eq 0") && !strings.Contains(line, "130") {
			if strings.Contains(line, "-n") {
				t.Error("exit 0 condition must not include -n check — buffer should be set unconditionally")
			}
		}
	}

	// exit 130 branch must require -n "$result"
	if !strings.Contains(script, `$exit_code -eq 130 && -n "$result"`) {
		t.Error("bash script should check -n for exit 130 only")
	}

	// old combined condition must not be present
	if strings.Contains(script, `($exit_code -eq 0 || $exit_code -eq 130) && -n "$result"`) {
		t.Error("bash script should not use combined condition with -n check for both exit codes")
	}
}

// TestZshSplitConditionAndInvalidate verifies zsh script uses split exit code
// handling and calls zle -I before zle reset-prompt.
func TestZshSplitConditionAndInvalidate(t *testing.T) {
	script, err := Script("zsh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(script, `$exit_code -eq 0`) {
		t.Error("zsh script should contain separate exit 0 check")
	}

	for _, line := range strings.Split(script, "\n") {
		if strings.Contains(line, "$exit_code -eq 0") && !strings.Contains(line, "130") {
			if strings.Contains(line, "-n") {
				t.Error("exit 0 condition must not include -n check — buffer should be set unconditionally")
			}
		}
	}

	if !strings.Contains(script, `$exit_code -eq 130 && -n "$result"`) {
		t.Error("zsh script should check -n for exit 130 only")
	}

	if strings.Contains(script, `($exit_code -eq 0 || $exit_code -eq 130) && -n "$result"`) {
		t.Error("zsh script should not use combined condition with -n check for both exit codes")
	}

	// zle -I must appear before zle reset-prompt
	if !strings.Contains(script, "zle -I") {
		t.Error("zsh script should contain 'zle -I' for display invalidation")
	}

	iIdx := strings.Index(script, "zle -I")
	rpIdx := strings.Index(script, "zle reset-prompt")
	if iIdx >= rpIdx {
		t.Error("'zle -I' must appear before 'zle reset-prompt'")
	}
}

// TestFishSplitCondition verifies fish script uses split exit code handling.
func TestFishSplitCondition(t *testing.T) {
	script, err := Script("fish")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(script, `$exit_code -eq 0`) {
		t.Error("fish script should contain separate exit 0 check")
	}

	for _, line := range strings.Split(script, "\n") {
		if strings.Contains(line, "$exit_code -eq 0") && !strings.Contains(line, "130") {
			if strings.Contains(line, "-n") {
				t.Error("exit 0 condition must not include -n check — buffer should be set unconditionally")
			}
		}
	}

	if !strings.Contains(script, `$exit_code -eq 130 -a -n "$result"`) {
		t.Error("fish script should check -n for exit 130 only")
	}

	// old combined condition must not be present
	if strings.Contains(script, `$exit_code -eq 0 -o $exit_code -eq 130`) {
		t.Error("fish script should not use combined condition for both exit codes")
	}
}
