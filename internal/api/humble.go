package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/smbl64/humble-cli/v2/internal/models"
)

const (
	defaultTimeout = 30 * time.Second
	baseURL        = "https://www.humblebundle.com"
	libraryURL     = baseURL + "/home/library"
	userAgent      = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36"
	nextPushMarker = "self.__next_f.push([1,"
)

var (
	// Server action ids are per deployment; the id is looked up in the chunks.
	signedURLsActionPattern = regexp.MustCompile(`createServerReference\)\("([0-9a-f]{40,64})"[^)]{0,200}"getSignedUrlsForEntitlements"`)
	chunkPattern            = regexp.MustCompile(`/_next/static/chunks/[^"\\]+?\.js`)
)

// HumbleApi provides methods to interact with the Humble Bundle site.
//
// Humble rebuilt the store on Next.js in November 2026 and removed the old
// JSON API (/api/v1/user/order, /api/v1/orders, /api/v1/order). The library is
// now served as a React Server Components flight payload, and the bundle pages
// carry the signed CDN download URLs in their inline flight payload. This
// client reads those pages instead of the API.
type HumbleApi struct {
	authKey          string
	client           *http.Client
	signedURLsAction string
}

// New creates a new HumbleApi instance
func New(authKey string) *HumbleApi {
	return &HumbleApi{
		authKey: authKey,
		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// addAuthCookie adds the authentication cookie to the request.
//
// The new site authenticates with an Ory Kratos session cookie whose name is
// account specific (ory_session_<suffix>), so a stored value of the form
// "name=value" is sent verbatim. A bare value keeps working as the legacy
// _simpleauth_sess cookie.
func (api *HumbleApi) addAuthCookie(req *http.Request) {
	if strings.Contains(api.authKey, "=") {
		req.Header.Set("Cookie", api.authKey)
		return
	}
	req.Header.Set("Cookie", "_simpleauth_sess="+api.authKey)
}

func (api *HumbleApi) get(url string, rsc bool) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", NewNetworkError(err)
	}
	req.Header.Set("User-Agent", userAgent)
	if rsc {
		req.Header.Set("RSC", "1")
		req.Header.Set("Accept", "text/x-component")
	} else {
		req.Header.Set("Accept", "text/html")
	}
	api.addAuthCookie(req)

	resp, err := api.client.Do(req)
	if err != nil {
		return "", NewNetworkError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", NewNetworkError(fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", NewNetworkError(err)
	}
	return string(body), nil
}

// ListBundleKeys fetches all bundle keys (library item ids) for the user.
func (api *HumbleApi) ListBundleKeys() ([]string, error) {
	payload, err := api.get(libraryURL, true)
	if err != nil {
		return nil, err
	}

	items, err := libraryItems(payload)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(items))
	for _, item := range items {
		if id, ok := item["id"].(string); ok && id != "" {
			keys = append(keys, id)
		}
	}
	return keys, nil
}

// ListBundles fetches all bundles for the authenticated user.
//
// The library page carries only id, title and type per bundle; sizes and claim
// state require opening the bundle (see ReadBundle).
func (api *HumbleApi) ListBundles() ([]models.Bundle, error) {
	payload, err := api.get(libraryURL, true)
	if err != nil {
		return nil, err
	}

	items, err := libraryItems(payload)
	if err != nil {
		return nil, err
	}

	bundles := make([]models.Bundle, 0, len(items))
	for _, item := range items {
		id, _ := item["id"].(string)
		title, _ := item["title"].(string)
		if id == "" {
			continue
		}
		bundles = append(bundles, models.Bundle{
			Gamekey: id,
			Details: models.BundleDetails{HumanName: title},
		})
	}
	return bundles, nil
}

// libraryItem returns the userLibraryItem of a bundle page's flight payload.
func libraryItem(html string) map[string]any {
	for _, row := range parseFlight(extractNextPayload(html)) {
		if m := findMapWithKey(row, "userLibraryItem"); m != nil {
			if it, ok := m["userLibraryItem"].(map[string]any); ok {
				return it
			}
		}
	}
	return nil
}

// ReadBundle fetches details for a specific bundle, including the signed
// download URLs embedded in the bundle page.
func (api *HumbleApi) ReadBundle(productKey string) (*models.Bundle, error) {
	html, err := api.get(libraryURL+"/"+productKey, false)
	if err != nil {
		return nil, err
	}
	item := libraryItem(html)
	if item == nil {
		return nil, NewBundleNotFoundError()
	}
	return bundleFromItem(productKey, item, api.productSignedURLs), nil
}

// bundleFromItem builds a bundle from a page's userLibraryItem. signedURLs
// resolves the signed download URLs of one product when the page lacks them.
func bundleFromItem(productKey string, item map[string]any, signedURLs func(bundleID, productID string) map[string]any) *models.Bundle {
	bundle := &models.Bundle{Gamekey: productKey, TpkdDict: map[string]any{}}
	if title, ok := item["title"].(string); ok {
		bundle.Details.HumanName = title
	}
	bundle.Created = parseHumbleTime(item["choiceDate"])
	if bundle.Created.IsZero() {
		bundle.Created = parseHumbleTime(item["releaseDate"])
	}

	entitlements, _ := item["entitlements"].(map[string]any)
	if entitlements == nil {
		return bundle
	}

	// Product keys (steam, origin, ...).
	tpks := make([]any, 0)
	for _, raw := range asSlice(entitlements["keys"]) {
		key, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		tpk := map[string]any{}
		if title, ok := key["title"].(string); ok {
			tpk["human_name"] = title
		}
		if rd, ok := key["redemptionData"].(map[string]any); ok {
			if formatted, ok := rd["keyFormatted"].(string); ok && formatted != "" {
				tpk["redeemed_key_val"] = formatted
			}
		}
		tpks = append(tpks, tpk)
	}
	if len(tpks) > 0 {
		bundle.TpkdDict["all_tpks"] = tpks
	}

	// DRM-free downloads.
	for _, raw := range asSlice(entitlements["downloads"]) {
		download, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if typ, _ := download["type"].(string); typ != "download_drm_free" {
			continue
		}

		product := models.Product{}
		product.HumanName, _ = download["title"].(string)
		product.MachineName, _ = download["id"].(string)

		productID, _ := download["id"].(string)
		signedByKey := signedURLsFrom(download["signedUrls"])
		if len(signedByKey) == 0 && signedURLs != nil {
			// The bundle page only carries signed URLs for some bundles; the
			// product page always does.
			signedByKey = signedURLs(productKey, productID)
		}

		productDownload := models.ProductDownload{}
		for _, rawDownload := range asSlice(download["downloads"]) {
			platformDownload, ok := rawDownload.(map[string]any)
			if !ok {
				continue
			}
			downloadKey, _ := platformDownload["key"].(string)
			signedByURL, _ := signedByKey[downloadKey].(map[string]any)

			for _, rawStruct := range asSlice(platformDownload["downloadStruct"]) {
				entry, ok := rawStruct.(map[string]any)
				if !ok {
					continue
				}
				info := models.DownloadInfo{}
				info.Format, _ = entry["name"].(string)
				info.MD5, _ = entry["md5"].(string)
				if size, ok := entry["fileSize"].(float64); ok && size > 0 {
					info.FileSize = uint64(size)
				}
				if urlMap, ok := entry["url"].(map[string]any); ok {
					relative, _ := urlMap["web"].(string)
					info.URL.Web = relative
					if signedByURL != nil {
						if signed, ok := signedByURL[relative].(string); ok && signed != "" {
							info.URL.Web = signed
						}
					}
					if torrent, ok := urlMap["bittorrent"].(string); ok && torrent != "" {
						// bittorrent holds the file name (e.g. foo.pdf.torrent),
						// which is also the key of its signed URL; the unsigned
						// CDN path answers 403.
						info.URL.Bittorrent = "https://cdn.humble.com/humblebundle/" + torrent
						if signedByURL != nil {
							if signed, ok := signedByURL[torrent].(string); ok && signed != "" {
								info.URL.Bittorrent = signed
							}
						}
					}
				}
				productDownload.Items = append(productDownload.Items, info)
			}
		}

		if len(productDownload.Items) == 0 {
			continue
		}
		product.Downloads = []models.ProductDownload{productDownload}
		bundle.Products = append(bundle.Products, product)
	}

	return bundle
}

// Summary returns the product titles and key states of a bundle without
// resolving download URLs. Used by search and the claimed filter.
func (api *HumbleApi) Summary(bundleID string) ([]string, []models.ProductKey, uint64, error) {
	html, err := api.get(libraryURL+"/"+bundleID, false)
	if err != nil {
		return nil, nil, 0, err
	}
	item := libraryItem(html)
	if item == nil {
		return nil, nil, 0, NewBundleNotFoundError()
	}
	products, keys, size := summaryFromItem(item)
	return products, keys, size, nil
}

// summaryFromItem lists the product titles and key states of a bundle's
// userLibraryItem, with the total size of its files.
func summaryFromItem(item map[string]any) ([]string, []models.ProductKey, uint64) {
	entitlements, _ := item["entitlements"].(map[string]any)
	products := make([]string, 0)
	keys := make([]models.ProductKey, 0)
	var totalSize uint64
	if entitlements == nil {
		return products, keys, 0
	}

	// The downloads and the keys of one product can carry the same title, and
	// a bundle can hold several keys for one product.
	seen := make(map[string]bool)
	addProduct := func(title string) {
		if title == "" || seen[title] {
			return
		}
		seen[title] = true
		products = append(products, title)
	}

	for _, raw := range asSlice(entitlements["downloads"]) {
		download, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if typ, _ := download["type"].(string); typ != "download_drm_free" {
			continue
		}
		files := asSlice(download["downloads"])
		// Humble ships a file-less "No Content" entry in almost every bundle.
		if len(files) == 0 {
			continue
		}
		if title, ok := download["title"].(string); ok {
			addProduct(title)
		}
		for _, rawFile := range files {
			platformDownload, ok := rawFile.(map[string]any)
			if !ok {
				continue
			}
			for _, rawStruct := range asSlice(platformDownload["downloadStruct"]) {
				entry, ok := rawStruct.(map[string]any)
				if !ok {
					continue
				}
				if size, ok := entry["fileSize"].(float64); ok && size > 0 {
					totalSize += uint64(size)
				}
			}
		}
	}

	for _, raw := range asSlice(entitlements["keys"]) {
		key, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		pk := models.ProductKey{}
		pk.HumanName, _ = key["title"].(string)
		// Key-only products (Steam keys without downloads) are searchable
		// through their key title.
		addProduct(pk.HumanName)
		if rd, ok := key["redemptionData"].(map[string]any); ok {
			if formatted, ok := rd["keyFormatted"].(string); ok && formatted != "" {
				pk.Redeemed = true
			}
		}
		keys = append(keys, pk)
	}

	return products, keys, totalSize
}

// ListChoices returns the games of a Humble Choice month ("home" for the
// current month). The redesigned membership page carries them in its flight
// payload as subscriptionProducts.
func (api *HumbleApi) ListChoices(period string) ([]string, error) {
	url := baseURL + "/membership/home"
	if period != "" && period != "home" {
		url = baseURL + "/membership/" + period
	}
	html, err := api.get(url, false)
	if err != nil {
		return nil, err
	}
	for _, row := range parseFlight(extractNextPayload(html)) {
		container := findMapWithKey(row, "subscriptionProducts")
		if container == nil {
			continue
		}
		titles := make([]string, 0)
		for _, raw := range asSlice(container["subscriptionProducts"]) {
			product, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			if title, ok := product["title"].(string); ok && title != "" {
				titles = append(titles, title)
			}
		}
		return titles, nil
	}
	return nil, NewBundleNotFoundError()
}

// ReadBundleChoices is not supported any more: the redesigned membership pages
// no longer carry the month's data and the old webpack payload is gone.
func (api *HumbleApi) ReadBundleChoices(when string) (*models.HumbleChoice, error) {
	return nil, NewNetworkError(fmt.Errorf(
		"Humble's new site no longer exposes membership data; open https://www.humblebundle.com/membership in a browser"))
}

// libraryItems extracts the bundleItems array from the library flight payload.
func libraryItems(payload string) ([]map[string]any, error) {
	rows := parseFlight(payload)
	for _, row := range rows {
		container := findMapWithKey(row, "bundleItems")
		if container == nil {
			continue
		}
		items := make([]map[string]any, 0)
		for _, raw := range asSlice(container["bundleItems"]) {
			if item, ok := raw.(map[string]any); ok {
				items = append(items, item)
			}
		}
		return items, nil
	}
	return nil, NewDeserializeError(fmt.Errorf("bundleItems not found in library payload"))
}

// extractNextPayload concatenates the inline RSC flight chunks of an HTML page.
func extractNextPayload(html string) string {
	var builder strings.Builder
	index := 0
	for {
		offset := strings.Index(html[index:], nextPushMarker)
		if offset < 0 {
			break
		}
		start := index + offset + len(nextPushMarker)
		decoder := json.NewDecoder(strings.NewReader(html[start:]))
		var chunk string
		if err := decoder.Decode(&chunk); err != nil {
			index = start + 1
			continue
		}
		builder.WriteString(chunk)
		index = start + int(decoder.InputOffset())
	}
	return builder.String()
}

// parseFlight parses React Server Components flight rows into values.
//
// A row is "<id>:<json>" on a single line, or "<id>:T<hex-byte-length>,<text>"
// for text chunks (which may contain newlines and carry no trailing newline).
func parseFlight(payload string) []any {
	data := []byte(payload)
	values := make([]any, 0, 64)

	for i := 0; i < len(data); {
		if data[i] == '\n' {
			i++
			continue
		}
		colon := bytes.IndexByte(data[i:], ':')
		if colon < 0 {
			break
		}
		colon += i
		start := colon + 1

		if start < len(data) && data[start] == 'T' {
			comma := bytes.IndexByte(data[start:], ',')
			if comma < 0 {
				break
			}
			comma += start
			length, err := strconv.ParseInt(string(data[start+1:comma]), 16, 64)
			if err != nil {
				break
			}
			end := comma + 1 + int(length)
			if end > len(data) {
				end = len(data)
			}
			values = append(values, string(data[comma+1:end]))
			i = end
			continue
		}

		newline := bytes.IndexByte(data[start:], '\n')
		var raw []byte
		if newline < 0 {
			raw = data[start:]
			i = len(data)
		} else {
			raw = data[start : start+newline]
			i = start + newline
		}
		raw = bytes.ReplaceAll(raw, []byte("$undefined"), []byte("null"))
		var value any
		if err := json.Unmarshal(raw, &value); err == nil {
			values = append(values, value)
		}
	}
	return values
}

// findMapWithKey walks a decoded flight value for the first object with key.
func findMapWithKey(value any, key string) map[string]any {
	switch v := value.(type) {
	case map[string]any:
		if _, ok := v[key]; ok {
			return v
		}
		for _, child := range v {
			if found := findMapWithKey(child, key); found != nil {
				return found
			}
		}
	case []any:
		for _, child := range v {
			if found := findMapWithKey(child, key); found != nil {
				return found
			}
		}
	}
	return nil
}

// signedURLsFrom unwraps a signedUrls object into downloadKey -> relativeURL -> URL.
func signedURLsFrom(value any) map[string]any {
	signed, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	subName, _ := signed["subproductMachineName"].(string)
	byProduct, ok := signed["signedUrls"].(map[string]any)
	if !ok {
		return nil
	}
	entry, _ := byProduct[subName].(map[string]any)
	return entry
}

// productSignedURLs resolves signed download URLs for one product.
//
// Bundle pages carry them inline for some bundles only. Otherwise the page's
// "getSignedUrlsForEntitlements" server action is called, exactly as the
// browser's download button does.
func (api *HumbleApi) productSignedURLs(bundleID, productID string) map[string]any {
	if productID == "" {
		return nil
	}

	// 1. Try the product page first (cheap and works for many products).
	if html, err := api.get(libraryURL+"/content/"+productID, false); err == nil {
		for _, row := range parseFlight(extractNextPayload(html)) {
			container := findMapWithKey(row, "signedUrls")
			if container == nil {
				continue
			}
			if entry := signedURLsFrom(container["signedUrls"]); len(entry) > 0 {
				return entry
			}
		}
	}

	// 2. Ask the server action the download button uses.
	actionID, err := api.signedURLsActionID(bundleID)
	if err != nil || actionID == "" {
		return nil
	}
	return api.callSignedURLsAction(bundleID, productID, actionID)
}

// signedURLsActionID discovers the server action id for signed URLs from the
// bundle page's JavaScript chunks. Action ids change with every deployment.
func (api *HumbleApi) signedURLsActionID(bundleID string) (string, error) {
	if api.signedURLsAction != "" {
		return api.signedURLsAction, nil
	}
	html, err := api.get(libraryURL+"/"+bundleID, false)
	if err != nil {
		return "", err
	}
	for _, chunk := range chunkURLs(html) {
		js, err := api.get(baseURL+chunk, false)
		if err != nil {
			continue
		}
		if match := signedURLsActionPattern.FindStringSubmatch(js); match != nil {
			api.signedURLsAction = match[1]
			return api.signedURLsAction, nil
		}
	}
	return "", nil
}

func (api *HumbleApi) callSignedURLsAction(bundleID, productID, actionID string) map[string]any {
	tree, err := routerStateTree(bundleID)
	if err != nil {
		return nil
	}
	body, err := json.Marshal([][]string{{productID}})
	if err != nil {
		return nil
	}
	req, err := http.NewRequest("POST", libraryURL+"/"+bundleID, strings.NewReader(string(body)))
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Next-Action", actionID)
	req.Header.Set("Next-Router-State-Tree", tree)
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	req.Header.Set("Accept", "text/x-component")
	req.Header.Set("Origin", baseURL)
	req.Header.Set("Referer", libraryURL+"/"+bundleID)
	api.addAuthCookie(req)

	resp, err := api.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	for _, row := range parseFlight(string(raw)) {
		container := findMapWithKey(row, "signedUrls")
		if container == nil {
			continue
		}
		if entry := signedURLsFrom(container["signedUrls"]); len(entry) > 0 {
			return entry
		}
	}
	return nil
}

// routerStateTree builds the Next-Router-State-Tree header for a bundle page.
func routerStateTree(bundleID string) (string, error) {
	page := []any{"__PAGE__", map[string]any{}, nil, nil, 0}
	segment := []any{[]any{"id", bundleID, "d", nil}, map[string]any{"children": page}, nil, nil, 0}
	library := []any{"library", map[string]any{"children": []any{segment, nil, nil, 4}}, nil, nil, 8}
	home := []any{"home", map[string]any{"children": library}, nil, nil, 8}
	account := []any{"(account)", map[string]any{"children": home}, nil, nil, 8}
	root := []any{[]any{"", map[string]any{"children": []any{"(default)", map[string]any{"children": account}, nil, nil, 24}}}}

	encoded, err := json.Marshal(root)
	if err != nil {
		return "", err
	}
	return url.QueryEscape(string(encoded)), nil
}

// chunkURLs lists the JavaScript chunk paths referenced by an HTML page.
func chunkURLs(html string) []string {
	seen := map[string]bool{}
	urls := make([]string, 0, 32)
	for _, match := range chunkPattern.FindAllStringSubmatch(html, -1) {
		if !seen[match[0]] {
			seen[match[0]] = true
			urls = append(urls, match[0])
		}
	}
	return urls
}

func asSlice(value any) []any {
	if list, ok := value.([]any); ok {
		return list
	}
	return nil
}

func parseHumbleTime(value any) models.HumbleTime {
	text, ok := value.(string)
	if !ok || text == "" || text == "null" {
		return models.HumbleTime{}
	}
	var parsed models.HumbleTime
	if err := parsed.UnmarshalJSON([]byte(strconv.Quote(text))); err != nil {
		return models.HumbleTime{}
	}
	return parsed
}
