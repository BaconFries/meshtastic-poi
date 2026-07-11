package source

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/BaconFries/meshtastic-poi/internal/downloader"
)

// Read fetches bytes from an HTTP(S) URL or local file path.
func Read(ctx context.Context, client *downloader.Client, rawURL string) ([]byte, error) {
	if rawURL == "" {
		return nil, fmt.Errorf("empty source url")
	}

	if strings.HasPrefix(rawURL, "file://") {
		path, err := fileURLPath(rawURL)
		if err != nil {
			return nil, err
		}
		return os.ReadFile(path)
	}

	if isHTTP(rawURL) {
		return client.Get(ctx, rawURL)
	}

	data, err := os.ReadFile(rawURL)
	if err != nil {
		return nil, fmt.Errorf("read local file %q: %w", rawURL, err)
	}
	return data, nil
}

func fileURLPath(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	return u.Path, nil
}

func isHTTP(raw string) bool {
	return strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://")
}
