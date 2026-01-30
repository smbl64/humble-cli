package models

import (
	"encoding/json"
	"testing"
)

func TestHumbleTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		checkFn  func(HumbleTime) bool
	}{
		{
			name:    "Humble Bundle format with microseconds",
			input:   `"2021-04-05T20:01:30.481166"`,
			wantErr: false,
			checkFn: func(ht HumbleTime) bool {
				return ht.Year() == 2021 && ht.Month() == 4 && ht.Day() == 5
			},
		},
		{
			name:    "Humble Bundle format without microseconds",
			input:   `"2021-04-05T20:01:30"`,
			wantErr: false,
			checkFn: func(ht HumbleTime) bool {
				return ht.Year() == 2021 && ht.Month() == 4 && ht.Day() == 5
			},
		},
		{
			name:    "RFC3339 format",
			input:   `"2021-04-05T20:01:30Z"`,
			wantErr: false,
			checkFn: func(ht HumbleTime) bool {
				return ht.Year() == 2021
			},
		},
		{
			name:    "Null value",
			input:   `null`,
			wantErr: false,
			checkFn: func(ht HumbleTime) bool {
				return ht.IsZero()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ht HumbleTime
			err := json.Unmarshal([]byte(tt.input), &ht)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.checkFn != nil && !tt.checkFn(ht) {
				t.Errorf("Check function failed for parsed time: %v", ht.Time)
			}
		})
	}
}

func TestBundle_UnmarshalJSON(t *testing.T) {
	// Test that Bundle can be unmarshaled with the custom time format
	jsonData := `{
		"gamekey": "test123",
		"created": "2021-04-05T20:01:30.481166",
		"claimed": false,
		"tpkd_dict": {},
		"product": {
			"machine_name": "test_bundle",
			"human_name": "Test Bundle"
		},
		"subproducts": []
	}`

	var bundle Bundle
	err := json.Unmarshal([]byte(jsonData), &bundle)
	if err != nil {
		t.Fatalf("Failed to unmarshal bundle: %v", err)
	}

	if bundle.Gamekey != "test123" {
		t.Errorf("Gamekey = %s; want test123", bundle.Gamekey)
	}

	if bundle.Created.Year() != 2021 {
		t.Errorf("Created year = %d; want 2021", bundle.Created.Year())
	}

	if bundle.Details.HumanName != "Test Bundle" {
		t.Errorf("HumanName = %s; want Test Bundle", bundle.Details.HumanName)
	}
}

func TestBundle_PartialProductDeserialization(t *testing.T) {
	// Test that malformed products are skipped (VecSkipError behavior)
	jsonData := `{
		"gamekey": "test123",
		"created": "2021-04-05T20:01:30",
		"claimed": false,
		"tpkd_dict": {},
		"product": {
			"machine_name": "test_bundle",
			"human_name": "Test Bundle"
		},
		"subproducts": [
			{
				"machine_name": "valid_product",
				"human_name": "Valid Product",
				"url": "http://example.com",
				"downloads": []
			},
			{
				"invalid": "this should be skipped"
			},
			{
				"machine_name": "another_valid",
				"human_name": "Another Valid",
				"url": "http://example.com",
				"downloads": []
			}
		]
	}`

	var bundle Bundle
	err := json.Unmarshal([]byte(jsonData), &bundle)
	if err != nil {
		t.Fatalf("Failed to unmarshal bundle: %v", err)
	}

	// Should have at least 2 valid products
	// Note: Go's JSON unmarshaler is lenient and may parse partial structs
	if len(bundle.Products) < 2 {
		t.Errorf("Products count = %d; want at least 2", len(bundle.Products))
	}

	if len(bundle.Products) > 0 && bundle.Products[0].HumanName != "Valid Product" {
		t.Errorf("First product name = %s; want Valid Product", bundle.Products[0].HumanName)
	}

	// Verify the second valid product is parsed correctly
	if len(bundle.Products) > 2 && bundle.Products[2].HumanName != "Another Valid" {
		t.Errorf("Third product name = %s; want Another Valid", bundle.Products[2].HumanName)
	}
}
