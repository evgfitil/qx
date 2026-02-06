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
		{"bash", []string{"READLINE_LINE", "bind -x", "QX_PATH", "--query", "mktemp", "err_file", "2>\"$err_file\"", "cat \"$err_file\" >/dev/tty", "rm -f \"$err_file\""}},
		{"zsh", []string{"LBUFFER", "bindkey", "QX_PATH", "--query", "mktemp", "err_file", "2>\"$err_file\"", "cat \"$err_file\" >/dev/tty", "rm -f \"$err_file\""}},
		{"fish", []string{"__qx_widget", "commandline -r", "\\cg", "QX_PATH", "--query", "string collect", "pipestatus", "mktemp", "err_file", "2>$err_file", "cat $err_file >/dev/tty", "rm -f $err_file"}},
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
