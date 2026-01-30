package util

import (
	"reflect"
	"testing"
)

func TestParseUsizeRange(t *testing.T) {
	tests := []struct {
		input    string
		maxValue int
		expected []int
		hasError bool
	}{
		{"5", 10, []int{5}, false},
		{"1-5", 10, []int{1, 2, 3, 4, 5}, false},
		{"-5", 10, []int{1, 2, 3, 4, 5}, false},
		{"10-", 15, []int{10, 11, 12, 13, 14, 15}, false},
		{"3-7", 10, []int{3, 4, 5, 6, 7}, false},
		{"invalid", 10, nil, true},
		{"1-invalid", 10, nil, true},
	}

	for _, tt := range tests {
		result, err := ParseUsizeRange(tt.input, tt.maxValue)
		if tt.hasError {
			if err == nil {
				t.Errorf("ParseUsizeRange(%s, %d) expected error but got none", tt.input, tt.maxValue)
			}
		} else {
			if err != nil {
				t.Errorf("ParseUsizeRange(%s, %d) unexpected error: %v", tt.input, tt.maxValue, err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseUsizeRange(%s, %d) = %v; want %v", tt.input, tt.maxValue, result, tt.expected)
			}
		}
	}
}

func TestUnionUsizeRanges(t *testing.T) {
	tests := []struct {
		input    []string
		maxValue int
		expected []int
		hasError bool
	}{
		{[]string{"1", "3", "5"}, 10, []int{1, 3, 5}, false},
		{[]string{"1-3", "5-7"}, 10, []int{1, 2, 3, 5, 6, 7}, false},
		{[]string{"1-5", "3-7"}, 10, []int{1, 2, 3, 4, 5, 6, 7}, false},
		{[]string{"-3", "8-"}, 10, []int{1, 2, 3, 8, 9, 10}, false},
		{[]string{"invalid"}, 10, nil, true},
	}

	for _, tt := range tests {
		result, err := UnionUsizeRanges(tt.input, tt.maxValue)
		if tt.hasError {
			if err == nil {
				t.Errorf("UnionUsizeRanges(%v, %d) expected error but got none", tt.input, tt.maxValue)
			}
		} else {
			if err != nil {
				t.Errorf("UnionUsizeRanges(%v, %d) unexpected error: %v", tt.input, tt.maxValue, err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("UnionUsizeRanges(%v, %d) = %v; want %v", tt.input, tt.maxValue, result, tt.expected)
			}
		}
	}
}

func TestStrVectorsIntersect(t *testing.T) {
	tests := []struct {
		first    []string
		second   []string
		expected bool
	}{
		{[]string{"a", "b", "c"}, []string{"b", "d"}, true},
		{[]string{"a", "b", "c"}, []string{"d", "e"}, false},
		{[]string{"PDF", "EPUB"}, []string{"pdf", "mobi"}, true},
		{[]string{}, []string{"a"}, false},
		{[]string{"a"}, []string{}, false},
	}

	for _, tt := range tests {
		result := StrVectorsIntersect(tt.first, tt.second)
		if result != tt.expected {
			t.Errorf("StrVectorsIntersect(%v, %v) = %v; want %v", tt.first, tt.second, result, tt.expected)
		}
	}
}
