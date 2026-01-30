package keymatch

import (
	"fmt"
	"strings"
)

// FullKeySize is the length of a complete Humble Bundle key
const FullKeySize = 16

// FindKey searches for a bundle key matching the input (case-insensitive prefix match)
// Returns the matching key or an error if no match or multiple matches found
func FindKey(keys []string, input string) (string, error) {
	// If input is already a full key, return it
	if len(input) == FullKeySize {
		return input, nil
	}

	// Case-insensitive prefix matching
	lowercaseInput := strings.ToLower(input)
	matches := []string{}

	for _, key := range keys {
		if strings.HasPrefix(strings.ToLower(key), lowercaseInput) {
			matches = append(matches, key)
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no bundle key found matching '%s'", input)
	}

	if len(matches) > 1 {
		return "", fmt.Errorf("multiple bundle keys match '%s': %v", input, matches)
	}

	return matches[0], nil
}

// GetMatches returns all keys matching the input (used for testing/debugging)
func GetMatches(keys []string, input string) []string {
	if len(input) == FullKeySize {
		return []string{input}
	}

	lowercaseInput := strings.ToLower(input)
	matches := []string{}

	for _, key := range keys {
		if strings.HasPrefix(strings.ToLower(key), lowercaseInput) {
			matches = append(matches, key)
		}
	}

	return matches
}
