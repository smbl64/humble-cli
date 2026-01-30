package util

import (
	"testing"
)

func TestReplaceInvalidCharsInFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal_file.txt", "normal_file.txt"},
		{"file/with/slashes", "file with slashes"},
		{"file\\with\\backslashes", "file with backslashes"},
		{"file?with*special:chars", "file with special chars"},
		{"file<with>pipes|quotes\"", "file with pipes quotes"},
		{"  spaces  around  ", "spaces  around"},
		{"file\nwith\nnewlines", "file with newlines"},
	}

	for _, tt := range tests {
		result := ReplaceInvalidCharsInFilename(tt.input)
		if result != tt.expected {
			t.Errorf("ReplaceInvalidCharsInFilename(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestExtractFilenameFromURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://example.com/path/to/file.pdf", "file.pdf"},
		{"https://example.com/file.epub", "file.epub"},
		{"https://example.com/path/", "path"},
		{"https://example.com", ""},
		{"invalid-url", "invalid-url"},
		{"https://example.com/path/to/file.pdf?query=param", "file.pdf"},
	}

	for _, tt := range tests {
		result := ExtractFilenameFromURL(tt.input)
		if result != tt.expected {
			t.Errorf("ExtractFilenameFromURL(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}
