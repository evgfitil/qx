package ui

import "testing"

func TestSelectedResultImplementsResult(t *testing.T) {
	// Compile-time check that SelectedResult implements Result.
	var _ Result = SelectedResult{Command: "ls -la", Query: "list files"}
}

func TestCancelledResultImplementsResult(t *testing.T) {
	// Compile-time check that CancelledResult implements Result.
	var _ Result = CancelledResult{Query: "list files"}
}

func TestSelectedResultFields(t *testing.T) {
	r := SelectedResult{Command: "ls -la", Query: "list files"}

	if r.Command != "ls -la" {
		t.Errorf("Command = %q, want %q", r.Command, "ls -la")
	}
	if r.Query != "list files" {
		t.Errorf("Query = %q, want %q", r.Query, "list files")
	}
}

func TestCancelledResultFields(t *testing.T) {
	r := CancelledResult{Query: "list files"}

	if r.Query != "list files" {
		t.Errorf("Query = %q, want %q", r.Query, "list files")
	}
}

func TestSelectedResultEmptyFields(t *testing.T) {
	r := SelectedResult{}

	if r.Command != "" {
		t.Errorf("Command = %q, want empty", r.Command)
	}
	if r.Query != "" {
		t.Errorf("Query = %q, want empty", r.Query)
	}
}

func TestCancelledResultEmptyQuery(t *testing.T) {
	r := CancelledResult{}

	if r.Query != "" {
		t.Errorf("Query = %q, want empty", r.Query)
	}
}

func TestModelResultSelected(t *testing.T) {
	m := Model{selected: "ls -la", originalQuery: "list files"}
	r := m.Result()

	sr, ok := r.(SelectedResult)
	if !ok {
		t.Fatalf("Result type = %T, want SelectedResult", r)
	}
	if sr.Command != "ls -la" {
		t.Errorf("Command = %q, want %q", sr.Command, "ls -la")
	}
	if sr.Query != "list files" {
		t.Errorf("Query = %q, want %q", sr.Query, "list files")
	}
}

func TestModelResultCancelled(t *testing.T) {
	m := Model{originalQuery: "list files"}
	r := m.Result()

	cr, ok := r.(CancelledResult)
	if !ok {
		t.Fatalf("Result type = %T, want CancelledResult", r)
	}
	if cr.Query != "list files" {
		t.Errorf("Query = %q, want %q", cr.Query, "list files")
	}
}
