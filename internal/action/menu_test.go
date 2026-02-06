package action

import (
	"bytes"
	"errors"
	"io"
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

func TestReadKeypress_UnknownKeyRetries(t *testing.T) {
	// Unknown key 'x' is ignored; readKeypress retries and reads 'q'.
	r := bytes.NewReader([]byte{'x', 'q'})
	act, err := readKeypress(r)
	if err != nil {
		t.Errorf("readKeypress('x','q') returned error: %v", err)
	}
	if act != ActionQuit {
		t.Errorf("readKeypress('x','q') = %d, want ActionQuit(%d)", act, ActionQuit)
	}
}

func TestReadKeypress_UnknownKeyOnlyReturnsError(t *testing.T) {
	// When only unknown keys are available, readKeypress eventually hits EOF.
	r := bytes.NewReader([]byte{'x'})
	_, err := readKeypress(r)
	if err == nil {
		t.Error("readKeypress('x' only) expected error on retry EOF, got nil")
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

	out, _ := io.ReadAll(r)
	_ = r.Close()
	if string(out) != "echo hello\n" {
		t.Errorf("dispatchAction(ActionQuit) output = %q, want %q", string(out), "echo hello\n")
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
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stderr pipe: %v", err)
	}
	defer func() { _ = stderrR.Close() }()
	os.Stderr = stderrW
	t.Cleanup(func() { os.Stderr = origStderr })

	input := bytes.NewReader([]byte{'q'})
	promptErr := promptActionWith("echo hello", input)
	_ = w.Close()
	_ = stderrW.Close()

	if promptErr != nil {
		t.Errorf("promptActionWith(\"echo hello\", 'q') returned error: %v", promptErr)
	}

	out, _ := io.ReadAll(r)
	_ = r.Close()
	if string(out) != "echo hello\n" {
		t.Errorf("promptActionWith quit output = %q, want %q", string(out), "echo hello\n")
	}
}

func TestPromptActionWith_Execute(t *testing.T) {
	t.Setenv("SHELL", "/bin/sh")

	// Redirect stderr to discard prompt output
	origStderr := os.Stderr
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stderr pipe: %v", err)
	}
	defer func() { _ = stderrR.Close() }()
	os.Stderr = stderrW
	t.Cleanup(func() { os.Stderr = origStderr })

	input := bytes.NewReader([]byte{'e'})
	promptErr := promptActionWith("true", input)
	_ = stderrW.Close()

	if promptErr != nil {
		t.Errorf("promptActionWith(\"true\", 'e') returned error: %v", promptErr)
	}
}

func TestReadKeypress_Cancel(t *testing.T) {
	for _, key := range []byte{0x03, 0x1b} {
		r := bytes.NewReader([]byte{key})
		act, err := readKeypress(r)
		if err != nil {
			t.Errorf("readKeypress(0x%02x) returned error: %v", key, err)
		}
		if act != ActionCancel {
			t.Errorf("readKeypress(0x%02x) = %d, want ActionCancel(%d)", key, act, ActionCancel)
		}
	}
}

func TestReadKeypress_EscapeSequence(t *testing.T) {
	// Arrow key sends \x1b[A (3 bytes). readKeypress should return ActionCancel
	// and drain the trailing bytes.
	r := bytes.NewReader([]byte{0x1b, '[', 'A'})
	act, err := readKeypress(r)
	if err != nil {
		t.Fatalf("readKeypress(escape sequence) returned error: %v", err)
	}
	if act != ActionCancel {
		t.Errorf("readKeypress(escape sequence) = %d, want ActionCancel(%d)", act, ActionCancel)
	}
}

func TestDispatchAction_Cancel(t *testing.T) {
	err := dispatchAction(ActionCancel, "echo hello")
	if !errors.Is(err, ErrCancelled) {
		t.Errorf("dispatchAction(ActionCancel) = %v, want ErrCancelled", err)
	}
}
