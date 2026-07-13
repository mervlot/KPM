package resolver

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"kpm/internal/downloader"
)

// InstallPlan represents a single artifact to be installed.
type InstallPlan struct {
	GroupID    string
	ArtifactID string
	Version    string
	Extension  string // "pom" or "jar"
	TargetDir  string // e.g., "./libs"
}

// BuildInstallPlan creates a list of artifacts to download based on the resolution result.
func BuildInstallPlan(result *Result, libsDir string) []InstallPlan {
	var plans []InstallPlan
	for _, winner := range result.Winners {
		plans = append(plans, InstallPlan{
			GroupID:    winner.Group,
			ArtifactID: winner.Artifact,
			Version:    winner.Version,
			Extension:  "pom",
			TargetDir:  libsDir,
		})
		plans = append(plans, InstallPlan{
			GroupID:    winner.Group,
			ArtifactID: winner.Artifact,
			Version:    winner.Version,
			Extension:  "jar",
			TargetDir:  libsDir,
		})
	}
	return plans
}

// InstallWithProgress downloads the planned artifacts with progress tracking.
// It uses the Resolver's fetcher, which is TRULY offline-first:
// 1. Checks the global cache (~/.kpm/repository) first.
// 2. Respects the --offline flag (fails immediately if not cached).
// 3. Verifies checksums for build artifacts.
// 4. Syncs the fetched bytes to the local project ./libs directory.
func (r *Resolver) InstallWithProgress(plans []InstallPlan, concurrency int, stepFunc func(detail string)) []error {
	// Enforce strict rate limiting to avoid Maven Central 429s
	downloader.SetGlobalPace(500 * time.Millisecond)

	var jobs []downloader.Job
	var errs []error

	for _, plan := range plans {
		plan := plan // capture loop variable for closure
		job := downloader.Job{
			Name: fmt.Sprintf("%s:%s:%s (%s)", plan.GroupID, plan.ArtifactID, plan.Version, plan.Extension),
			Run: func() error {
				// 1. TRUE OFFLINE-FIRST FETCH: Use the robust fetcher (handles global cache, offline flag, checksums)
				var data []byte
				var err error

				if plan.Extension == "pom" {
					data, err = r.fetcher.FetchPOM(plan.GroupID, plan.ArtifactID, plan.Version)
				} else {
					data, err = r.fetcher.FetchArtifact(plan.GroupID, plan.ArtifactID, plan.Version, "", plan.Extension)
				}

				if err != nil {
					return fmt.Errorf("failed to fetch %s:%s:%s: %w", plan.GroupID, plan.ArtifactID, plan.Version, err)
				}

				// 2. SAVE TO LOCAL PROJECT DIR: The fetcher caches globally, but KPM also 
				// maintains a local ./libs directory for the project as per the README.
				groupPath := filepath.Join(plan.TargetDir, filepath.FromSlash(plan.GroupID))
				artifactPath := filepath.Join(groupPath, plan.ArtifactID, plan.Version)
				fileName := fmt.Sprintf("%s-%s.%s", plan.ArtifactID, plan.Version, plan.Extension)
				localPath := filepath.Join(artifactPath, fileName)

				if err := os.MkdirAll(artifactPath, 0755); err != nil {
					return fmt.Errorf("failed to create local dir %s: %w", artifactPath, err)
				}

				if err := os.WriteFile(localPath, data, 0644); err != nil {
					return fmt.Errorf("failed to write to local %s: %w", localPath, err)
				}

				// 3. Report progress
				if stepFunc != nil {
					stepFunc(fmt.Sprintf("%s:%s:%s", plan.GroupID, plan.ArtifactID, plan.Version))
				}
				return nil
			},
		}
		jobs = append(jobs, job)
	}

	// Execute sequentially (concurrency=1) to respect the global downloader pace limiter
	jobErrs := downloader.RunPool(jobs, 1)

	for _, err := range jobErrs {
		errs = append(errs, err)
	}

	return errs
}