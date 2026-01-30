package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/smbl64/humble-cli/internal/models"
)

const (
	chunkSize           = 10
	defaultTimeout      = 30 * time.Second
	baseURL             = "https://www.humblebundle.com"
	apiUserOrderURL     = baseURL + "/api/v1/user/order"
	apiBundlesURL       = baseURL + "/api/v1/orders"
	apiBundleDetailsURL = baseURL + "/api/v1/order"
	membershipURL       = baseURL + "/membership"
)

// HumbleApi provides methods to interact with the Humble Bundle API
type HumbleApi struct {
	authKey string
	client  *http.Client
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

// addAuthCookie adds the authentication cookie to the request
func (api *HumbleApi) addAuthCookie(req *http.Request) {
	req.Header.Set("Cookie", fmt.Sprintf("_simpleauth_sess=%s", api.authKey))
}

// ListBundleKeys fetches all bundle keys for the authenticated user
func (api *HumbleApi) ListBundleKeys() ([]string, error) {
	req, err := http.NewRequest("GET", apiUserOrderURL, nil)
	if err != nil {
		return nil, NewNetworkError(err)
	}

	req.Header.Set("Accept", "application/json")
	api.addAuthCookie(req)

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, NewNetworkError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewNetworkError(fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewNetworkError(err)
	}

	var gameKeys []models.GameKey
	if err := json.Unmarshal(body, &gameKeys); err != nil {
		return nil, NewDeserializeError(err)
	}

	keys := make([]string, 0, len(gameKeys))
	for _, gk := range gameKeys {
		keys = append(keys, gk.Gamekey)
	}

	return keys, nil
}

// ListBundles fetches all bundles for the authenticated user
func (api *HumbleApi) ListBundles() ([]models.Bundle, error) {
	gameKeys, err := api.ListBundleKeys()
	if err != nil {
		return nil, err
	}

	// Split keys into chunks
	chunks := make([][]string, 0)
	for i := 0; i < len(gameKeys); i += chunkSize {
		end := i + chunkSize
		if end > len(gameKeys) {
			end = len(gameKeys)
		}
		chunks = append(chunks, gameKeys[i:end])
	}

	// Fetch bundles concurrently
	type bundleResult struct {
		bundles []models.Bundle
		err     error
	}

	resultChan := make(chan bundleResult, len(chunks))
	var wg sync.WaitGroup

	for _, chunk := range chunks {
		wg.Add(1)
		go func(keys []string) {
			defer wg.Done()
			bundles, err := api.readBundlesData(keys)
			resultChan <- bundleResult{bundles, err}
		}(chunk)
	}

	wg.Wait()
	close(resultChan)

	// Collect results
	allBundles := []models.Bundle{}
	for result := range resultChan {
		if result.err != nil {
			return nil, result.err
		}
		allBundles = append(allBundles, result.bundles...)
	}

	// Sort by creation date (oldest first)
	sort.Slice(allBundles, func(i, j int) bool {
		return allBundles[i].Created.Time.Before(allBundles[j].Created.Time)
	})

	return allBundles, nil
}

// readBundlesData fetches bundle data for a chunk of keys
func (api *HumbleApi) readBundlesData(keys []string) ([]models.Bundle, error) {
	// Build query parameters
	params := url.Values{}
	params.Set("all_tpkds", "true")
	for _, key := range keys {
		params.Add("gamekeys", key)
	}

	reqURL := fmt.Sprintf("%s?%s", apiBundlesURL, params.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, NewNetworkError(err)
	}

	req.Header.Set("Accept", "application/json")
	api.addAuthCookie(req)

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, NewNetworkError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewNetworkError(fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewNetworkError(err)
	}

	var bundleMap models.BundleMap
	if err := json.Unmarshal(body, &bundleMap); err != nil {
		return nil, NewDeserializeError(err)
	}

	bundles := make([]models.Bundle, 0, len(bundleMap))
	for _, bundle := range bundleMap {
		bundles = append(bundles, bundle)
	}

	return bundles, nil
}

// ReadBundle fetches details for a specific bundle
func (api *HumbleApi) ReadBundle(productKey string) (*models.Bundle, error) {
	reqURL := fmt.Sprintf("%s/%s?all_tpkds=true", apiBundleDetailsURL, productKey)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, NewNetworkError(err)
	}

	req.Header.Set("Accept", "application/json")
	api.addAuthCookie(req)

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, NewNetworkError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewNetworkError(fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewNetworkError(err)
	}

	var bundle models.Bundle
	if err := json.Unmarshal(body, &bundle); err != nil {
		return nil, NewDeserializeError(err)
	}

	return &bundle, nil
}

// ReadBundleChoices fetches Humble Choice data for a given period
// when should be in "month-year" format (e.g., "january-2023") or "home" for current
func (api *HumbleApi) ReadBundleChoices(when string) (*models.HumbleChoice, error) {
	reqURL := fmt.Sprintf("%s/%s", membershipURL, when)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, NewNetworkError(err)
	}

	api.addAuthCookie(req)

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, NewNetworkError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewNetworkError(fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewNetworkError(err)
	}

	return api.parseBundleChoices(string(body))
}

// parseBundleChoices extracts choice data from HTML
func (api *HumbleApi) parseBundleChoices(html string) (*models.HumbleChoice, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, NewDeserializeError(err)
	}

	// Find the script tag with choice data
	var jsonData string
	doc.Find("script#webpack-subscriber-hub-data, script#webpack-monthly-product-data").Each(func(i int, s *goquery.Selection) {
		if jsonData == "" {
			jsonData = s.Text()
		}
	})

	if jsonData == "" {
		return nil, NewBundleNotFoundError()
	}

	var choice models.HumbleChoice
	if err := json.Unmarshal([]byte(jsonData), &choice); err != nil {
		return nil, NewDeserializeError(err)
	}

	return &choice, nil
}
