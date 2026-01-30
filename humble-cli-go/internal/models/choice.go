package models

// HumbleChoice represents a Humble Choice subscription period
type HumbleChoice struct {
	Options ContentChoiceOptions `json:"contentChoiceOptions"`
}

// ContentChoiceOptions contains choice subscription details
type ContentChoiceOptions struct {
	Data            ContentChoiceData `json:"contentChoiceData"`
	Gamekey         *string           `json:"gamekey,omitempty"`
	IsActiveContent bool              `json:"isActiveContent"`
	Title           string            `json:"title"`
}

// ContentChoiceData contains game data for the choice period
type ContentChoiceData struct {
	GameData map[string]GameData `json:"game_data"`
}

// GameData represents a game in the choice subscription
type GameData struct {
	Title string `json:"title"`
	Tpkds []Tpkd `json:"tpkds"`
}

// Tpkd represents a third-party key distribution
type Tpkd struct {
	Gamekey        *string `json:"gamekey,omitempty"`
	HumanName      string  `json:"human_name"`
	RedeemedKeyVal *string `json:"redeemed_key_val,omitempty"`
}

// ClaimStatus returns the claim status of this key
func (t *Tpkd) ClaimStatus() string {
	redeemed := t.RedeemedKeyVal != nil
	isActive := t.Gamekey != nil

	if isActive && redeemed {
		return "Yes"
	} else if isActive {
		return "No"
	}
	return "-"
}
