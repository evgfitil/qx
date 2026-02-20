package history

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	return NewStore(dir)
}

func sampleEntry(query string) Entry {
	return Entry{
		Query:     query,
		Commands:  []string{"cmd1", "cmd2"},
		Selected:  "cmd1",
		Timestamp: time.Now().Truncate(time.Millisecond),
	}
}

func TestNewStore(t *testing.T) {
	s := NewStore("/tmp/qx-test")
	want := filepath.Join("/tmp/qx-test", fileName)
	if s.filePath != want {
		t.Errorf("filePath = %q, want %q", s.filePath, want)
	}
}

func TestAdd_EmptyFile(t *testing.T) {
	s := tempStore(t)
	entry := sampleEntry("list files")

	if err := s.Add(entry); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	entries, err := s.readAll()
	if err != nil {
		t.Fatalf("readAll() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].Query != "list files" {
		t.Errorf("query = %q, want %q", entries[0].Query, "list files")
	}
}

func TestAdd_Append(t *testing.T) {
	s := tempStore(t)

	if err := s.Add(sampleEntry("first")); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if err := s.Add(sampleEntry("second")); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	entries, err := s.readAll()
	if err != nil {
		t.Fatalf("readAll() error = %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
	if entries[0].Query != "first" || entries[1].Query != "second" {
		t.Errorf("entries order: [%q, %q], want [first, second]", entries[0].Query, entries[1].Query)
	}
}

func TestAdd_RotationAtLimit(t *testing.T) {
	s := tempStore(t)

	for i := range maxEntries + 10 {
		if err := s.Add(sampleEntry("query-" + string(rune('A'+i%26)))); err != nil {
			t.Fatalf("Add() error on iteration %d: %v", i, err)
		}
	}

	entries, err := s.readAll()
	if err != nil {
		t.Fatalf("readAll() error = %v", err)
	}
	if len(entries) != maxEntries {
		t.Errorf("got %d entries, want %d", len(entries), maxEntries)
	}
}

func TestAdd_PipeContext(t *testing.T) {
	s := tempStore(t)
	entry := Entry{
		Query:       "delete old files",
		Commands:    []string{"find . -mtime +30 -delete"},
		Selected:    "find . -mtime +30 -delete",
		PipeContext: "ls -la output",
		Timestamp:   time.Now(),
	}

	if err := s.Add(entry); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	got, err := s.Last()
	if err != nil {
		t.Fatalf("Last() error = %v", err)
	}
	if got.PipeContext != "ls -la output" {
		t.Errorf("PipeContext = %q, want %q", got.PipeContext, "ls -la output")
	}
}

func TestAdd_AtomicWrite(t *testing.T) {
	s := tempStore(t)
	if err := s.Add(sampleEntry("test")); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	tmp := s.filePath + ".tmp"
	if _, err := os.Stat(tmp); !os.IsNotExist(err) {
		t.Error("temp file should not exist after successful write")
	}
}

func TestLast_EmptyHistory(t *testing.T) {
	s := tempStore(t)

	_, err := s.Last()
	if err != ErrEmpty {
		t.Errorf("Last() error = %v, want ErrEmpty", err)
	}
}

func TestLast_SingleEntry(t *testing.T) {
	s := tempStore(t)
	entry := sampleEntry("only one")
	if err := s.Add(entry); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	got, err := s.Last()
	if err != nil {
		t.Fatalf("Last() error = %v", err)
	}
	if got.Query != "only one" {
		t.Errorf("query = %q, want %q", got.Query, "only one")
	}
}

func TestLast_MultipleEntries(t *testing.T) {
	s := tempStore(t)

	for _, q := range []string{"first", "second", "third"} {
		if err := s.Add(sampleEntry(q)); err != nil {
			t.Fatalf("Add() error = %v", err)
		}
	}

	got, err := s.Last()
	if err != nil {
		t.Fatalf("Last() error = %v", err)
	}
	if got.Query != "third" {
		t.Errorf("query = %q, want %q", got.Query, "third")
	}
}

func TestList_Empty(t *testing.T) {
	s := tempStore(t)

	entries, err := s.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("got %d entries, want 0", len(entries))
	}
}

func TestList_Ordering(t *testing.T) {
	s := tempStore(t)

	for _, q := range []string{"first", "second", "third"} {
		if err := s.Add(sampleEntry(q)); err != nil {
			t.Fatalf("Add() error = %v", err)
		}
	}

	entries, err := s.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(entries))
	}

	want := []string{"third", "second", "first"}
	for i, e := range entries {
		if e.Query != want[i] {
			t.Errorf("entries[%d].Query = %q, want %q", i, e.Query, want[i])
		}
	}
}

func TestReadAll_CorruptedFile(t *testing.T) {
	s := tempStore(t)
	if err := os.WriteFile(s.filePath, []byte("not json"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := s.readAll()
	if err == nil {
		t.Error("readAll() expected error for corrupted file")
	}
}

func TestReadAll_EmptyFile(t *testing.T) {
	s := tempStore(t)
	if err := os.WriteFile(s.filePath, []byte(""), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	entries, err := s.readAll()
	if err != nil {
		t.Fatalf("readAll() error = %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("got %d entries, want 0", len(entries))
	}
}
