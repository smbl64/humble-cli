package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// HumbleTime is a custom time type that handles Humble Bundle's datetime format
type HumbleTime struct {
	time.Time
}

// UnmarshalJSON implements custom JSON unmarshaling for HumbleTime
func (ht *HumbleTime) UnmarshalJSON(data []byte) error {
	// Remove quotes
	s := strings.Trim(string(data), `"`)
	if s == "null" || s == "" {
		return nil
	}

	// Humble Bundle uses format: "2021-04-05T20:01:30.481166"
	// This is like RFC3339 but without timezone
	formats := []string{
		"2006-01-02T15:04:05.999999",
		"2006-01-02T15:04:05",
		time.RFC3339,
	}

	var err error
	for _, format := range formats {
		ht.Time, err = time.Parse(format, s)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("cannot parse time %q: %w", s, err)
}

// BundleMap is a map of gamekey to Bundle
type BundleMap map[string]Bundle

// GameKey represents a simple gamekey response
type GameKey struct {
	Gamekey string `json:"gamekey"`
}

// Bundle represents a purchased Humble Bundle
type Bundle struct {
	Gamekey     string                 `json:"gamekey"`
	Created     HumbleTime             `json:"created"`
	Claimed     bool                   `json:"claimed"`
	TpkdDict    map[string]any         `json:"tpkd_dict"`
	Details     BundleDetails          `json:"product"`
	Products    []Product              `json:"-"` // Custom unmarshal
	AmountSpent *float64               `json:"amount_spent,omitempty"`
	Currency    *string                `json:"currency,omitempty"`
}

// UnmarshalJSON implements custom JSON unmarshaling to skip malformed products
func (b *Bundle) UnmarshalJSON(data []byte) error {
	type Alias Bundle
	aux := &struct {
		RawProducts []json.RawMessage `json:"subproducts"`
		*Alias
	}{
		Alias: (*Alias)(b),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Parse products individually, skip failures (VecSkipError equivalent)
	b.Products = make([]Product, 0, len(aux.RawProducts))
	for _, raw := range aux.RawProducts {
		var p Product
		if err := json.Unmarshal(raw, &p); err == nil {
			b.Products = append(b.Products, p)
		}
		// Silently skip malformed products
	}

	return nil
}

// TotalSize returns the sum of all product sizes
func (b *Bundle) TotalSize() uint64 {
	var total uint64
	for _, p := range b.Products {
		total += p.TotalSize()
	}
	return total
}

// ProductKey represents a redeemable product key
type ProductKey struct {
	Redeemed  bool
	HumanName string
}

// ProductKeys extracts all product keys from tpkd_dict
func (b *Bundle) ProductKeys() []ProductKey {
	allTpks, ok := b.TpkdDict["all_tpks"]
	if !ok {
		return []ProductKey{}
	}

	tpksArray, ok := allTpks.([]any)
	if !ok {
		return []ProductKey{}
	}

	result := make([]ProductKey, 0, len(tpksArray))
	for _, item := range tpksArray {
		tpk, ok := item.(map[string]any)
		if !ok {
			continue
		}

		// Check if redeemed_key_val exists and is a string
		redeemed := false
		if redeemedVal, exists := tpk["redeemed_key_val"]; exists {
			_, redeemed = redeemedVal.(string)
		}

		humanName := ""
		if hn, exists := tpk["human_name"]; exists {
			if hnStr, ok := hn.(string); ok {
				humanName = hnStr
			}
		}

		result = append(result, ProductKey{
			Redeemed:  redeemed,
			HumanName: humanName,
		})
	}

	return result
}

// ClaimStatus returns the claim status of the bundle
func (b *Bundle) ClaimStatus() string {
	productKeys := b.ProductKeys()
	totalCount := len(productKeys)

	if totalCount == 0 {
		return "-"
	}

	unusedCount := 0
	for _, pk := range productKeys {
		if !pk.Redeemed {
			unusedCount++
		}
	}

	if unusedCount > 0 {
		return "No"
	}
	return "Yes"
}

// BundleDetails contains bundle metadata
type BundleDetails struct {
	MachineName string `json:"machine_name"`
	HumanName   string `json:"human_name"`
}

// Product represents a product within a bundle
type Product struct {
	MachineName        string            `json:"machine_name"`
	HumanName          string            `json:"human_name"`
	ProductDetailsURL  string            `json:"url"`
	Downloads          []ProductDownload `json:"downloads"`
}

// TotalSize returns the sum of all download sizes for this product
func (p *Product) TotalSize() uint64 {
	var total uint64
	for _, d := range p.Downloads {
		total += d.TotalSize()
	}
	return total
}

// FormatsAsVec returns all formats as a slice
func (p *Product) FormatsAsVec() []string {
	formats := []string{}
	for _, d := range p.Downloads {
		formats = append(formats, d.FormatsAsVec()...)
	}
	return formats
}

// Formats returns all formats as a comma-separated string
func (p *Product) Formats() string {
	return strings.Join(p.FormatsAsVec(), ", ")
}

// NameMatches checks if the product name matches the given keywords
func (p *Product) NameMatches(keywords []string, mode MatchMode) bool {
	humanName := strings.ToLower(p.HumanName)
	words := strings.Fields(humanName)
	wordSet := make(map[string]bool)
	for _, w := range words {
		wordSet[w] = true
	}

	kwMatched := 0
	for _, kw := range keywords {
		kw = strings.ToLower(kw)
		if !wordSet[kw] {
			continue
		}

		switch mode {
		case MatchModeAny:
			return true
		case MatchModeAll:
			kwMatched++
			if kwMatched == len(keywords) {
				return true
			}
		}
	}

	return false
}

// ProductDownload represents a downloadable item
type ProductDownload struct {
	Items []DownloadInfo `json:"download_struct"`
}

// TotalSize returns the sum of all item file sizes
func (pd *ProductDownload) TotalSize() uint64 {
	var total uint64
	for _, item := range pd.Items {
		total += item.FileSize
	}
	return total
}

// FormatsAsVec returns all formats as a slice
func (pd *ProductDownload) FormatsAsVec() []string {
	formats := make([]string, 0, len(pd.Items))
	for _, item := range pd.Items {
		formats = append(formats, item.Format)
	}
	return formats
}

// Formats returns all formats as a comma-separated string
func (pd *ProductDownload) Formats() string {
	return strings.Join(pd.FormatsAsVec(), ", ")
}

// DownloadInfo contains download metadata
type DownloadInfo struct {
	MD5      string      `json:"md5"`
	Format   string      `json:"name"`
	FileSize uint64      `json:"file_size"`
	URL      DownloadURL `json:"url"`
}

// DownloadURL contains web and torrent URLs
type DownloadURL struct {
	Web        string `json:"web"`
	Bittorrent string `json:"bittorrent"`
}
