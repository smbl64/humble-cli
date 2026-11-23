package api

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"
)

func TestParseFlight(t *testing.T) {
	payload := "1:\"hello\"\n2:T3,a\nb\n3:{\"a\":1}\n4:\"$undefined\"\n"

	values := parseFlight(payload)
	if len(values) != 4 {
		t.Fatalf("expected 4 values, got %d: %#v", len(values), values)
	}
	if values[0] != "hello" {
		t.Errorf("row 1 = %#v, want hello", values[0])
	}
	if values[1] != "a\nb" {
		t.Errorf("text row = %#v, want a\\nb", values[1])
	}
	obj, ok := values[2].(map[string]any)
	if !ok || obj["a"] != float64(1) {
		t.Errorf("row 3 = %#v, want map[a:1]", values[2])
	}
	if values[3] != "null" {
		t.Errorf("row 4 = %#v, want the sanitized string null", values[3])
	}
}

func TestParseFlightSkipsMalformedRows(t *testing.T) {
	payload := "1:{\"ok\":true}\n2:not-json\n3:{\"still\":\"parsed\"}\n"

	values := parseFlight(payload)
	if len(values) != 2 {
		t.Fatalf("expected 2 values, got %d: %#v", len(values), values)
	}
}

func TestExtractNextPayload(t *testing.T) {
	html := `<script>self.__next_f.push([1,"1:\"a\"\n2:T3,x\ny"])</script>` +
		`<script>self.__next_f.push([1,"3:{\"b\":2}\n"])</script>`

	payload := extractNextPayload(html)
	want := "1:\"a\"\n2:T3,x\ny3:{\"b\":2}\n"
	if payload != want {
		t.Errorf("payload = %q, want %q", payload, want)
	}
}

func TestSignedURLsFrom(t *testing.T) {
	raw := map[string]any{
		"subproductMachineName": "memoria",
		"signedUrls": map[string]any{
			"memoria": map[string]any{
				"memoria_mac": map[string]any{
					"Memoria.dmg": "https://cdn.example/Memoria.dmg",
				},
			},
		},
	}

	byKey := signedURLsFrom(raw)
	byURL, ok := byKey["memoria_mac"].(map[string]any)
	if !ok {
		t.Fatalf("download key missing: %#v", byKey)
	}
	if byURL["Memoria.dmg"] != "https://cdn.example/Memoria.dmg" {
		t.Errorf("signed url = %#v", byURL["Memoria.dmg"])
	}
}

func TestSignedURLsFromIgnoresNull(t *testing.T) {
	if got := signedURLsFrom(nil); got != nil {
		t.Errorf("nil input = %#v, want nil", got)
	}
	if got := signedURLsFrom("$undefined"); got != nil {
		t.Errorf("string input = %#v, want nil", got)
	}
}

func TestRouterStateTree(t *testing.T) {
	const bundleID = "c7a551c0-093e-4a5a-8da9-65b6119538b9"

	encoded, err := routerStateTree(bundleID)
	if err != nil {
		t.Fatalf("routerStateTree: %v", err)
	}
	decoded, err := url.QueryUnescape(encoded)
	if err != nil {
		t.Fatalf("unescape: %v", err)
	}

	var tree any
	if err := json.Unmarshal([]byte(decoded), &tree); err != nil {
		t.Fatalf("tree is not JSON: %v", err)
	}

	// Root: [["", {children: [...]}]]
	root, ok := tree.([]any)
	if !ok || len(root) != 1 {
		t.Fatalf("root = %#v", tree)
	}
	first, ok := root[0].([]any)
	if !ok || len(first) != 2 || first[0] != "" {
		t.Fatalf("root segment = %#v", root[0])
	}

	if !strings.Contains(decoded, bundleID) {
		t.Errorf("tree does not reference bundle id: %s", decoded)
	}
	if !strings.Contains(decoded, `"__PAGE__"`) {
		t.Errorf("tree has no __PAGE__ segment: %s", decoded)
	}
}
