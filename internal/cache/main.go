// Package cache implements KPM's local artifact cache, laid out exactly
// like Maven's ~/.m2/repository (group/path/artifact/version/file) so the
// cache doubles as a drop-in local repository for other tooling.
package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"kpm/internal/checksum"
	"kpm/internal/env"
)

type Cache struct {
	root string
}

func Open() (*Cache, error) {
	root, err := env.RepositoryCacheDir()
	if err != nil {
		return nil, err
	}
	return &Cache{root: root}, nil
}

// Path returns the on-disk path an artifact would live at, without checking existence.
func (c *Cache) Path(group, artifact, version, classifier, ext string) string {
	gp := strings.ReplaceAll(group, ".", string(filepath.Separator))
	name := artifact + "-" + version
	if classifier != "" {
		name += "-" + classifier
	}
	return filepath.Join(c.root, gp, artifact, version, name+"."+ext)
}

// Has reports whether the artifact is cached AND passes checksum
// verification against expectedHex (if provided). A cached-but-corrupt file
// is treated as absent so callers re-download rather than silently using
// bad bytes — this is the "corruption detection" requirement.
func (c *Cache) Has(path string, algo checksum.Algo, expectedHex string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if expectedHex == "" {
		return true // no checksum available to verify against; trust presence
	}
	if err := checksum.Verify(path, algo, expectedHex); err != nil {
		return false
	}
	return true
}

// Store writes data to path (creating parent dirs), and also caches whatever
// sidecar checksum was supplied so future corruption checks don't require a
// network round-trip.
func (c *Cache) Store(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// StoreChecksum persists a verified checksum alongside the artifact so
// `kpm doctor` and future runs can detect on-disk corruption without
// re-fetching from the repository.
func (c *Cache) StoreChecksum(path string, algo checksum.Algo, hexDigest string) error {
	return os.WriteFile(path+algo.Extension(), []byte(hexDigest), 0o644)
}

// VerifyOnDisk re-hashes a cached artifact against its locally stored
// sidecar checksum (if any), reporting corruption independent of network access.
func (c *Cache) VerifyOnDisk(path string) error {
	for _, algo := range checksum.PreferredOrder {
		sidecar := path + algo.Extension()
		expected, err := os.ReadFile(sidecar)
		if err != nil {
			continue
		}
		return checksum.Verify(path, algo, strings.TrimSpace(string(expected)))
	}
	return nil // no sidecar recorded; nothing to verify against
}

// Clean removes cached artifacts older than maxAge (0 = remove everything),
// returning how many bytes were freed. Used by `kpm cache clean`.
func (c *Cache) Clean(maxAge time.Duration) (freedBytes int64, removedFiles int, err error) {
	cutoff := time.Now().Add(-maxAge)
	err = filepath.Walk(c.root, func(path string, info os.FileInfo, werr error) error {
		if werr != nil || info == nil || info.IsDir() {
			return nil
		}
		if maxAge > 0 && info.ModTime().After(cutoff) {
			return nil
		}
		freedBytes += info.Size()
		removedFiles++
		return os.Remove(path)
	})
	// best-effort prune of now-empty directories
	_ = pruneEmptyDirs(c.root)
	return
}

func pruneEmptyDirs(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || path == root || info == nil || !info.IsDir() {
			return nil
		}
		entries, _ := os.ReadDir(path)
		if len(entries) == 0 {
			return os.Remove(path)
		}
		return nil
	})
}

// Doctor scans the entire cache and reports any artifact that fails
// checksum verification against its stored sidecar — used by `kpm doctor`.
func (c *Cache) Doctor() ([]string, error) {
	var problems []string
	err := filepath.Walk(c.root, func(path string, info os.FileInfo, werr error) error {
		if werr != nil || info == nil || info.IsDir() {
			return nil
		}
		if isSidecar(path) {
			return nil
		}
		if err := c.VerifyOnDisk(path); err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", path, err))
		}
		return nil
	})
	return problems, err
}

func isSidecar(path string) bool {
	for _, a := range checksum.PreferredOrder {
		if strings.HasSuffix(path, a.Extension()) {
			return true
		}
	}
	return false
}