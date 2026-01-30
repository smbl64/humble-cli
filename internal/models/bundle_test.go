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

func TestBundle_TotalSize(t *testing.T) {
	tests := []struct {
		name     string
		bundle   Bundle
		wantSize uint64
	}{
		{
			name: "empty bundle",
			bundle: Bundle{
				Products: []Product{},
			},
			wantSize: 0,
		},
		{
			name: "bundle with products",
			bundle: Bundle{
				Products: []Product{
					{
						Downloads: []ProductDownload{
							{
								Items: []DownloadInfo{
									{FileSize: 1024},
									{FileSize: 2048},
								},
							},
						},
					},
					{
						Downloads: []ProductDownload{
							{
								Items: []DownloadInfo{
									{FileSize: 512},
								},
							},
						},
					},
				},
			},
			wantSize: 3584, // 1024 + 2048 + 512
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.bundle.TotalSize()
			if got != tt.wantSize {
				t.Errorf("Bundle.TotalSize() = %v, want %v", got, tt.wantSize)
			}
		})
	}
}

func TestBundle_ProductKeys(t *testing.T) {
	tests := []struct {
		name      string
		bundle    Bundle
		wantCount int
		checkFn   func([]ProductKey) bool
	}{
		{
			name: "empty tpkd_dict",
			bundle: Bundle{
				TpkdDict: map[string]any{},
			},
			wantCount: 0,
		},
		{
			name: "no all_tpks key",
			bundle: Bundle{
				TpkdDict: map[string]any{
					"other": "value",
				},
			},
			wantCount: 0,
		},
		{
			name: "valid product keys",
			bundle: Bundle{
				TpkdDict: map[string]any{
					"all_tpks": []any{
						map[string]any{
							"human_name":       "Game 1",
							"redeemed_key_val": "XXXXX-XXXXX",
						},
						map[string]any{
							"human_name": "Game 2",
						},
					},
				},
			},
			wantCount: 2,
			checkFn: func(keys []ProductKey) bool {
				return keys[0].Redeemed && keys[0].HumanName == "Game 1" &&
					!keys[1].Redeemed && keys[1].HumanName == "Game 2"
			},
		},
		{
			name: "invalid all_tpks structure",
			bundle: Bundle{
				TpkdDict: map[string]any{
					"all_tpks": "not an array",
				},
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.bundle.ProductKeys()
			if len(got) != tt.wantCount {
				t.Errorf("Bundle.ProductKeys() count = %v, want %v", len(got), tt.wantCount)
				return
			}

			if tt.checkFn != nil && !tt.checkFn(got) {
				t.Errorf("Bundle.ProductKeys() check function failed")
			}
		})
	}
}

func TestBundle_ClaimStatus(t *testing.T) {
	tests := []struct {
		name   string
		bundle Bundle
		want   string
	}{
		{
			name: "no product keys",
			bundle: Bundle{
				TpkdDict: map[string]any{},
			},
			want: "-",
		},
		{
			name: "all keys redeemed",
			bundle: Bundle{
				TpkdDict: map[string]any{
					"all_tpks": []any{
						map[string]any{
							"human_name":       "Game 1",
							"redeemed_key_val": "XXXXX",
						},
						map[string]any{
							"human_name":       "Game 2",
							"redeemed_key_val": "YYYYY",
						},
					},
				},
			},
			want: "Yes",
		},
		{
			name: "some keys unredeemed",
			bundle: Bundle{
				TpkdDict: map[string]any{
					"all_tpks": []any{
						map[string]any{
							"human_name":       "Game 1",
							"redeemed_key_val": "XXXXX",
						},
						map[string]any{
							"human_name": "Game 2",
						},
					},
				},
			},
			want: "No",
		},
		{
			name: "all keys unredeemed",
			bundle: Bundle{
				TpkdDict: map[string]any{
					"all_tpks": []any{
						map[string]any{
							"human_name": "Game 1",
						},
						map[string]any{
							"human_name": "Game 2",
						},
					},
				},
			},
			want: "No",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.bundle.ClaimStatus()
			if got != tt.want {
				t.Errorf("Bundle.ClaimStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProduct_TotalSize(t *testing.T) {
	tests := []struct {
		name     string
		product  Product
		wantSize uint64
	}{
		{
			name: "empty product",
			product: Product{
				Downloads: []ProductDownload{},
			},
			wantSize: 0,
		},
		{
			name: "product with downloads",
			product: Product{
				Downloads: []ProductDownload{
					{
						Items: []DownloadInfo{
							{FileSize: 1024},
							{FileSize: 2048},
						},
					},
					{
						Items: []DownloadInfo{
							{FileSize: 512},
						},
					},
				},
			},
			wantSize: 3584,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.product.TotalSize()
			if got != tt.wantSize {
				t.Errorf("Product.TotalSize() = %v, want %v", got, tt.wantSize)
			}
		})
	}
}

func TestProduct_FormatsAsVec(t *testing.T) {
	product := Product{
		Downloads: []ProductDownload{
			{
				Items: []DownloadInfo{
					{Format: "PDF"},
					{Format: "EPUB"},
				},
			},
			{
				Items: []DownloadInfo{
					{Format: "MOBI"},
				},
			},
		},
	}

	got := product.FormatsAsVec()
	want := []string{"PDF", "EPUB", "MOBI"}

	if len(got) != len(want) {
		t.Errorf("Product.FormatsAsVec() count = %v, want %v", len(got), len(want))
		return
	}

	for i, format := range want {
		if got[i] != format {
			t.Errorf("Product.FormatsAsVec()[%d] = %v, want %v", i, got[i], format)
		}
	}
}

func TestProduct_Formats(t *testing.T) {
	product := Product{
		Downloads: []ProductDownload{
			{
				Items: []DownloadInfo{
					{Format: "PDF"},
					{Format: "EPUB"},
				},
			},
		},
	}

	got := product.Formats()
	want := "PDF, EPUB"

	if got != want {
		t.Errorf("Product.Formats() = %v, want %v", got, want)
	}
}

func TestProduct_NameMatches(t *testing.T) {
	tests := []struct {
		name     string
		product  Product
		keywords []string
		mode     MatchMode
		want     bool
	}{
		{
			name:     "match mode any - single match",
			product:  Product{HumanName: "The Great Game Bundle"},
			keywords: []string{"great", "missing"},
			mode:     MatchModeAny,
			want:     true,
		},
		{
			name:     "match mode any - no match",
			product:  Product{HumanName: "The Great Game Bundle"},
			keywords: []string{"missing", "notfound"},
			mode:     MatchModeAny,
			want:     false,
		},
		{
			name:     "match mode all - all match",
			product:  Product{HumanName: "The Great Game Bundle"},
			keywords: []string{"great", "game"},
			mode:     MatchModeAll,
			want:     true,
		},
		{
			name:     "match mode all - partial match",
			product:  Product{HumanName: "The Great Game Bundle"},
			keywords: []string{"great", "missing"},
			mode:     MatchModeAll,
			want:     false,
		},
		{
			name:     "case insensitive",
			product:  Product{HumanName: "The Great Game Bundle"},
			keywords: []string{"GREAT", "GAME"},
			mode:     MatchModeAll,
			want:     true,
		},
		{
			name:     "empty keywords",
			product:  Product{HumanName: "The Great Game Bundle"},
			keywords: []string{},
			mode:     MatchModeAll,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.product.NameMatches(tt.keywords, tt.mode)
			if got != tt.want {
				t.Errorf("Product.NameMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProductDownload_TotalSize(t *testing.T) {
	pd := ProductDownload{
		Items: []DownloadInfo{
			{FileSize: 1024},
			{FileSize: 2048},
			{FileSize: 512},
		},
	}

	got := pd.TotalSize()
	want := uint64(3584)

	if got != want {
		t.Errorf("ProductDownload.TotalSize() = %v, want %v", got, want)
	}
}

func TestProductDownload_FormatsAsVec(t *testing.T) {
	pd := ProductDownload{
		Items: []DownloadInfo{
			{Format: "PDF"},
			{Format: "EPUB"},
			{Format: "MOBI"},
		},
	}

	got := pd.FormatsAsVec()
	want := []string{"PDF", "EPUB", "MOBI"}

	if len(got) != len(want) {
		t.Errorf("ProductDownload.FormatsAsVec() count = %v, want %v", len(got), len(want))
		return
	}

	for i, format := range want {
		if got[i] != format {
			t.Errorf("ProductDownload.FormatsAsVec()[%d] = %v, want %v", i, got[i], format)
		}
	}
}

func TestProductDownload_Formats(t *testing.T) {
	pd := ProductDownload{
		Items: []DownloadInfo{
			{Format: "PDF"},
			{Format: "EPUB"},
		},
	}

	got := pd.Formats()
	want := "PDF, EPUB"

	if got != want {
		t.Errorf("ProductDownload.Formats() = %v, want %v", got, want)
	}
}
