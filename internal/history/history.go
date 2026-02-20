package history

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"
)

const (
	fileName   = "history.json"
	maxEntries = 100
)

// Entry represents a single history record.
type Entry struct {
	Query       string    `json:"query"`
	Selected    string    `json:"selected"`
	PipeContext string    `json:"pipe_context,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// Store provides read/write access to the history file.
type Store struct {
	filePath string
}

// NewStore creates a Store that persists history in the given directory.
func NewStore(dir string) *Store {
	return &Store{filePath: filepath.Join(dir, fileName)}
}

// Add appends an entry, rotates to the last maxEntries, and writes atomically.
func (s *Store) Add(entry Entry) error {
	entries, err := s.readAll()
	if err != nil {
		return err
	}

	entries = append(entries, entry)

	if len(entries) > maxEntries {
		entries = entries[len(entries)-maxEntries:]
	}

	return s.writeAll(entries)
}

// Last returns the most recent entry.
func (s *Store) Last() (Entry, error) {
	entries, err := s.readAll()
	if err != nil {
		return Entry{}, err
	}
	if len(entries) == 0 {
		return Entry{}, ErrEmpty
	}
	return entries[len(entries)-1], nil
}

// List returns all entries, newest first.
func (s *Store) List() ([]Entry, error) {
	entries, err := s.readAll()
	if err != nil {
		return nil, err
	}
	slices.Reverse(entries)
	return entries, nil
}

// ErrEmpty is returned when history has no entries.
var ErrEmpty = errors.New("history is empty")

func (s *Store) readAll() ([]Entry, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading history: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parsing history: %w", err)
	}
	return entries, nil
}

func (s *Store) writeAll(entries []Entry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling history: %w", err)
	}

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating history directory: %w", err)
	}

	tmp := s.filePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("writing history temp file: %w", err)
	}

	if err := os.Rename(tmp, s.filePath); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("renaming history temp file: %w", err)
	}

	return nil
}
