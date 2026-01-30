package api

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/smbl64/humble-cli/internal/models"
	"github.com/smbl64/humble-cli/internal/testutil"
)

func TestNew(t *testing.T) {
	authKey := "test-auth-key"
	api := New(authKey)

	if api == nil {
		t.Fatal("New() returned nil")
	}

	if api.authKey != authKey {
		t.Errorf("New() authKey = %v, want %v", api.authKey, authKey)
	}

	if api.client == nil {
		t.Error("New() client is nil")
	}

	if api.client.Timeout != defaultTimeout {
		t.Errorf("New() client.Timeout = %v, want %v", api.client.Timeout, defaultTimeout)
	}
}

func TestAddAuthCookie(t *testing.T) {
	authKey := "test-session-key"
	api := New(authKey)

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	api.addAuthCookie(req)

	cookie := req.Header.Get("Cookie")
	expected := fmt.Sprintf("_simpleauth_sess=%s", authKey)

	if cookie != expected {
		t.Errorf("addAuthCookie() set Cookie = %q, want %q", cookie, expected)
	}
}

// mockTransport is a custom http.RoundTripper for testing
type mockTransport struct {
	handler func(*http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.handler(req)
}

func newMockClient(handler func(*http.Request) (*http.Response, error)) *http.Client {
	return &http.Client{
		Transport: &mockTransport{handler: handler},
	}
}

func jsonResponse(statusCode int, data any) *http.Response {
	var body []byte
	if data != nil {
		var err error
		body, err = testutil.JSONEncode(data)
		if err != nil {
			panic(err)
		}
	}

	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

func stringResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"text/html"}},
	}
}

func TestListBundleKeys(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   any
		wantErr    bool
		wantKeys   []string
	}{
		{
			name:       "successful response",
			statusCode: http.StatusOK,
			response: []models.GameKey{
				{Gamekey: "key1"},
				{Gamekey: "key2"},
				{Gamekey: "key3"},
			},
			wantErr:  false,
			wantKeys: []string{"key1", "key2", "key3"},
		},
		{
			name:       "empty response",
			statusCode: http.StatusOK,
			response:   []models.GameKey{},
			wantErr:    false,
			wantKeys:   []string{},
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			response:   nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := New("test-key")
			api.client = newMockClient(func(req *http.Request) (*http.Response, error) {
				// Verify headers
				if req.Header.Get("Accept") != "application/json" {
					t.Errorf("Request missing Accept header")
				}

				if !strings.Contains(req.Header.Get("Cookie"), "_simpleauth_sess=") {
					t.Errorf("Request missing auth cookie")
				}

				return jsonResponse(tt.statusCode, tt.response), nil
			})

			got, err := api.ListBundleKeys()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListBundleKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if len(got) != len(tt.wantKeys) {
				t.Errorf("ListBundleKeys() returned %d keys, want %d", len(got), len(tt.wantKeys))
				return
			}

			for i, key := range tt.wantKeys {
				if got[i] != key {
					t.Errorf("ListBundleKeys()[%d] = %v, want %v", i, got[i], key)
				}
			}
		})
	}
}

func TestListBundleKeys_InvalidJSON(t *testing.T) {
	api := New("test-key")
	api.client = newMockClient(func(req *http.Request) (*http.Response, error) {
		return stringResponse(http.StatusOK, "invalid json"), nil
	})

	_, err := api.ListBundleKeys()
	if err == nil {
		t.Error("ListBundleKeys() expected error for invalid JSON, got nil")
		return
	}

	var apiErr *ApiError
	if !errors.As(err, &apiErr) || apiErr.Type != DeserializeError {
		t.Errorf("ListBundleKeys() expected DeserializeError, got %v", err)
	}
}

func TestListBundleKeys_NetworkError(t *testing.T) {
	api := New("test-key")
	api.client = newMockClient(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("connection refused")
	})

	_, err := api.ListBundleKeys()
	if err == nil {
		t.Error("ListBundleKeys() expected network error, got nil")
		return
	}

	var apiErr *ApiError
	if !errors.As(err, &apiErr) || apiErr.Type != NetworkError {
		t.Errorf("ListBundleKeys() expected NetworkError, got %v", err)
	}
}

func TestReadBundle(t *testing.T) {
	sampleBundle := testutil.SampleBundle("test-key")

	tests := []struct {
		name       string
		statusCode int
		response   any
		wantErr    bool
	}{
		{
			name:       "successful response",
			statusCode: http.StatusOK,
			response:   sampleBundle,
			wantErr:    false,
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			response:   nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := New("test-key")
			api.client = newMockClient(func(req *http.Request) (*http.Response, error) {
				// Verify URL contains all_tpkds param
				if !strings.Contains(req.URL.RawQuery, "all_tpkds=true") {
					t.Errorf("Request missing all_tpkds parameter")
				}

				// Verify auth cookie
				if !strings.Contains(req.Header.Get("Cookie"), "_simpleauth_sess=") {
					t.Errorf("Request missing auth cookie")
				}

				return jsonResponse(tt.statusCode, tt.response), nil
			})

			got, err := api.ReadBundle("test-key")
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadBundle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if got.Gamekey != sampleBundle.Gamekey {
				t.Errorf("ReadBundle() Gamekey = %v, want %v", got.Gamekey, sampleBundle.Gamekey)
			}
		})
	}
}

func TestReadBundle_InvalidJSON(t *testing.T) {
	api := New("test-key")
	api.client = newMockClient(func(req *http.Request) (*http.Response, error) {
		return stringResponse(http.StatusOK, "invalid json"), nil
	})

	_, err := api.ReadBundle("test-key")
	if err == nil {
		t.Error("ReadBundle() expected error for invalid JSON, got nil")
		return
	}

	var apiErr *ApiError
	if !errors.As(err, &apiErr) || apiErr.Type != DeserializeError {
		t.Errorf("ReadBundle() expected DeserializeError, got %v", err)
	}
}

func TestListBundles(t *testing.T) {
	tests := []struct {
		name         string
		numBundles   int
		wantErr      bool
		checkSorting bool
	}{
		{
			name:         "small number of bundles",
			numBundles:   5,
			wantErr:      false,
			checkSorting: true,
		},
		{
			name:         "triggers chunking",
			numBundles:   25,
			wantErr:      false,
			checkSorting: true,
		},
		{
			name:       "empty response",
			numBundles: 0,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := testutil.SampleGameKeys(tt.numBundles)

			api := New("test-key")
			api.client = newMockClient(func(req *http.Request) (*http.Response, error) {
				// Handle bundle keys request
				if strings.Contains(req.URL.Path, "user/order") {
					gameKeys := make([]models.GameKey, len(keys))
					for i, k := range keys {
						gameKeys[i] = models.GameKey{Gamekey: k}
					}
					return jsonResponse(http.StatusOK, gameKeys), nil
				}

				// Handle bundle data request
				if strings.Contains(req.URL.Path, "/orders") {
					// Parse requested keys from query params
					queryKeys := req.URL.Query()["gamekeys"]

					bundleMap := make(models.BundleMap)
					for _, key := range queryKeys {
						bundle := testutil.SampleBundle(key)
						bundleMap[key] = bundle
					}
					return jsonResponse(http.StatusOK, bundleMap), nil
				}

				return jsonResponse(http.StatusNotFound, nil), nil
			})

			got, err := api.ListBundles()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListBundles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if len(got) != tt.numBundles {
				t.Errorf("ListBundles() returned %d bundles, want %d", len(got), tt.numBundles)
			}

			// Verify sorting (oldest first)
			if tt.checkSorting && len(got) > 1 {
				for i := 1; i < len(got); i++ {
					if got[i].Created.Time.Before(got[i-1].Created.Time) {
						t.Errorf("ListBundles() not sorted correctly at index %d", i)
					}
				}
			}
		})
	}
}

func TestReadBundleChoices(t *testing.T) {
	validHTML := `
		<html>
			<script id="webpack-subscriber-hub-data">
				{"monthlyProductData": {"productName": "January 2023"}}
			</script>
		</html>
	`

	tests := []struct {
		name       string
		statusCode int
		response   string
		wantErr    bool
	}{
		{
			name:       "successful response",
			statusCode: http.StatusOK,
			response:   validHTML,
			wantErr:    false,
		},
		{
			name:       "missing script tag",
			statusCode: http.StatusOK,
			response:   "<html><body>No data</body></html>",
			wantErr:    true,
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			response:   "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := New("test-key")
			api.client = newMockClient(func(req *http.Request) (*http.Response, error) {
				// Verify auth cookie
				if !strings.Contains(req.Header.Get("Cookie"), "_simpleauth_sess=") {
					t.Errorf("Request missing auth cookie")
				}

				return stringResponse(tt.statusCode, tt.response), nil
			})

			got, err := api.ReadBundleChoices("home")
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadBundleChoices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if got == nil {
				t.Error("ReadBundleChoices() returned nil")
			}
		})
	}
}

func TestParseBundleChoices(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		wantErr bool
	}{
		{
			name:    "webpack-subscriber-hub-data",
			html:    `<html><script id="webpack-subscriber-hub-data">{"monthlyProductData": {}}</script></html>`,
			wantErr: false,
		},
		{
			name:    "webpack-monthly-product-data",
			html:    `<html><script id="webpack-monthly-product-data">{"monthlyProductData": {}}</script></html>`,
			wantErr: false,
		},
		{
			name:    "missing script tag",
			html:    `<html><body>No script</body></html>`,
			wantErr: true,
		},
		{
			name:    "invalid JSON in script",
			html:    `<html><script id="webpack-subscriber-hub-data">invalid json</script></html>`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := New("test-key")
			got, err := api.parseBundleChoices(tt.html)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseBundleChoices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Error("parseBundleChoices() returned nil")
			}
		})
	}
}

func TestParseBundleChoices_BundleNotFoundError(t *testing.T) {
	api := New("test-key")
	_, err := api.parseBundleChoices("<html><body>No script</body></html>")

	if err == nil {
		t.Error("parseBundleChoices() expected error, got nil")
		return
	}

	var apiErr *ApiError
	if !errors.As(err, &apiErr) || apiErr.Type != BundleNotFound {
		t.Errorf("parseBundleChoices() expected BundleNotFound error, got %v", err)
	}
}

func TestReadBundlesData(t *testing.T) {
	tests := []struct {
		name       string
		keys       []string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful response",
			keys:       []string{"key1", "key2"},
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "single key",
			keys:       []string{"key1"},
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "server error",
			keys:       []string{"key1"},
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := New("test-key")
			api.client = newMockClient(func(req *http.Request) (*http.Response, error) {
				// Verify all_tpkds parameter
				if req.URL.Query().Get("all_tpkds") != "true" {
					t.Errorf("Missing all_tpkds parameter")
				}

				// Verify gamekeys parameters
				gamekeys := req.URL.Query()["gamekeys"]
				if len(gamekeys) != len(tt.keys) {
					t.Errorf("Got %d gamekeys, want %d", len(gamekeys), len(tt.keys))
				}

				bundleMap := make(models.BundleMap)
				for _, key := range tt.keys {
					bundleMap[key] = testutil.SampleBundle(key)
				}

				return jsonResponse(tt.statusCode, bundleMap), nil
			})

			got, err := api.readBundlesData(tt.keys)
			if (err != nil) != tt.wantErr {
				t.Errorf("readBundlesData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(got) != len(tt.keys) {
				t.Errorf("readBundlesData() returned %d bundles, want %d", len(got), len(tt.keys))
			}
		})
	}
}

func TestListBundles_ErrorPropagation(t *testing.T) {
	// Test that errors from goroutines are properly propagated
	api := New("test-key")
	api.client = newMockClient(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Path, "user/order") {
			// Return many keys to trigger chunking and concurrency
			keys := testutil.SampleGameKeys(25)
			gameKeys := make([]models.GameKey, len(keys))
			for i, k := range keys {
				gameKeys[i] = models.GameKey{Gamekey: k}
			}
			return jsonResponse(http.StatusOK, gameKeys), nil
		}

		// Return error for bundle data request
		return jsonResponse(http.StatusInternalServerError, nil), nil
	})

	_, err := api.ListBundles()
	if err == nil {
		t.Error("ListBundles() expected error from goroutine, got nil")
	}
}
