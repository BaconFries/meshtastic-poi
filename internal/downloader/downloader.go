package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// Client performs HTTP downloads with retries and optional disk caching.
type Client struct {
	httpClient *retryablehttp.Client
	cacheDir   string
}

// New creates a downloader client.
func New(cacheDir string) *Client {
	rc := retryablehttp.NewClient()
	rc.RetryMax = 5
	rc.RetryWaitMin = 500 * time.Millisecond
	rc.RetryWaitMax = 10 * time.Second
	rc.HTTPClient.Timeout = 120 * time.Second
	return &Client{
		httpClient: rc,
		cacheDir:   cacheDir,
	}
}

// Get performs an HTTP GET and returns the response body.
func (c *Client) Get(ctx context.Context, url string) ([]byte, error) {
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "meshtastic-poi/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("HTTP %d from %s: %s", resp.StatusCode, url, string(body))
	}
	return io.ReadAll(resp.Body)
}

// Post performs an HTTP POST and returns the response body.
func (c *Client) Post(ctx context.Context, rawURL, contentType string, body io.Reader) ([]byte, error) {
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, rawURL, body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("User-Agent", "meshtastic-poi/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("HTTP %d from %s: %s", resp.StatusCode, rawURL, string(respBody))
	}
	return io.ReadAll(resp.Body)
}

// GetJSON is an alias for Get with JSON accept header (same as Get).
func (c *Client) GetJSON(ctx context.Context, url string) ([]byte, error) {
	return c.Get(ctx, url)
}

// CachePath returns the on-disk cache path for a URL key.
func (c *Client) CachePath(key string) string {
	if c.cacheDir == "" {
		return ""
	}
	return filepath.Join(c.cacheDir, key)
}

// EnsureCacheDir creates the cache directory if needed.
func (c *Client) EnsureCacheDir() error {
	if c.cacheDir == "" {
		return nil
	}
	return os.MkdirAll(c.cacheDir, 0o755)
}
