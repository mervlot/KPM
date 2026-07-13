package resolver

import (
	"fmt"
	"os"
	"path/filepath"

	"kpm/internal/downloader"
	"kpm/internal/graph"
)

// InstallPlan is one artifact that needs to land on disk under libsDir.
type InstallPlan struct {
	Node    *graph.Node
	DestDir string // e.g. "./libs/<group>/<artifact>/<version>"
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
		plans = append(plans, InstallPlan{Node: n, DestDir: dest})
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
				name := p.Node.Artifact + "-" + p.Node.Version
				if p.Node.Classifier != "" {
					name += "-" + p.Node.Classifier
				}
				dest := filepath.Join(p.DestDir, name+"."+ext)
				werr := os.WriteFile(dest, data, 0o644)
				if onStep != nil {
					onStep(p.Node.Coordinate.String())
				}
				return werr
			},
		})
	}
	return downloader.RunPool(jobs, concurrency)
}