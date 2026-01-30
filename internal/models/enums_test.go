package models

import (
	"strings"
	"testing"
)

func TestClaimStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status ClaimStatus
		want   string
	}{
		{
			name:   "claim status all",
			status: ClaimStatusAll,
			want:   "all",
		},
		{
			name:   "claim status yes",
			status: ClaimStatusYes,
			want:   "yes",
		},
		{
			name:   "claim status no",
			status: ClaimStatusNo,
			want:   "no",
		},
		{
			name:   "invalid claim status",
			status: ClaimStatus(999),
			want:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.String()
			if got != tt.want {
				t.Errorf("ClaimStatus.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseClaimStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ClaimStatus
		wantErr bool
	}{
		{
			name:    "parse all",
			input:   "all",
			want:    ClaimStatusAll,
			wantErr: false,
		},
		{
			name:    "parse ALL uppercase",
			input:   "ALL",
			want:    ClaimStatusAll,
			wantErr: false,
		},
		{
			name:    "parse yes",
			input:   "yes",
			want:    ClaimStatusYes,
			wantErr: false,
		},
		{
			name:    "parse YES uppercase",
			input:   "YES",
			want:    ClaimStatusYes,
			wantErr: false,
		},
		{
			name:    "parse no",
			input:   "no",
			want:    ClaimStatusNo,
			wantErr: false,
		},
		{
			name:    "parse NO uppercase",
			input:   "NO",
			want:    ClaimStatusNo,
			wantErr: false,
		},
		{
			name:    "parse invalid",
			input:   "invalid",
			want:    ClaimStatus(0),
			wantErr: true,
		},
		{
			name:    "parse empty",
			input:   "",
			want:    ClaimStatus(0),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseClaimStatus(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseClaimStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseClaimStatus() = %v, want %v", got, tt.want)
			}

			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), "invalid claim status") {
					t.Errorf("ParseClaimStatus() error = %v, want error containing 'invalid claim status'", err)
				}
			}
		})
	}
}

func TestMatchMode_String(t *testing.T) {
	tests := []struct {
		name string
		mode MatchMode
		want string
	}{
		{
			name: "match mode any",
			mode: MatchModeAny,
			want: "any",
		},
		{
			name: "match mode all",
			mode: MatchModeAll,
			want: "all",
		},
		{
			name: "invalid match mode",
			mode: MatchMode(999),
			want: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mode.String()
			if got != tt.want {
				t.Errorf("MatchMode.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseMatchMode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    MatchMode
		wantErr bool
	}{
		{
			name:    "parse any",
			input:   "any",
			want:    MatchModeAny,
			wantErr: false,
		},
		{
			name:    "parse ANY uppercase",
			input:   "ANY",
			want:    MatchModeAny,
			wantErr: false,
		},
		{
			name:    "parse all",
			input:   "all",
			want:    MatchModeAll,
			wantErr: false,
		},
		{
			name:    "parse ALL uppercase",
			input:   "ALL",
			want:    MatchModeAll,
			wantErr: false,
		},
		{
			name:    "parse invalid",
			input:   "invalid",
			want:    MatchMode(0),
			wantErr: true,
		},
		{
			name:    "parse empty",
			input:   "",
			want:    MatchMode(0),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMatchMode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMatchMode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseMatchMode() = %v, want %v", got, tt.want)
			}

			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), "invalid match mode") {
					t.Errorf("ParseMatchMode() error = %v, want error containing 'invalid match mode'", err)
				}
			}
		})
	}
}

func TestChoicePeriod(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantStr   string
		isCurrent bool
	}{
		{
			name:      "current period",
			input:     "current",
			wantStr:   "current",
			isCurrent: true,
		},
		{
			name:      "current period uppercase",
			input:     "CURRENT",
			wantStr:   "current",
			isCurrent: true,
		},
		{
			name:      "specific period",
			input:     "january-2023",
			wantStr:   "january-2023",
			isCurrent: false,
		},
		{
			name:      "another specific period",
			input:     "December-2022",
			wantStr:   "december-2022",
			isCurrent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			period := NewChoicePeriod(tt.input)

			if period.String() != tt.wantStr {
				t.Errorf("ChoicePeriod.String() = %v, want %v", period.String(), tt.wantStr)
			}

			if period.IsCurrent() != tt.isCurrent {
				t.Errorf("ChoicePeriod.IsCurrent() = %v, want %v", period.IsCurrent(), tt.isCurrent)
			}
		})
	}
}
