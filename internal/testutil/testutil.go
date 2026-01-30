package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// NewTestServer creates a test HTTP server with the given handler
func NewTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// JSONResponse writes a JSON response with the given status code
func JSONResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// CreateTempConfig creates a temporary config file for testing
// Returns the path to the config file
func CreateTempConfig(t *testing.T, sessionKey string) string {
	t.Helper()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".humble-cli-key")

	err := os.WriteFile(configPath, []byte(sessionKey), 0600)
	if err != nil {
		t.Fatalf("failed to create temp config: %v", err)
	}

	return configPath
}

// SetTestHomeDir sets the HOME environment variable for testing
// This allows tests to control where config files are read/written
func SetTestHomeDir(t *testing.T, homeDir string) {
	t.Helper()
	t.Setenv("HOME", homeDir)
	// Windows compatibility
	t.Setenv("USERPROFILE", homeDir)
}

// JSONEncode encodes data to JSON bytes
func JSONEncode(data any) ([]byte, error) {
	return json.Marshal(data)
}
