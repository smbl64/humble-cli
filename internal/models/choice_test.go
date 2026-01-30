package models

import "testing"

func TestTpkd_ClaimStatus(t *testing.T) {
	tests := []struct {
		name string
		tpkd Tpkd
		want string
	}{
		{
			name: "active and redeemed",
			tpkd: Tpkd{
				Gamekey:        stringPtr("gamekey123"),
				RedeemedKeyVal: stringPtr("XXXXX-XXXXX"),
			},
			want: "Yes",
		},
		{
			name: "active but not redeemed",
			tpkd: Tpkd{
				Gamekey:        stringPtr("gamekey123"),
				RedeemedKeyVal: nil,
			},
			want: "No",
		},
		{
			name: "not active",
			tpkd: Tpkd{
				Gamekey:        nil,
				RedeemedKeyVal: nil,
			},
			want: "-",
		},
		{
			name: "not active but has redeemed value",
			tpkd: Tpkd{
				Gamekey:        nil,
				RedeemedKeyVal: stringPtr("XXXXX-XXXXX"),
			},
			want: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tpkd.ClaimStatus()
			if got != tt.want {
				t.Errorf("Tpkd.ClaimStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

// stringPtr is a helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
