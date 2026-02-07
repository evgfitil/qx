package action

import (
	"testing"

	"github.com/atotto/clipboard"
)

func TestCopyToClipboard(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("clipboard not available in this environment")
	}

	const text = "echo hello world"
	if err := CopyToClipboard(text); err != nil {
		t.Fatalf("CopyToClipboard(%q) returned error: %v", text, err)
	}

	got, err := clipboard.ReadAll()
	if err != nil {
		t.Fatalf("clipboard.ReadAll() returned error: %v", err)
	}
	if got != text {
		t.Errorf("clipboard content = %q, want %q", got, text)
	}
}

func TestCopyToClipboard_EmptyString(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("clipboard not available in this environment")
	}

	if err := CopyToClipboard(""); err != nil {
		t.Fatalf("CopyToClipboard(\"\") returned error: %v", err)
	}

	got, err := clipboard.ReadAll()
	if err != nil {
		t.Fatalf("clipboard.ReadAll() returned error: %v", err)
	}
	if got != "" {
		t.Errorf("clipboard content = %q, want empty string", got)
	}
}

func TestCopyToClipboard_SpecialCharacters(t *testing.T) {
	if clipboard.Unsupported {
		t.Skip("clipboard not available in this environment")
	}

	const text = `docker ps -q | xargs -I{} docker stop {} && echo "done"`
	if err := CopyToClipboard(text); err != nil {
		t.Fatalf("CopyToClipboard(%q) returned error: %v", text, err)
	}

	got, err := clipboard.ReadAll()
	if err != nil {
		t.Fatalf("clipboard.ReadAll() returned error: %v", err)
	}
	if got != text {
		t.Errorf("clipboard content = %q, want %q", got, text)
	}
}
