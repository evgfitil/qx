package stdin

import (
	"bytes"
	"strings"
	"testing"
)

func TestReader_Read_WithPipedContent(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "simple content",
			input:   "hello world",
			want:    "hello world",
			wantErr: false,
		},
		{
			name:    "multiline content",
			input:   "line1\nline2\nline3",
			want:    "line1\nline2\nline3",
			wantErr: false,
		},
		{
			name:    "empty content",
			input:   "",
			want:    "",
			wantErr: false,
		},
		{
			name:    "content with special characters",
			input:   "docker ps | grep nginx",
			want:    "docker ps | grep nginx",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := New(strings.NewReader(tt.input))
			got, err := reader.Read()

			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("Read() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReader_IsPiped_WithNonFile(t *testing.T) {
	reader := New(bytes.NewBufferString("test"))

	if !reader.IsPiped() {
		t.Error("IsPiped() should return true for non-file reader")
	}
}

func TestReader_IsPiped_WithStringsReader(t *testing.T) {
	reader := New(strings.NewReader("test"))

	if !reader.IsPiped() {
		t.Error("IsPiped() should return true for strings.Reader")
	}
}
