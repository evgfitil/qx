package action

import (
	"bytes"
	"os"
	"testing"
)

func TestShouldPrompt_WithPipe(t *testing.T) {
	// When stdout is redirected to a pipe, ShouldPrompt should return false.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	defer func() { _ = r.Close() }()
	defer func() { _ = w.Close() }()

	origStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	if ShouldPrompt() {
		t.Error("ShouldPrompt() = true for pipe stdout, want false")
	}
}

func TestReadKeypress_Execute(t *testing.T) {
	for _, key := range []byte{'e', 'E'} {
		r := bytes.NewReader([]byte{key})
		act, err := readKeypress(r)
		if err != nil {
			t.Errorf("readKeypress(%q) returned error: %v", key, err)
		}
		if act != ActionExecute {
			t.Errorf("readKeypress(%q) = %d, want ActionExecute(%d)", key, act, ActionExecute)
		}
	}
}

func TestReadKeypress_Copy(t *testing.T) {
	for _, key := range []byte{'c', 'C'} {
		r := bytes.NewReader([]byte{key})
		act, err := readKeypress(r)
		if err != nil {
			t.Errorf("readKeypress(%q) returned error: %v", key, err)
		}
		if act != ActionCopy {
			t.Errorf("readKeypress(%q) = %d, want ActionCopy(%d)", key, act, ActionCopy)
		}
	}
}

func TestReadKeypress_Quit(t *testing.T) {
	for _, key := range []byte{'q', 'Q', '\r', '\n'} {
		r := bytes.NewReader([]byte{key})
		act, err := readKeypress(r)
		if err != nil {
			t.Errorf("readKeypress(%q) returned error: %v", key, err)
		}
		if act != ActionQuit {
			t.Errorf("readKeypress(%q) = %d, want ActionQuit(%d)", key, act, ActionQuit)
		}
	}
}

func TestReadKeypress_UnknownKey(t *testing.T) {
	r := bytes.NewReader([]byte{'x'})
	act, err := readKeypress(r)
	if err != nil {
		t.Errorf("readKeypress('x') returned error: %v", err)
	}
	if act != ActionQuit {
		t.Errorf("readKeypress('x') = %d, want ActionQuit(%d)", act, ActionQuit)
	}
}

func TestReadKeypress_EmptyReader(t *testing.T) {
	r := bytes.NewReader(nil)
	_, err := readKeypress(r)
	if err == nil {
		t.Error("readKeypress(empty) expected error, got nil")
	}
}

func TestDispatchAction_Execute(t *testing.T) {
	t.Setenv("SHELL", "/bin/sh")

	err := dispatchAction(ActionExecute, "true")
	if err != nil {
		t.Errorf("dispatchAction(ActionExecute, \"true\") returned error: %v", err)
	}
}

func TestDispatchAction_Execute_Failure(t *testing.T) {
	t.Setenv("SHELL", "/bin/sh")

	err := dispatchAction(ActionExecute, "false")
	if err == nil {
		t.Error("dispatchAction(ActionExecute, \"false\") expected error, got nil")
	}
}

func TestDispatchAction_Quit(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	origStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	dispatchErr := dispatchAction(ActionQuit, "echo hello")
	_ = w.Close()

	if dispatchErr != nil {
		t.Errorf("dispatchAction(ActionQuit) returned error: %v", dispatchErr)
	}

	buf := make([]byte, 128)
	n, _ := r.Read(buf)
	got := string(buf[:n])
	if got != "echo hello\n" {
		t.Errorf("dispatchAction(ActionQuit) output = %q, want %q", got, "echo hello\n")
	}
}

func TestPromptActionWith_Quit(t *testing.T) {
	// Redirect stdout to capture output
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	origStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	// Redirect stderr to discard prompt output
	origStderr := os.Stderr
	_, stderrW, _ := os.Pipe()
	os.Stderr = stderrW
	t.Cleanup(func() { os.Stderr = origStderr })

	input := bytes.NewReader([]byte{'q'})
	promptErr := promptActionWith("echo hello", input)
	_ = w.Close()
	_ = stderrW.Close()

	if promptErr != nil {
		t.Errorf("promptActionWith(\"echo hello\", 'q') returned error: %v", promptErr)
	}

	buf := make([]byte, 128)
	n, _ := r.Read(buf)
	got := string(buf[:n])
	if got != "echo hello\n" {
		t.Errorf("promptActionWith quit output = %q, want %q", got, "echo hello\n")
	}
}

func TestPromptActionWith_Execute(t *testing.T) {
	t.Setenv("SHELL", "/bin/sh")

	// Redirect stderr to discard prompt output
	origStderr := os.Stderr
	_, stderrW, _ := os.Pipe()
	os.Stderr = stderrW
	t.Cleanup(func() { os.Stderr = origStderr })

	input := bytes.NewReader([]byte{'e'})
	err := promptActionWith("true", input)
	_ = stderrW.Close()

	if err != nil {
		t.Errorf("promptActionWith(\"true\", 'e') returned error: %v", err)
	}
}
