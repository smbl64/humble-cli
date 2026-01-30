package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

const (
	retryCount   = 3
	retryDelay   = 5 * time.Second
	readTimeout  = 30 * time.Second
	writeTimeout = 30 * time.Second
)

// DownloadFile downloads a file from url to filePath with retry logic and progress bar
func DownloadFile(url, filePath, title string, expectedSize uint64) error {
	retries := retryCount

	for retries > 0 {
		err := downloadFileAttempt(url, filePath, title, expectedSize)
		if err == nil {
			return nil
		}

		retries--
		if retries > 0 && isRetryable(err) {
			fmt.Printf("  Will retry in %d seconds...\n", int(retryDelay.Seconds()))
			time.Sleep(retryDelay)
			continue
		}

		return err
	}

	return fmt.Errorf("download failed after %d attempts", retryCount)
}

// isRetryable checks if an error should trigger a retry
func isRetryable(err error) bool {
	// Retry on network errors and timeouts
	// Don't retry on 4xx errors (except 408, 429)
	// Retry on 5xx errors
	if err == nil {
		return false
	}

	// Check for timeout errors
	if os.IsTimeout(err) {
		return true
	}

	// For HTTP errors, check status code
	// This is a simplified check - in production you'd want more sophisticated error type checking
	return true
}

// downloadFileAttempt performs a single download attempt
func downloadFileAttempt(url, filePath, title string, expectedSize uint64) error {
	// Check if file exists and get current size
	file, downloaded, err := openFileForWrite(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// If file is already complete, skip
	if downloaded >= expectedSize {
		fmt.Println("  Nothing to do. File already exists.")
		return nil
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: readTimeout,
	}

	// Create request with Range header for resume capability
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if downloaded > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", downloaded))
	}

	// Perform request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Create progress bar
	bar := progressbar.NewOptions64(
		int64(expectedSize),
		progressbar.OptionSetDescription(fmt.Sprintf("  Downloading %s", title)),
		progressbar.OptionSetWriter(os.Stdout),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Println()
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	)

	// Set initial progress if resuming
	if downloaded > 0 {
		bar.Set64(int64(downloaded))
	}

	// Download in chunks
	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := file.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("failed to write to file: %w", writeErr)
			}

			downloaded += uint64(n)
			bar.Set64(int64(downloaded))
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading response: %w", err)
		}
	}

	bar.Finish()
	fmt.Printf("  Downloaded %s\n", title)
	return nil
}

// openFileForWrite opens a file for writing, creating it if necessary
// Returns the file, current size, and any error
func openFileForWrite(filePath string) (*os.File, uint64, error) {
	// Check if file exists
	info, err := os.Stat(filePath)
	if err == nil {
		// File exists, open for append
		file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, 0, err
		}
		return file, uint64(info.Size()), nil
	}

	if !os.IsNotExist(err) {
		return nil, 0, err
	}

	// File doesn't exist, create it
	file, err := os.Create(filePath)
	if err != nil {
		return nil, 0, err
	}

	return file, 0, nil
}

// GetContentLength fetches the content length from a URL without downloading
func GetContentLength(url string) (uint64, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Head(url)
	if err != nil {
		// Some servers don't support HEAD, try GET with no body read
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return 0, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err = client.Do(req)
		if err != nil {
			return 0, fmt.Errorf("failed to get content length: %w", err)
		}
		defer resp.Body.Close()
	} else {
		defer resp.Body.Close()
	}

	if resp.ContentLength < 0 {
		return 0, fmt.Errorf("failed to get content length from '%s'", url)
	}

	return uint64(resp.ContentLength), nil
}
