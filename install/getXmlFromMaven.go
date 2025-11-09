// ...existing code...
package install

import (
	"encoding/xml"
	"fmt"
	"io"
	"kpm/types"
	"net/http"
	"sync"
	"time"
)

// shared HTTP client with timeout to avoid repeated default-client allocations
var httpClient = &http.Client{Timeout: 15 * time.Second}

// simple in-memory cache for maven metadata (keyed by metadata URL)
var metadataCache = struct {
	sync.RWMutex
	m map[string]*types.Metadata
}{m: make(map[string]*types.Metadata)}

// GetMavenMetadata fetches and parses Maven XML metadata from a URL with caching.
func GetMavenMetadata(url string) (*types.Metadata, error) {
	// fast path: read from cache
	metadataCache.RLock()
	if m, ok := metadataCache.m[url]; ok {
		metadataCache.RUnlock()
		return m, nil
	}
	metadataCache.RUnlock()

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metadata: %w", err)
	}
	defer resp.Body.Close()

	// Ensure we got a successful HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var meta types.Metadata
	if err := xml.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	// store in cache
	metadataCache.Lock()
	metadataCache.m[url] = &meta
	metadataCache.Unlock()

	return &meta, nil
}

// ...existing code...
