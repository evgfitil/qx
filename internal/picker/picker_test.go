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

func TestPickIndex_Empty(t *testing.T) {
	_, err := PickIndex(0, func(i int) string { return "" })
	if err == nil {
		t.Error("expected error for zero items")
	}
}

func TestPickIndex_NegativeCount(t *testing.T) {
	_, err := PickIndex(-1, func(i int) string { return "" })
	if err == nil {
		t.Error("expected error for negative count")
	}
}
