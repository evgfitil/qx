package shell

import (
	"strings"
	"testing"
)

func TestScript(t *testing.T) {
	tests := []struct {
		name      string
		shell     string
		wantErr   bool
		wantCheck func(script string) bool
	}{
		{
			name:    "bash returns script with bind",
			shell:   "bash",
			wantErr: false,
			wantCheck: func(script string) bool {
				return strings.Contains(script, "__qx_widget") &&
					strings.Contains(script, "bind -x") &&
					strings.Contains(script, "READLINE_LINE")
			},
		},
		{
			name:    "zsh returns script with bindkey",
			shell:   "zsh",
			wantErr: false,
			wantCheck: func(script string) bool {
				return strings.Contains(script, "__qx_widget") &&
					strings.Contains(script, "bindkey") &&
					strings.Contains(script, "LBUFFER")
			},
		},
		{
			name:    "fish returns script with bind",
			shell:   "fish",
			wantErr: false,
			wantCheck: func(script string) bool {
				return strings.Contains(script, "__qx_widget") &&
					strings.Contains(script, "bind \\cg") &&
					strings.Contains(script, "commandline")
			},
		},
		{
			name:    "unsupported shell returns error",
			shell:   "powershell",
			wantErr: true,
		},
		{
			name:    "empty shell returns error",
			shell:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, err := Script(tt.shell)
			if (err != nil) != tt.wantErr {
				t.Errorf("Script(%q) error = %v, wantErr %v", tt.shell, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if script == "" {
					t.Errorf("Script(%q) returned empty script", tt.shell)
				}
				if tt.wantCheck != nil && !tt.wantCheck(script) {
					t.Errorf("Script(%q) content check failed, got:\n%s", tt.shell, script)
				}
			}
		})
	}
}

func TestScriptErrorMessage(t *testing.T) {
	_, err := Script("invalid")
	if err == nil {
		t.Fatal("expected error for invalid shell")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "bash") ||
		!strings.Contains(errMsg, "zsh") ||
		!strings.Contains(errMsg, "fish") {
		t.Errorf("error message should list all supported shells, got: %s", errMsg)
	}
}
