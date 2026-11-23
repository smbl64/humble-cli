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

func TestBundleFromItemUsesSignedTorrentURL(t *testing.T) {
	item := map[string]any{
		"title": "DevOps Bundle",
		"entitlements": map[string]any{
			"downloads": []any{
				map[string]any{
					"type":  "download_drm_free",
					"id":    "devops",
					"title": "Go in Action",
					"signedUrls": map[string]any{
						"subproductMachineName": "devops",
						"signedUrls": map[string]any{
							"devops": map[string]any{
								"ebook": map[string]any{
									"goinaction.pdf":         "https://cdn.example/goinaction.pdf?sig=1",
									"goinaction.pdf.torrent": "https://cdn.example/goinaction.pdf.torrent?sig=1",
								},
							},
						},
					},
					"downloads": []any{
						map[string]any{
							"key": "ebook",
							"downloadStruct": []any{
								map[string]any{
									"name":     "PDF",
									"md5":      "abc",
									"fileSize": float64(1024),
									"url": map[string]any{
										"web":        "goinaction.pdf",
										"bittorrent": "goinaction.pdf.torrent",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	bundle := bundleFromItem("devops-bundle", item, nil)
	if len(bundle.Products) != 1 || len(bundle.Products[0].Downloads) != 1 {
		t.Fatalf("products = %#v", bundle.Products)
	}
	info := bundle.Products[0].Downloads[0].Items[0]
	if info.URL.Web != "https://cdn.example/goinaction.pdf?sig=1" {
		t.Errorf("web url = %q", info.URL.Web)
	}
	if info.URL.Bittorrent != "https://cdn.example/goinaction.pdf.torrent?sig=1" {
		t.Errorf("torrent url = %q, want the signed one", info.URL.Bittorrent)
	}
}

func TestBundleFromItemFallsBackToUnsignedTorrentURL(t *testing.T) {
	item := map[string]any{
		"entitlements": map[string]any{
			"downloads": []any{
				map[string]any{
					"type": "download_drm_free",
					"id":   "devops",
					"downloads": []any{
						map[string]any{
							"key": "ebook",
							"downloadStruct": []any{
								map[string]any{
									"name": "PDF",
									"url": map[string]any{
										"web":        "goinaction.pdf",
										"bittorrent": "goinaction.pdf.torrent",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	bundle := bundleFromItem("devops-bundle", item, nil)
	info := bundle.Products[0].Downloads[0].Items[0]
	if info.URL.Bittorrent != "https://cdn.humble.com/humblebundle/goinaction.pdf.torrent" {
		t.Errorf("torrent url = %q", info.URL.Bittorrent)
	}
}

func TestSummaryFromItemSkipsEmptyDownloadsAndAddsKeyTitles(t *testing.T) {
	item := map[string]any{
		"entitlements": map[string]any{
			"downloads": []any{
				map[string]any{"type": "download_drm_free", "title": "No Content", "id": "empty"},
				map[string]any{
					"type":  "download_drm_free",
					"id":    "civ",
					"title": "Civilization VI",
					"downloads": []any{
						map[string]any{
							"key": "ebook",
							"downloadStruct": []any{
								map[string]any{"fileSize": float64(100)},
							},
						},
					},
				},
			},
			"keys": []any{
				map[string]any{"title": "Civilization VI"},
				map[string]any{"title": "Steam Only Game"},
			},
		},
	}

	products, keys, size := summaryFromItem(item)
	want := []string{"Civilization VI", "Steam Only Game"}
	if len(products) != len(want) {
		t.Fatalf("products = %#v, want %#v", products, want)
	}
	for i := range want {
		if products[i] != want[i] {
			t.Errorf("products[%d] = %q, want %q", i, products[i], want[i])
		}
	}
	if len(keys) != 2 || keys[1].HumanName != "Steam Only Game" {
		t.Errorf("keys = %#v", keys)
	}
	if size != 100 {
		t.Errorf("size = %d, want 100", size)
	}
}
