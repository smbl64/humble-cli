package testutil

import (
	"fmt"
	"time"

	"github.com/smbl64/humble-cli/internal/models"
)

// SampleBundle creates a fully populated sample bundle for testing
func SampleBundle(gamekey string) models.Bundle {
	claimed := false
	amountSpent := 15.0
	currency := "USD"

	return models.Bundle{
		Gamekey: gamekey,
		Created: models.HumbleTime{Time: time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)},
		Claimed: claimed,
		TpkdDict: map[string]any{
			"all_tpks": []any{
				map[string]any{
					"human_name":       "Test Game Key",
					"redeemed_key_val": "",
				},
				map[string]any{
					"human_name": "Another Game Key",
				},
			},
		},
		Details: models.BundleDetails{
			MachineName: "test_bundle",
			HumanName:   "Test Bundle",
		},
		Products: []models.Product{
			SampleProduct("Test Game 1"),
			SampleProduct("Test Game 2"),
		},
		AmountSpent: &amountSpent,
		Currency:    &currency,
	}
}

// SampleProduct creates a sample product for testing
func SampleProduct(name string) models.Product {
	return models.Product{
		MachineName:       fmt.Sprintf("%s_machine", name),
		HumanName:         name,
		ProductDetailsURL: "https://example.com/product",
		Downloads: []models.ProductDownload{
			{
				Items: []models.DownloadInfo{
					{
						MD5:      "abc123",
						Format:   "PDF",
						FileSize: 1024 * 1024, // 1MB
						URL: models.DownloadURL{
							Web:        "https://example.com/download.pdf",
							Bittorrent: "",
						},
					},
					{
						MD5:      "def456",
						Format:   "EPUB",
						FileSize: 512 * 1024, // 512KB
						URL: models.DownloadURL{
							Web:        "https://example.com/download.epub",
							Bittorrent: "",
						},
					},
				},
			},
		},
	}
}

// SampleGameKeys generates N sample game keys for testing
func SampleGameKeys(count int) []string {
	keys := make([]string, count)
	for i := range count {
		keys[i] = fmt.Sprintf("gamekey_%d", i+1)
	}
	return keys
}
