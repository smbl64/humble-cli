package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// HumanizeBytes formats bytes as human-readable string (KiB, MiB, GiB, TiB)
func HumanizeBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div := uint64(unit)
	exp := 0
	units := []string{"KiB", "MiB", "GiB", "TiB"}

	for n := bytes / unit; n >= unit && exp < len(units)-1; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// ByteStringToNumber parses human-readable byte strings like "500MB", "2GiB"
func ByteStringToNumber(input string) (uint64, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return 0, fmt.Errorf("empty input")
	}

	// Pattern: number followed by optional unit
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([KMGT]i?B?)?$`)
	matches := re.FindStringSubmatch(strings.ToUpper(input))
	if matches == nil {
		return 0, fmt.Errorf("invalid format: %s", input)
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", matches[1])
	}

	unit := matches[2]
	multiplier := uint64(1)

	switch {
	case strings.HasPrefix(unit, "K"):
		multiplier = 1024
	case strings.HasPrefix(unit, "M"):
		multiplier = 1024 * 1024
	case strings.HasPrefix(unit, "G"):
		multiplier = 1024 * 1024 * 1024
	case strings.HasPrefix(unit, "T"):
		multiplier = 1024 * 1024 * 1024 * 1024
	}

	return uint64(value * float64(multiplier)), nil
}
