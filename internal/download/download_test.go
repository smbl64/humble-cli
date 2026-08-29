package download

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetContentLengthHeadHostileServer verifies that GetContentLength falls
// back to GET when the server rejects HEAD with a non-200 status (e.g. 405)
// instead of erroring out. Regression test: previously only a network-level
// error from client.Head triggered the GET fallback, so a 405 HEAD response
// caused an immediate failure even though GET would succeed.
func TestGetContentLengthHeadHostileServer(t *testing.T) {
	const body = "hello world" // 11 bytes

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodHead:
			// Server that does not support HEAD.
			w.WriteHeader(http.StatusMethodNotAllowed)
		case http.MethodGet:
			w.Header().Set("Content-Length", "11")
			_, _ = w.Write([]byte(body))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	length, err := GetContentLength(srv.URL)
	if err != nil {
		t.Fatalf("expected GET fallback to succeed, got error: %v", err)
	}
	if length != uint64(len(body)) {
		t.Fatalf("expected length %d, got %d", len(body), length)
	}
}

// TestGetContentLengthHeadSupported verifies the normal path: a server that
// supports HEAD returns the length directly without a GET fallback.
func TestGetContentLengthHeadSupported(t *testing.T) {
	const body = "0123456789" // 10 bytes

	var gotHead bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			gotHead = true
			w.Header().Set("Content-Length", "10")
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	length, err := GetContentLength(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if length != uint64(len(body)) {
		t.Fatalf("expected length %d, got %d", len(body), length)
	}
	if !gotHead {
		t.Fatalf("expected HEAD to be used")
	}
}
