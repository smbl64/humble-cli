package util

import (
	"net/url"
	"strings"
)

// ReplaceInvalidCharsInFilename sanitizes filenames by replacing invalid characters
func ReplaceInvalidCharsInFilename(input string) string {
	invalidChars := []rune{'/', '\\', '?', '%', '*', ':', '|', '"', '<', '>', ';', '=', '\n'}
	replacement := ' '

	result := make([]rune, 0, len(input))
	for _, c := range input {
		isInvalid := false
		for _, inv := range invalidChars {
			if c == inv {
				isInvalid = true
				break
			}
		}

		if isInvalid {
			result = append(result, replacement)
		} else {
			result = append(result, c)
		}
	}

	return strings.TrimSpace(string(result))
}

// ExtractFilenameFromURL extracts the filename from a URL path
func ExtractFilenameFromURL(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	path := parsedURL.Path
	if path == "" {
		return ""
	}

	parts := strings.Split(path, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}

	return ""
}
