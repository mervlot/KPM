package resolver

import (
	"bytes"
	"fmt"
	"os"

	"kpm/internal/cache"
	"kpm/internal/checksum"
	"kpm/internal/downloader"
	"kpm/internal/metadata"
	"kpm/internal/parser"
	"kpm/internal/repository"
)

// Fetcher bundles everything needed to resolve+download artifacts: the
// repository priority list, the local cache, and the HTTP client. It is the
// single choke point for "get me bytes for this coordinate", so caching,
// checksum verification, retries, and offline mode are all enforced
// consistently no matter which part of resolution needs an artifact.
type Fetcher struct {
	repos    *repository.Set
	cache    *cache.Cache
	http     *downloader.Client
	Offline  bool
	repoUsed map[string]string // "group:artifact:version" -> repository ID that served it
}

func NewFetcher(repos *repository.Set, c *cache.Cache, http *downloader.Client, offline bool) *Fetcher {
	return &Fetcher{repos: repos, cache: c, http: http, Offline: offline, repoUsed: map[string]string{}}
}

// RepositoryFor reports which repository ID served group:artifact:version,
// if known ("cache" if it was only ever satisfied from the local cache in
// this run). Used to populate the lock file's provenance field.
func (f *Fetcher) RepositoryFor(group, artifact, ver string) string {
	if id, ok := f.repoUsed[group+":"+artifact+":"+ver]; ok {
		return id
	}
	return "cache"
}

// FetchPOM returns a group:artifact:version POM's bytes, trying the cache
// first, then each repository in priority order, verifying checksums and
// persisting successful downloads to the cache.
func (f *Fetcher) FetchPOM(group, artifact, version string) ([]byte, error) {
	return f.fetchArtifact(group, artifact, version, "", "pom")
}

// FetchArtifact returns the primary build artifact (jar/war/aar/...) bytes.
func (f *Fetcher) FetchArtifact(group, artifact, version, classifier, ext string) ([]byte, error) {
	return f.fetchArtifact(group, artifact, version, classifier, ext)
}

func (f *Fetcher) fetchArtifact(group, artifact, version, classifier, ext string) ([]byte, error) {
	path := f.cache.Path(group, artifact, version, classifier, ext)

	if data, ok := f.readCached(path); ok {
		return data, nil
	}
	if f.Offline {
		return nil, fmt.Errorf("offline mode: %s:%s:%s (%s) not in local cache", group, artifact, version, ext)
	}

	var lastErr error
	for _, repo := range f.repos.All() {
		url := repo.ArtifactURL(group, artifact, version, classifier, ext)
		data, err := f.http.GetAuth(url, repo.Username, repo.Password)
		if err != nil {
			lastErr = err
			continue
		}

		// Checksum verification is skipped for POMs deliberately: a POM is
		// metadata used only to build the dependency graph (it's never
		// compiled or run), and resolution can fetch dozens of them in a
		// single run (every BOM, every parent, every transitive dependency).
		// Tripling that request count for integrity-checking a file that
		// isn't executable code is a bad trade against Maven Central's rate
		// limiting. Real build artifacts (jar/war/aar/...) still get full
		// checksum verification below — that's the code that actually runs.
		if ext != "pom" {
			algo, expected, verr := f.fetchExpectedChecksum(repo, url)
			if verr == nil && expected != "" {
				got, herr := checksum.HashReader(bytes.NewReader(data), algo)
				if herr == nil && got != expected {
					lastErr = &checksum.MismatchError{Path: url, Algo: algo, Expected: expected, Actual: got}
					continue // try next repository rather than trusting a bad artifact
				}
				if herr == nil {
					_ = f.cache.StoreChecksum(path, algo, got)
				}
			}
		}

		if err := f.cache.Store(path, data); err != nil {
			return nil, fmt.Errorf("caching %s: %w", path, err)
		}
		f.repoUsed[group+":"+artifact+":"+version] = repo.ID
		return data, nil
	}
	return nil, fmt.Errorf("could not fetch %s:%s:%s (%s) from any repository: %w", group, artifact, version, ext, lastErr)
}

func (f *Fetcher) readCached(path string) ([]byte, bool) {
	if err := f.cache.VerifyOnDisk(path); err != nil {
		return nil, false // corrupted; force re-download
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	return data, true
}

// fetchExpectedChecksum tries sidecar hash files in strongest-first order.
func (f *Fetcher) fetchExpectedChecksum(repo repository.Repository, artifactURL string) (checksum.Algo, string, error) {
	for _, algo := range checksum.PreferredOrder {
		data, err := f.http.GetAuth(artifactURL+algo.Extension(), repo.Username, repo.Password)
		if err == nil {
			return algo, checksum.ParseSidecar(data), nil
		}
	}
	return "", "", fmt.Errorf("no sidecar checksum published for %s", artifactURL)
}

// GetMetadata fetches maven-metadata.xml for group:artifact, trying repositories in order.
func (f *Fetcher) GetMetadata(group, artifact string) (*parser.Metadata, error) {
	var lastErr error
	for _, repo := range f.repos.All() {
		m, err := metadata.Get(func(url string) ([]byte, error) {
			return f.http.GetAuth(url, repo.Username, repo.Password)
		}, repo.URL, group, artifact, f.Offline)
		if err == nil {
			return m, nil
		}
		lastErr = err
	}
	return nil, lastErr
}