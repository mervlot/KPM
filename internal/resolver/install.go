package resolver

import (
	"fmt"
	"os"
	"path/filepath"

	"kpm/internal/downloader"
	"kpm/internal/graph"
)

// InstallPlan is one artifact that needs to land on disk under libsRoot.
type InstallPlan struct {
	Node     *graph.Node
	LibsRoot string // e.g. "./libs" — passed through so ArtifactPath can be recomputed without guessing
	DestDir  string // e.g. "./libs/<group>/<artifact>/<version>/"
}

// ArtifactPath computes the deterministic on-disk path an installed
// artifact lives at under libsRoot: libsRoot/group/artifact/version/
// artifact-version[-classifier].ext. This is the single source of truth
// for that layout — both the installer (below) and internal/classpath
// (which needs to find already-downloaded jars without re-resolving or
// re-downloading anything) call this instead of each re-deriving the
// naming convention, which would risk the two silently drifting apart.
func ArtifactPath(libsRoot, group, artifact, version, classifier, ext string) string {
	if ext == "" {
		ext = "jar"
	}
	name := artifact + "-" + version
	if classifier != "" {
		name += "-" + classifier
	}
	return filepath.Join(libsRoot, group, artifact, version, name+"."+ext)
}

// BuildInstallPlan returns the winning node for every coordinate in the
// result (skipping "pom"-only packaging, which has nothing to download,
// and test-scope nodes, which never need to be installed for a build).
func BuildInstallPlan(res *Result, libsRoot string) []InstallPlan {
	var plans []InstallPlan
	for _, n := range res.Winners {
		if n.Type == "pom" {
			continue
		}
		if n.Scope == "test" {
			continue
		}
		dest := filepath.Join(libsRoot, n.Group, n.Artifact, n.Version)
		plans = append(plans, InstallPlan{Node: n, LibsRoot: libsRoot, DestDir: dest})
	}
	return plans
}

// Install downloads every artifact in the plan in parallel (bounded by
// concurrency), verifying checksums via the Fetcher, and reports every
// failure rather than stopping at the first so a single bad artifact
// doesn't obscure unrelated ones.
func (r *Resolver) Install(plans []InstallPlan, concurrency int) []error {
	return r.InstallWithProgress(plans, concurrency, nil)
}

// InstallWithProgress is Install, additionally invoking onStep(detail) once
// per completed artifact (success or failure) so callers can drive a live
// progress line instead of the download phase going silent.
func (r *Resolver) InstallWithProgress(plans []InstallPlan, concurrency int, onStep func(detail string)) []error {
	jobs := make([]downloader.Job, 0, len(plans))
	for _, p := range plans {
		p := p
		ext := p.Node.Type
		if ext == "" {
			ext = "jar"
		}
		jobs = append(jobs, downloader.Job{
			Name: fmt.Sprintf("%s:%s:%s", p.Node.Group, p.Node.Artifact, p.Node.Version),
			Run: func() error {
				data, err := r.fetcher.FetchArtifact(p.Node.Group, p.Node.Artifact, p.Node.Version, p.Node.Classifier, ext)
				if err != nil {
					if onStep != nil {
						onStep(p.Node.Coordinate.String() + " (failed)")
					}
					return err
				}
				if err := os.MkdirAll(p.DestDir, 0o755); err != nil {
					return err
				}
				dest := ArtifactPath(p.LibsRoot, p.Node.Group, p.Node.Artifact, p.Node.Version, p.Node.Classifier, ext)

				// CRITICAL FIX: ATOMIC WRITE. Write to temp file first, then rename.
				// This prevents corrupt partial downloads from being left on disk
				// and mistaken for valid cache entries or build artifacts.
				tmpDest := dest + ".tmp"
				if werr := os.WriteFile(tmpDest, data, 0o644); werr != nil {
					return werr
				}
				if werr := os.Rename(tmpDest, dest); werr != nil {
					return werr
				}

				if onStep != nil {
					onStep(p.Node.Coordinate.String())
				}
				return nil
			},
		})
	}
	return downloader.RunPool(jobs, concurrency)
}