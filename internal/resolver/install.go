package resolver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"kpm/internal/downloader"
)

// InstallPlan represents a single artifact to be downloaded.
type InstallPlan struct {
	GroupID    string
	ArtifactID string
	Version    string
	Extension  string // "pom" or "jar"
}

// BuildInstallPlan creates a list of artifacts to download based on the resolution result.
// This matches the signature expected by internal/cli/command.go
func BuildInstallPlan(result ResolutionResult, libsDir string) []InstallPlan {
	var plans []InstallPlan
	for _, winner := range result.Winners {
		// We need both POM (for transitive resolution) and JAR (for compilation) for each winner
		plans = append(plans, InstallPlan{
			GroupID:    winner.Group,
			ArtifactID: winner.Artifact,
			Version:    winner.Version,
			Extension:  "pom",
		})
		plans = append(plans, InstallPlan{
			GroupID:    winner.Group,
			ArtifactID: winner.Artifact,
			Version:    winner.Version,
			Extension:  "jar",
		})
	}
	return plans
}

// fetchArtifact is the core "offline-first" function.
// It checks the local cache first. If missing, it downloads from the network and saves to cache.
func fetchArtifact(groupID, artifactID, version, ext string, client *downloader.Client) (string, error) {
	groupPath := filepath.Join("libs", filepath.FromSlash(groupID))
	artifactPath := filepath.Join(groupPath, artifactID, version)
	fileName := fmt.Sprintf("%s-%s.%s", artifactID, version, ext)
	localPath := filepath.Join(artifactPath, fileName)

	// 1. OFFLINE-FIRST: Check local cache first
	info, err := os.Stat(localPath)
	if err == nil && !info.IsDir() && info.Size() > 0 {
		// Found in cache! Zero network requests.
		return localPath, nil
	}

	// 2. NETWORK FALLBACK: Only hit the network if not cached
	mavenGroupPath := strings.ReplaceAll(groupID, ".", "/")
	url := fmt.Sprintf("https://repo1.maven.org/maven2/%s/%s/%s/%s-%s.%s",
		mavenGroupPath, artifactID, version, artifactID, version, ext)

	data, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %s: %w", url, err)
	}

	// 3. SAVE TO CACHE: Persist to disk for future offline-first use
	if err := os.MkdirAll(artifactPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache dir %s: %w", artifactPath, err)
	}

	if err := os.WriteFile(localPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write to cache %s: %w", localPath, err)
	}

	return localPath, nil
}

// InstallWithProgress downloads the planned artifacts sequentially with progress tracking.
// This method attaches to the existing Resolver struct defined elsewhere in your package.
func (r *Resolver) InstallWithProgress(plans []InstallPlan, concurrency int, stepFunc func()) []error {
	// Enforce strict rate limiting (500ms = 2 requests per second) to avoid Maven Central 429s
	downloader.SetGlobalPace(500 * time.Millisecond)

	var jobs []downloader.Job
	var errs []error

	for _, plan := range plans {
		plan := plan // capture loop variable for closure
		job := downloader.Job{
			Name: fmt.Sprintf("%s:%s:%s (%s)", plan.GroupID, plan.ArtifactID, plan.Version, plan.Extension),
			Run: func() error {
				// Note: If your Resolver struct uses a different field name for the downloader 
				// (e.g., r.downloader instead of r.Client), update "r.Client" below to match.
				localPath, err := fetchArtifact(plan.GroupID, plan.ArtifactID, plan.Version, plan.Extension, r.Client)
				if err != nil {
					return err
				}
				if stepFunc != nil {
					stepFunc()
				}
				return nil
			},
		}
		jobs = append(jobs, job)
	}

	// Execute sequentially (no goroutines) via the updated downloader.RunPool.
	// The 'concurrency' argument is intentionally ignored here to guarantee 
	// sequential execution, preventing burst requests that trigger 429s.
	jobErrs := downloader.RunPool(jobs, 1)
	
	for _, err := range jobErrs {
		errs = append(errs, err)
	}

	return errs
}