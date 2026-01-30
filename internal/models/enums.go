package models

import (
	"fmt"
	"strings"
)

// ClaimStatus represents whether a bundle has been claimed
type ClaimStatus int

const (
	ClaimStatusAll ClaimStatus = iota
	ClaimStatusYes
	ClaimStatusNo
)

func (c ClaimStatus) String() string {
	switch c {
	case ClaimStatusAll:
		return "all"
	case ClaimStatusYes:
		return "yes"
	case ClaimStatusNo:
		return "no"
	default:
		return "unknown"
	}
}

func ParseClaimStatus(s string) (ClaimStatus, error) {
	switch strings.ToLower(s) {
	case "all":
		return ClaimStatusAll, nil
	case "yes":
		return ClaimStatusYes, nil
	case "no":
		return ClaimStatusNo, nil
	default:
		return 0, fmt.Errorf("invalid claim status: %s (valid: all, yes, no)", s)
	}
}

// ChoicePeriod represents a Humble Choice period
type ChoicePeriod struct {
	value string
}

func NewChoicePeriod(s string) ChoicePeriod {
	return ChoicePeriod{value: strings.ToLower(s)}
}

func (c ChoicePeriod) String() string {
	return c.value
}

func (c ChoicePeriod) IsCurrent() bool {
	return c.value == "current"
}

// MatchMode represents how keywords should match products
type MatchMode int

const (
	MatchModeAny MatchMode = iota
	MatchModeAll
)

func (m MatchMode) String() string {
	switch m {
	case MatchModeAny:
		return "any"
	case MatchModeAll:
		return "all"
	default:
		return "unknown"
	}
}

func ParseMatchMode(s string) (MatchMode, error) {
	switch strings.ToLower(s) {
	case "any":
		return MatchModeAny, nil
	case "all":
		return MatchModeAll, nil
	default:
		return 0, fmt.Errorf("invalid match mode: %s (valid: any, all)", s)
	}
}
