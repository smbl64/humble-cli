package util

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ParseUsizeRange parses a range string and returns all values in that range.
// Supported formats:
// - Single value: "42" -> [42]
// - Range with beginning and end: "1-5" -> [1,2,3,4,5]
// - Range with no end: "10-" -> [10...maxValue]
// - Range with no beginning: "-5" -> [1,2,3,4,5]
// Note: ranges start at 1, not 0
func ParseUsizeRange(value string, maxValue int) ([]int, error) {
	dashIdx := strings.Index(value, "-")

	// Single value
	if dashIdx == -1 {
		v, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", value)
		}
		return []int{v}, nil
	}

	// Range: split on dash
	left := value[:dashIdx]
	right := value[dashIdx+1:]

	rangeLeft := 1
	if left != "" {
		v, err := strconv.Atoi(left)
		if err != nil {
			return nil, fmt.Errorf("invalid left bound: %s", left)
		}
		rangeLeft = v
	}

	rangeRight := maxValue
	if right != "" {
		v, err := strconv.Atoi(right)
		if err != nil {
			return nil, fmt.Errorf("invalid right bound: %s", right)
		}
		rangeRight = v
	}

	// Generate range (inclusive on both ends)
	result := make([]int, 0, rangeRight-rangeLeft+1)
	for i := rangeLeft; i <= rangeRight; i++ {
		result = append(result, i)
	}

	return result, nil
}

// UnionUsizeRanges parses multiple range strings and returns the union of all values
func UnionUsizeRanges(values []string, maxValue int) ([]int, error) {
	invalidValues := []string{}
	parsed := make(map[int]bool)

	for _, v := range values {
		usizeValues, err := ParseUsizeRange(v, maxValue)
		if err != nil {
			invalidValues = append(invalidValues, v)
			continue
		}

		for _, num := range usizeValues {
			parsed[num] = true
		}
	}

	if len(invalidValues) > 0 {
		quoted := make([]string, len(invalidValues))
		for i, v := range invalidValues {
			quoted[i] = fmt.Sprintf("'%s'", v)
		}
		return nil, fmt.Errorf("invalid values: %s", strings.Join(quoted, ", "))
	}

	// Convert to sorted slice
	output := make([]int, 0, len(parsed))
	for num := range parsed {
		output = append(output, num)
	}
	sort.Ints(output)

	return output, nil
}

// StrVectorsIntersect checks if two string slices have any common element (case-insensitive)
func StrVectorsIntersect(first, second []string) bool {
	if len(first) == 0 || len(second) == 0 {
		return false
	}

	firstSet := make(map[string]bool)
	for _, item := range first {
		firstSet[strings.ToLower(item)] = true
	}

	for _, item := range second {
		if firstSet[strings.ToLower(item)] {
			return true
		}
	}

	return false
}
