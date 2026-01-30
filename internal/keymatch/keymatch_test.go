package keymatch

import (
	"testing"
)

func TestFindKey_ExactMatch(t *testing.T) {
	keys := []string{"1234567890123456", "abcdefghijklmnop"}
	result, err := FindKey(keys, "1234567890123456")

	if err != nil {
		t.Errorf("FindKey() unexpected error: %v", err)
	}
	if result != "1234567890123456" {
		t.Errorf("FindKey() = %s; want %s", result, "1234567890123456")
	}
}

func TestFindKey_SingleMatch(t *testing.T) {
	keys := []string{"1aAaBbCcDdEeFfGg", "2bBbCcDdEeFfGgHh"}
	result, err := FindKey(keys, "1a")

	if err != nil {
		t.Errorf("FindKey() unexpected error: %v", err)
	}
	if result != "1aAaBbCcDdEeFfGg" {
		t.Errorf("FindKey() = %s; want %s", result, "1aAaBbCcDdEeFfGg")
	}
}

func TestFindKey_CaseInsensitive(t *testing.T) {
	keys := []string{"AaBbCcDdEeFfGgHh", "BbCcDdEeFfGgHhIi"}
	result, err := FindKey(keys, "aa")

	if err != nil {
		t.Errorf("FindKey() unexpected error: %v", err)
	}
	if result != "AaBbCcDdEeFfGgHh" {
		t.Errorf("FindKey() = %s; want %s", result, "AaBbCcDdEeFfGgHh")
	}
}

func TestFindKey_NoMatch(t *testing.T) {
	keys := []string{"1aAaBbCcDdEeFfGg", "2bBbCcDdEeFfGgHh"}
	_, err := FindKey(keys, "3c")

	if err == nil {
		t.Error("FindKey() expected error for no match but got none")
	}
}

func TestFindKey_MultipleMatches(t *testing.T) {
	keys := []string{"1aAaBbCcDdEeFfGg", "1aXxYyZzAaBbCcDd"}
	_, err := FindKey(keys, "1a")

	if err == nil {
		t.Error("FindKey() expected error for multiple matches but got none")
	}
}

func TestGetMatches(t *testing.T) {
	keys := []string{"1aAaBbCcDdEeFfGg", "1aXxYyZzAaBbCcDd", "2bBbCcDdEeFfGgHh"}
	result := GetMatches(keys, "1a")

	if len(result) != 2 {
		t.Errorf("GetMatches() returned %d matches; want 2", len(result))
	}
}
