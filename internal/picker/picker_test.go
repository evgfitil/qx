package picker

import "testing"

func TestPick_EmptyCommands(t *testing.T) {
	_, err := Pick([]string{})
	if err == nil {
		t.Error("expected error for empty commands")
	}
}

func TestPick_SingleCommand(t *testing.T) {
	cmd, err := Pick([]string{"ls -la"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cmd != "ls -la" {
		t.Errorf("expected 'ls -la', got '%s'", cmd)
	}
}

func TestPick_NilCommands(t *testing.T) {
	_, err := Pick(nil)
	if err == nil {
		t.Error("expected error for nil commands")
	}
}
