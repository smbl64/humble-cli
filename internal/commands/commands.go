package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/smbl64/humble-cli/internal/api"
	"github.com/smbl64/humble-cli/internal/config"
	"github.com/smbl64/humble-cli/internal/download"
	"github.com/smbl64/humble-cli/internal/keymatch"
	"github.com/smbl64/humble-cli/internal/models"
	"github.com/smbl64/humble-cli/internal/util"
)

var validFields = []string{"key", "name", "size", "claimed"}

// HandleHTTPErrors maps API errors to user-friendly messages
func HandleHTTPErrors(err error) error {
	if err == nil {
		return nil
	}

	apiErr, ok := err.(*api.ApiError)
	if !ok {
		return fmt.Errorf("failed: %w", err)
	}

	if apiErr.Type == api.NetworkError && apiErr.Err != nil {
		// Check for HTTP status codes in error message
		errMsg := apiErr.Err.Error()
		if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "Unauthorized") {
			return fmt.Errorf("Unauthorized request (401). Is the session key correct?")
		}
		if strings.Contains(errMsg, "404") || strings.Contains(errMsg, "Not Found") {
			return fmt.Errorf("Bundle not found (404). Is the bundle key correct?")
		}
		return fmt.Errorf("failed with network error: %w", apiErr.Err)
	}

	return fmt.Errorf("failed: %w", err)
}

// ListHumbleChoices displays Humble Choice games for a given period
func ListHumbleChoices(period models.ChoicePeriod) error {
	sessionKey, err := config.GetConfig()
	if err != nil {
		return err
	}

	apiClient := api.New(sessionKey)
	periodStr := period.String()
	if period.IsCurrent() {
		periodStr = "home"
	}

	choices, err := apiClient.ReadBundleChoices(periodStr)
	if err != nil {
		return HandleHTTPErrors(err)
	}

	fmt.Println()
	fmt.Println(choices.Options.Title)
	fmt.Println()

	table := util.NewTable([]string{"#", "Title", "Redeemed"})
	counter := 1
	allRedeemed := true

	for _, gameData := range choices.Options.Data.GameData {
		for _, tpkd := range gameData.Tpkds {
			table.AddRow([]string{
				fmt.Sprintf("%d", counter),
				tpkd.HumanName,
				tpkd.ClaimStatus(),
			})
			counter++

			if tpkd.ClaimStatus() == "No" {
				allRedeemed = false
			}
		}
	}

	table.Render()

	if !allRedeemed {
		url := "https://www.humblebundle.com/membership/home"
		fmt.Printf("Visit %s to redeem your keys.\n", url)
	}

	return nil
}

// Search searches for products matching keywords
func Search(keywords string, matchMode models.MatchMode) error {
	sessionKey, err := config.GetConfig()
	if err != nil {
		return err
	}

	apiClient := api.New(sessionKey)
	bundles, err := apiClient.ListBundles()
	if err != nil {
		return HandleHTTPErrors(err)
	}

	lowercaseKeywords := strings.ToLower(keywords)
	kwList := strings.Fields(lowercaseKeywords)

	type BundleItem struct {
		bundle      *models.Bundle
		productName string
	}

	searchResult := []BundleItem{}
	for i := range bundles {
		for j := range bundles[i].Products {
			if bundles[i].Products[j].NameMatches(kwList, matchMode) {
				searchResult = append(searchResult, BundleItem{
					bundle:      &bundles[i],
					productName: bundles[i].Products[j].HumanName,
				})
			}
		}
	}

	if len(searchResult) == 0 {
		fmt.Println("Nothing found")
		return nil
	}

	table := util.NewTable([]string{"Key", "Name", "Sub Item"})
	for _, item := range searchResult {
		table.AddRow([]string{
			item.bundle.Gamekey,
			item.bundle.Details.HumanName,
			item.productName,
		})
	}

	table.Render()
	return nil
}

// ListBundles lists all bundles with optional filtering
func ListBundles(fields []string, claimedFilter string) error {
	sessionKey, err := config.GetConfig()
	if err != nil {
		return err
	}

	apiClient := api.New(sessionKey)

	// Optimization: if only key field is requested and no filter, just list keys
	keyOnly := len(fields) == 1 && fields[0] == "key"
	if keyOnly && claimedFilter == "all" {
		keys, err := apiClient.ListBundleKeys()
		if err != nil {
			return HandleHTTPErrors(err)
		}
		for _, key := range keys {
			fmt.Println(key)
		}
		return nil
	}

	bundles, err := apiClient.ListBundles()
	if err != nil {
		return HandleHTTPErrors(err)
	}

	// Filter by claimed status
	if claimedFilter != "all" {
		claimed := claimedFilter == "yes"
		filtered := []models.Bundle{}
		for _, b := range bundles {
			status := b.ClaimStatus()
			if (status == "Yes" && claimed) || (status == "No" && !claimed) {
				filtered = append(filtered, b)
			}
		}
		bundles = filtered
	}

	// CSV output mode
	if len(fields) > 0 {
		return bulkFormat(fields, bundles)
	}

	// Table output mode
	fmt.Printf("%d bundle(s) found.\n\n", len(bundles))

	if len(bundles) == 0 {
		return nil
	}

	table := util.NewTable([]string{"Key", "Name", "Size", "Claimed"})
	for _, bundle := range bundles {
		table.AddRow([]string{
			bundle.Gamekey,
			bundle.Details.HumanName,
			util.HumanizeBytes(bundle.TotalSize()),
			bundle.ClaimStatus(),
		})
	}

	table.Render()
	return nil
}

// findKey searches for a matching bundle key
func findKey(allKeys []string, keyToFind string) (string, error) {
	key, err := keymatch.FindKey(allKeys, keyToFind)
	if err != nil {
		return "", err
	}
	return key, nil
}

// ShowBundleDetails displays detailed information about a specific bundle
func ShowBundleDetails(bundleKey string) error {
	sessionKey, err := config.GetConfig()
	if err != nil {
		return err
	}

	apiClient := api.New(sessionKey)
	allKeys, err := apiClient.ListBundleKeys()
	if err != nil {
		return HandleHTTPErrors(err)
	}

	foundKey, err := findKey(allKeys, bundleKey)
	if err != nil {
		return err
	}

	bundle, err := apiClient.ReadBundle(foundKey)
	if err != nil {
		return HandleHTTPErrors(err)
	}

	fmt.Println()
	fmt.Println(bundle.Details.HumanName)
	fmt.Println()
	fmt.Printf("Purchased    : %s\n", bundle.Created.Time.Format("2006-01-02"))
	if bundle.AmountSpent != nil && bundle.Currency != nil {
		fmt.Printf("Amount spent : %.2f %s\n", *bundle.AmountSpent, *bundle.Currency)
	}
	fmt.Printf("Total size   : %s\n", util.HumanizeBytes(bundle.TotalSize()))
	fmt.Println()

	if len(bundle.Products) > 0 {
		table := util.NewTable([]string{"#", "Sub-item", "Format", "Total Size"})
		for idx, product := range bundle.Products {
			table.AddRow([]string{
				fmt.Sprintf("%d", idx+1),
				product.HumanName,
				product.Formats(),
				util.HumanizeBytes(product.TotalSize()),
			})
		}
		table.Render()
	} else {
		fmt.Println("No items to show.")
	}

	// Product keys
	productKeys := bundle.ProductKeys()
	if len(productKeys) > 0 {
		fmt.Println()
		fmt.Println("Keys in this bundle:")
		fmt.Println()

		table := util.NewTable([]string{"#", "Key Name", "Redeemed"})
		allRedeemed := true

		for idx, pk := range productKeys {
			redeemed := "No"
			if pk.Redeemed {
				redeemed = "Yes"
			} else {
				allRedeemed = false
			}

			table.AddRow([]string{
				fmt.Sprintf("%d", idx+1),
				pk.HumanName,
				redeemed,
			})
		}

		table.Render()

		if !allRedeemed {
			url := "https://www.humblebundle.com/home/keys"
			fmt.Printf("Visit %s to redeem your keys.\n", url)
		}
	}

	return nil
}

// DownloadBundle downloads files from a specific bundle
func DownloadBundle(bundleKey string, formats []string, maxSize uint64, itemNumbers string, torrentsOnly, curDir bool) error {
	sessionKey, err := config.GetConfig()
	if err != nil {
		return err
	}

	apiClient := api.New(sessionKey)
	allKeys, err := apiClient.ListBundleKeys()
	if err != nil {
		return HandleHTTPErrors(err)
	}

	foundKey, err := findKey(allKeys, bundleKey)
	if err != nil {
		return err
	}

	bundle, err := apiClient.ReadBundle(foundKey)
	if err != nil {
		return HandleHTTPErrors(err)
	}

	// Parse item numbers
	var itemNumbersList []int
	if itemNumbers != "" {
		ranges := strings.Split(itemNumbers, ",")
		itemNumbersList, err = util.UnionUsizeRanges(ranges, len(bundle.Products))
		if err != nil {
			return err
		}
	}

	// Filter products
	products := []*models.Product{}
	for i := range bundle.Products {
		product := &bundle.Products[i]

		// Filter by item number
		if len(itemNumbersList) > 0 {
			found := false
			for _, num := range itemNumbersList {
				if num == i+1 {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filter by size
		if maxSize > 0 && product.TotalSize() >= maxSize {
			continue
		}

		// Filter by format
		if len(formats) > 0 && !util.StrVectorsIntersect(product.FormatsAsVec(), formats) {
			continue
		}

		products = append(products, product)
	}

	if len(products) == 0 {
		fmt.Println("Nothing to download")
		return nil
	}

	// Create bundle directory
	dirName := util.ReplaceInvalidCharsInFilename(bundle.Details.HumanName)
	var bundleDir string
	if curDir {
		bundleDir = "."
	} else {
		bundleDir = dirName
		if err := os.MkdirAll(bundleDir, 0755); err != nil {
			return fmt.Errorf("failed to create bundle directory: %w", err)
		}
	}

	// Download products
	for _, product := range products {
		fmt.Println()
		fmt.Println(product.HumanName)

		productDirName := util.ReplaceInvalidCharsInFilename(product.HumanName)
		productDir := filepath.Join(bundleDir, productDirName)
		if err := os.MkdirAll(productDir, 0755); err != nil {
			return fmt.Errorf("failed to create product directory: %w", err)
		}

		for _, productDownload := range product.Downloads {
			for _, dlInfo := range productDownload.Items {
				// Filter by format
				if len(formats) > 0 {
					formatLower := strings.ToLower(dlInfo.Format)
					found := false
					for _, f := range formats {
						if strings.ToLower(f) == formatLower {
							found = true
							break
						}
					}
					if !found {
						fmt.Printf("Skipping '%s'\n", dlInfo.Format)
						continue
					}
				}

				// Choose download URL
				downloadURL := dlInfo.URL.Web
				if torrentsOnly {
					downloadURL = dlInfo.URL.Bittorrent
				}

				// Extract filename
				filename := util.ExtractFilenameFromURL(downloadURL)
				if filename == "" {
					return fmt.Errorf("cannot get file name from URL '%s'", downloadURL)
				}

				downloadPath := filepath.Join(productDir, filename)

				// Download file
				err := download.DownloadFile(downloadURL, downloadPath, filename, dlInfo.FileSize)
				if err != nil {
					return fmt.Errorf("download failed: %w", err)
				}
			}
		}
	}

	return nil
}

// DownloadBundles downloads multiple bundles from a file
func DownloadBundles(bundleListFile string, formats []string, maxSize uint64, torrentsOnly, curDir bool) error {
	data, err := os.ReadFile(bundleListFile)
	if err != nil {
		return fmt.Errorf("failed to read bundle list file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	errors := []struct {
		name string
		err  error
	}{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, ",")
		bundleKey := parts[0]
		bundleName := bundleKey
		if len(parts) > 1 {
			bundleName = parts[1]
		}

		err := DownloadBundle(bundleKey, formats, maxSize, "", torrentsOnly, curDir)
		if err != nil {
			errors = append(errors, struct {
				name string
				err  error
			}{bundleName, err})
		}
	}

	// Print errors
	for _, e := range errors {
		fmt.Printf("Error handling: %s\n", e.name)
		fmt.Printf("Error: %v\n", e.err)
	}

	return nil
}

// validateFields checks if all fields are valid
func validateFields(fields []string) bool {
	for _, field := range fields {
		fieldLower := strings.ToLower(field)
		found := false
		for _, valid := range validFields {
			if fieldLower == valid {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// bulkFormat outputs bundles in CSV format
func bulkFormat(fields []string, bundles []models.Bundle) error {
	if !validateFields(fields) {
		return fmt.Errorf("invalid field in fields: %s", strings.Join(fields, ","))
	}

	printKey := contains(fields, "key")
	printName := contains(fields, "name")
	printSize := contains(fields, "size")
	printClaimed := contains(fields, "claimed")

	for _, bundle := range bundles {
		printVec := []string{}

		if printKey {
			printVec = append(printVec, bundle.Gamekey)
		}
		if printName {
			printVec = append(printVec, bundle.Details.HumanName)
		}
		if printSize {
			printVec = append(printVec, util.HumanizeBytes(bundle.TotalSize()))
		}
		if printClaimed {
			printVec = append(printVec, bundle.ClaimStatus())
		}

		fmt.Println(strings.Join(printVec, ","))
	}

	return nil
}

// contains checks if a slice contains a string (case-insensitive)
func contains(slice []string, str string) bool {
	strLower := strings.ToLower(str)
	for _, item := range slice {
		if strings.ToLower(item) == strLower {
			return true
		}
	}
	return false
}
