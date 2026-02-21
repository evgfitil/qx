package action

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mattn/go-isatty"
	"golang.org/x/term"
)

// ErrCancelled indicates the user cancelled without choosing an action.
var ErrCancelled = errors.New("action cancelled")

// ReviseRequestedError indicates the user wants to revise the selected command.
type ReviseRequestedError struct{}

func (e *ReviseRequestedError) Error() string {
	return "revise requested"
}

// Action represents a post-selection action chosen by the user.
type Action int

const (
	ActionExecute Action = iota
	ActionCopy
	ActionRevise
	ActionQuit
	ActionCancel
)

// ShouldPrompt returns true if stdout is a TTY, meaning the user is
// interacting directly with the terminal and should see the action menu.
func ShouldPrompt() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

// ShouldPromptStderr returns true if stderr is a TTY. This is used in shell
// integration mode where stdout is captured by $() but stderr is redirected
// to /dev/tty, allowing the action menu to render on stderr.
func ShouldPromptStderr() bool {
	return isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())
}

// readKeypress reads a single keypress from the given reader, which should
// be in raw mode. Handles multi-byte escape sequences by draining trailing
// bytes so they don't leak into the parent shell.
func readKeypress(r io.Reader) (Action, error) {
	buf := make([]byte, 1)
	for {
		if _, err := r.Read(buf); err != nil {
			return ActionQuit, fmt.Errorf("failed to read keypress: %w", err)
		}

		switch buf[0] {
		case 'e', 'E':
			return ActionExecute, nil
		case 'c', 'C':
			return ActionCopy, nil
		case 'r', 'R':
			return ActionRevise, nil
		case 'q', 'Q', '\r', '\n':
			return ActionQuit, nil
		case 0x03: // Ctrl+C
			return ActionCancel, nil
		case 0x1b: // Escape (may be start of multi-byte sequence)
			drainEscapeSequence(r)
			return ActionCancel, nil
		}
	}
}

// drainEscapeSequence reads and discards trailing bytes of a multi-byte
// escape sequence (e.g. arrow keys send \x1b[A — 3 bytes). Uses a short
// deadline when the reader supports it (e.g. *os.File on Unix).
func drainEscapeSequence(r io.Reader) {
	type deadliner interface {
		SetReadDeadline(t time.Time) error
	}
	if d, ok := r.(deadliner); ok {
		_ = d.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
		defer func() { _ = d.SetReadDeadline(time.Time{}) }()
	}
	discard := make([]byte, 8)
	_, _ = r.Read(discard)
}

// PromptAction displays the selected command and an action menu,
// then dispatches the chosen action. It reads input from /dev/tty
// to avoid conflicts with piped stdin.
func PromptAction(command string) error {
	return promptActionWith(command, nil)
}

// promptActionWith is the testable core of PromptAction. When ttyReader
// is nil, it opens /dev/tty and sets raw mode; otherwise it reads from
// the provided reader.
func promptActionWith(command string, ttyReader io.Reader) error {
	hi := "\033[38;5;205m"
	rs := "\033[0m"
	fmt.Fprintf(os.Stderr, "\n  %s\n\n  [%se%s]xecute  [%sc%s]opy  [%sr%s]evise  [%sq%s]uit ", command, hi, rs, hi, rs, hi, rs, hi, rs)

	act, err := readAction(ttyReader)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr)
	if act == ActionExecute {
		fmt.Fprintln(os.Stderr)
	}

	return dispatchAction(act, command)
}

// readAction reads a single-keypress action from the given reader or /dev/tty.
func readAction(ttyReader io.Reader) (Action, error) {
	if ttyReader != nil {
		return readKeypress(ttyReader)
	}

	tty, err := os.Open("/dev/tty")
	if err != nil {
		return ActionQuit, fmt.Errorf("failed to open /dev/tty: %w", err)
	}
	defer func() { _ = tty.Close() }()

	oldState, err := term.MakeRaw(int(tty.Fd()))
	if err != nil {
		return ActionQuit, fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer func() { _ = term.Restore(int(tty.Fd()), oldState) }()

	return readKeypress(tty)
}

// dispatchAction executes the chosen action on the command.
func dispatchAction(act Action, command string) error {
	switch act {
	case ActionExecute:
		return Execute(command)
	case ActionCopy:
		if err := CopyToClipboard(command); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Copied to clipboard.")
		return nil
	case ActionRevise:
		return &ReviseRequestedError{}
	case ActionQuit:
		fmt.Println(command)
		return nil
	case ActionCancel:
		return ErrCancelled
	default:
		return fmt.Errorf("unknown action: %d", act)
	}
}
