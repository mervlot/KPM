// Package repository models Maven repositories: Central, custom, mirrors,
// and the priority/fallback order KPM tries them in when resolving an
// artifact.
package repository

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"kpm/internal/config"
)

const CentralURL = "https://repo1.maven.org/maven2"

// Repository is one Maven-layout HTTP repository KPM can fetch from.
type Repository struct {
	ID       string
	URL      string
	Priority int // lower tried first
	Username string
	Password string // resolved from env/credential store, never persisted
	MirrorOf string // if set, this repo is a mirror substituting for the named repo/"*"
}

// ArtifactURL builds the URL for group:artifact:version's primary artifact
// (or POM, when ext == "pom").
func (r Repository) ArtifactURL(group, artifact, version, classifier, ext string) string {
	gp := strings.ReplaceAll(group, ".", "/")
	name := artifact + "-" + version
	if classifier != "" {
		name += "-" + classifier
	}
	return fmt.Sprintf("%s/%s/%s/%s/%s.%s", strings.TrimRight(r.URL, "/"), gp, artifact, version, name, ext)
}

func (r Repository) MetadataURL(group, artifact string) string {
	gp := strings.ReplaceAll(group, ".", "/")
	return fmt.Sprintf("%s/%s/%s/maven-metadata.xml", strings.TrimRight(r.URL, "/"), gp, artifact)
}

// Set is an ordered, mirror-resolved list of repositories to try in order.
type Set struct {
	repos []Repository
}

// BuildSet assembles the effective repository list: Maven Central plus any
// custom repositories from package.kpm, sorted by priority, with mirrors
// substituted in for the repositories they declare mirrorOf for.
// Credentials are resolved from KPM_REPO_<ID>_USERNAME / _PASSWORD env vars
// (uppercased, non-alnum replaced with '_') rather than stored in
// package.kpm, so secrets never end up in version control.
func BuildSet(manifest *config.Manifest) *Set {
	repos := []Repository{{ID: "central", URL: CentralURL, Priority: 1000}}

	for _, r := range manifest.Repositories {
		rep := Repository{ID: r.ID, URL: r.URL, Priority: r.Priority, Username: r.Username}
		if rep.Priority == 0 {
			rep.Priority = 100 // custom repos default ahead of Central unless specified
		}
		rep.Password = os.Getenv(envKey(r.ID, "PASSWORD"))
		if rep.Username == "" {
			rep.Username = os.Getenv(envKey(r.ID, "USERNAME"))
		}
		if r.Mirrors {
			rep.MirrorOf = "*"
		}
		repos = append(repos, rep)
	}

	// Mirrors replace what they mirror rather than sitting alongside it.
	var mirrors, direct []Repository
	for _, r := range repos {
		if r.MirrorOf != "" {
			mirrors = append(mirrors, r)
		} else {
			direct = append(direct, r)
		}
	}
	final := direct
	if len(mirrors) > 0 {
		final = nil
		for _, r := range direct {
			mirrored := false
			for _, m := range mirrors {
				if m.MirrorOf == "*" || m.MirrorOf == r.ID {
					mirrored = true
				}
			}
			if !mirrored {
				final = append(final, r)
			}
		}
		final = append(final, mirrors...)
	}

	sort.Slice(final, func(i, j int) bool { return final[i].Priority < final[j].Priority })
	return &Set{repos: final}
}

func (s *Set) All() []Repository { return s.repos }

func envKey(id, suffix string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(id) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return "KPM_REPO_" + b.String() + "_" + suffix
}