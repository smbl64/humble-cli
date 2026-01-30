package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/smbl64/humble-cli/internal/testutil"
)

func TestGetConfigFileName(t *testing.T) {
	tests := []struct {
		name      string
		setupHome bool
		wantErr   bool
	}{
		{
			name:      "valid home directory",
			setupHome: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupHome {
				tempDir := t.TempDir()
				testutil.SetTestHomeDir(t, tempDir)
			}

			got, err := GetConfigFileName()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfigFileName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !strings.HasSuffix(got, configFileName) {
					t.Errorf("GetConfigFileName() = %v, want path ending with %v", got, configFileName)
				}
			}
		})
	}
}

func TestGetConfig(t *testing.T) {
	tests := []struct {
		name         string
		setupConfig  bool
		sessionKey   string
		wantErr      bool
		errContains  string
		wantKey      string
	}{
		{
			name:        "valid config file",
			setupConfig: true,
			sessionKey:  "test-session-key-123",
			wantErr:     false,
			wantKey:     "test-session-key-123",
		},
		{
			name:        "config file with whitespace",
			setupConfig: true,
			sessionKey:  "  test-key-with-spaces  \n",
			wantErr:     false,
			wantKey:     "test-key-with-spaces",
		},
		{
			name:        "empty config file",
			setupConfig: true,
			sessionKey:  "   \n\t  ",
			wantErr:     true,
			errContains: "empty",
		},
		{
			name:        "missing config file",
			setupConfig: false,
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			testutil.SetTestHomeDir(t, tempDir)

			if tt.setupConfig {
				configPath := filepath.Join(tempDir, configFileName)
				err := os.WriteFile(configPath, []byte(tt.sessionKey), 0600)
				if err != nil {
					t.Fatalf("failed to setup config: %v", err)
				}
			}

			got, err := GetConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetConfig() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if got != tt.wantKey {
				t.Errorf("GetConfig() = %v, want %v", got, tt.wantKey)
			}
		})
	}
}

func TestSetConfig(t *testing.T) {
	tests := []struct {
		name        string
		sessionKey  string
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid session key",
			sessionKey: "valid-session-key",
			wantErr:    false,
		},
		{
			name:       "session key with whitespace",
			sessionKey: "  key-with-spaces  \n",
			wantErr:    false,
		},
		{
			name:        "empty session key",
			sessionKey:  "",
			wantErr:     true,
			errContains: "empty",
		},
		{
			name:        "whitespace only session key",
			sessionKey:  "   \n\t  ",
			wantErr:     true,
			errContains: "empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			testutil.SetTestHomeDir(t, tempDir)

			err := SetConfig(tt.sessionKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("SetConfig() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			// Verify file was created
			configPath := filepath.Join(tempDir, configFileName)
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				t.Errorf("SetConfig() did not create config file")
				return
			}

			// Verify content
			content, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("failed to read config file: %v", err)
			}
			expectedContent := strings.TrimSpace(tt.sessionKey)
			if string(content) != expectedContent {
				t.Errorf("SetConfig() wrote %q, want %q", string(content), expectedContent)
			}
		})
	}
}

func TestConfigRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	testutil.SetTestHomeDir(t, tempDir)

	testKey := "test-session-key-roundtrip"

	// Set config
	err := SetConfig(testKey)
	if err != nil {
		t.Fatalf("SetConfig() failed: %v", err)
	}

	// Get config
	got, err := GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() failed: %v", err)
	}

	if got != testKey {
		t.Errorf("Config roundtrip: got %q, want %q", got, testKey)
	}
}
