package util

import (
	"testing"
)

func TestHumanizeBytes(t *testing.T) {
	tests := []struct {
		input    uint64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{1048576, "1.0 MiB"},
		{1073741824, "1.0 GiB"},
		{1099511627776, "1.0 TiB"},
	}

	for _, tt := range tests {
		result := HumanizeBytes(tt.input)
		if result != tt.expected {
			t.Errorf("HumanizeBytes(%d) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestByteStringToNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected uint64
		hasError bool
	}{
		{"100", 100, false},
		{"1024", 1024, false},
		{"1KB", 1024, false},
		{"1K", 1024, false},
		{"1MB", 1048576, false},
		{"1M", 1048576, false},
		{"1GB", 1073741824, false},
		{"1G", 1073741824, false},
		{"500MB", 524288000, false},
		{"2.5GB", 2684354560, false},
		{"", 0, true},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		result, err := ByteStringToNumber(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("ByteStringToNumber(%s) expected error but got none", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("ByteStringToNumber(%s) unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("ByteStringToNumber(%s) = %d; want %d", tt.input, result, tt.expected)
			}
		}
	}
}
