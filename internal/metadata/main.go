// Package metadata fetches and caches maven-metadata.xml, and resolves
// SNAPSHOT coordinates to their concrete timestamped build (SNAPSHOT
// artifacts on a real Maven repo are published as e.g.
// 1.2.3-20240102.153000-4, not literally "1.2.3-SNAPSHOT" on disk).
package metadata

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"kpm/internal/env"
	"kpm/internal/parser"
)

// Fetcher downloads raw bytes for a URL. Implemented by internal/downloader
// so this package doesn't depend on HTTP directly.
type Fetcher func(url string) ([]byte, error)

const defaultTTL = 24 * time.Hour

// snapshotTTL is much shorter: snapshot metadata changes frequently and a
// day-old cache would defeat the purpose of using a SNAPSHOT at all.
const snapshotTTL = 5 * time.Minute

// Get returns parsed maven-metadata.xml for group:artifact, using a local
// TTL cache keyed by repository+group+artifact so repeated resolutions
// within the same window don't re-hit the network. offline forces
// cache-only lookups (erroring if nothing cached).
func Get(fetch Fetcher, repoURL, group, artifact string, offline bool) (*parser.Metadata, error) {
	cacheDir, err := env.MetadataCacheDir()
	if err != nil {
		return nil, err
	}
	cacheFile := filepath.Join(cacheDir, safeName(repoURL, group, artifact)+".xml")

	if data, fresh := readIfFresh(cacheFile, defaultTTL); fresh {
		return parser.ParseMetadata(data)
	}
	if offline {
		if data, err := os.ReadFile(cacheFile); err == nil {
			return parser.ParseMetadata(data)
		}
		return nil, fmt.Errorf("offline mode: no cached metadata for %s:%s", group, artifact)
	}

	url := fmt.Sprintf("%s/%s/%s/maven-metadata.xml", trimSlash(repoURL), groupPath(group), artifact)
	data, err := fetch(url)
	if err != nil {
		// Network failed — fall back to stale cache if we have one, rather
		// than hard-failing (this is what "resilient offline-ish" behavior
		// looks like without a user explicitly requesting --offline).
		if stale, serr := os.ReadFile(cacheFile); serr == nil {
			return parser.ParseMetadata(stale)
		}
		return nil, fmt.Errorf("fetching metadata for %s:%s: %w", group, artifact, err)
	}
	_ = os.WriteFile(cacheFile, data, 0o644)
	return parser.ParseMetadata(data)
}

// GetSnapshot fetches the version-scoped maven-metadata.xml (living inside
// the version directory) used to resolve a -SNAPSHOT coordinate to its
// current timestamped build.
func GetSnapshot(fetch Fetcher, repoURL, group, artifact, version string, offline bool) (*parser.Metadata, error) {
	cacheDir, err := env.MetadataCacheDir()
	if err != nil {
		return nil, err
	}
	cacheFile := filepath.Join(cacheDir, safeName(repoURL, group, artifact+"-"+version)+".xml")

	if data, fresh := readIfFresh(cacheFile, snapshotTTL); fresh {
		return parser.ParseMetadata(data)
	}
	if offline {
		if data, err := os.ReadFile(cacheFile); err == nil {
			return parser.ParseMetadata(data)
		}
		return nil, fmt.Errorf("offline mode: no cached snapshot metadata for %s:%s:%s", group, artifact, version)
	}

	url := fmt.Sprintf("%s/%s/%s/%s/maven-metadata.xml", trimSlash(repoURL), groupPath(group), artifact, version)
	data, err := fetch(url)
	if err != nil {
		if stale, serr := os.ReadFile(cacheFile); serr == nil {
			return parser.ParseMetadata(stale)
		}
		return nil, fmt.Errorf("fetching snapshot metadata for %s:%s:%s: %w", group, artifact, version, err)
	}
	_ = os.WriteFile(cacheFile, data, 0o644)
	return parser.ParseMetadata(data)
}

// ResolveSnapshotFilename returns the concrete artifact filename fragment
// (e.g. "1.2.3-20240102.153000-4") for a -SNAPSHOT version, or the version
// unchanged if it's not a snapshot or no timestamped build info is published
// (some internal repos keep literal "-SNAPSHOT" filenames — that's valid too).
func ResolveSnapshotFilename(m *parser.Metadata, version string) string {
	ts := m.Versioning.Snapshot.Timestamp
	bn := m.Versioning.Snapshot.BuildNumber
	if ts == "" || bn == 0 {
		return version
	}
	base := version
	if len(base) > 9 && base[len(base)-9:] == "-SNAPSHOT" {
		base = base[:len(base)-9]
	}
	return fmt.Sprintf("%s-%s-%d", base, ts, bn)
}

func readIfFresh(path string, ttl time.Duration) ([]byte, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, false
	}
	if time.Since(info.ModTime()) > ttl {
		return nil, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	return data, true
}

func groupPath(group string) string {
	out := make([]byte, 0, len(group))
	for i := 0; i < len(group); i++ {
		if group[i] == '.' {
			out = append(out, '/')
		} else {
			out = append(out, group[i])
		}
	}
	return string(out)
}

func trimSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

func safeName(parts ...string) string {
	out := ""
	for _, p := range parts {
		for _, r := range p {
			switch {
			case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
				out += string(r)
			default:
				out += "_"
			}
		}
		out += "__"
	}
	return out
}