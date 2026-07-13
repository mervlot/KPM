// Package env centralizes OS detection and the standard on-disk locations
// KPM uses, analogous to Maven's ~/.m2.
package env

import (
	"os"
	"path/filepath"
	"runtime"
)

func DetectOS() string { return runtime.GOOS }

// HomeCacheDir returns "~/.kpm" (or the KPM_HOME override), creating it if
// necessary. This is the root for the artifact cache, metadata cache and
// global config — the equivalent of Maven's ~/.m2/repository.
func HomeCacheDir() (string, error) {
	if v := os.Getenv("KPM_HOME"); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".kpm")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// RepositoryCacheDir returns "~/.kpm/repository", where downloaded jars/poms
// live, keyed exactly like Maven's local repo layout
// (group/as/path/artifact/version/artifact-version.ext) so tooling that
// understands an m2 layout can point at it directly.
func RepositoryCacheDir() (string, error) {
	root, err := HomeCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, "repository")
	return dir, os.MkdirAll(dir, 0o755)
}

// MetadataCacheDir returns "~/.kpm/metadata-cache" used for maven-metadata.xml TTL caching.
func MetadataCacheDir() (string, error) {
	root, err := HomeCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, "metadata-cache")
	return dir, os.MkdirAll(dir, 0o755)
}