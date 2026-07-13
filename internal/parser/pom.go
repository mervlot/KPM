package parser

import (
	"encoding/xml"
	"fmt"
)

// PomFetcher retrieves the raw bytes of a POM for group:artifact:version.
// Implemented by internal/repository so this package stays decoupled from
// network/cache concerns and is easy to unit test with a fake.
type PomFetcher func(group, artifact, version string) ([]byte, error)

// ParsePOM parses raw POM bytes without resolving parent inheritance.
func ParsePOM(data []byte) (*POM, error) {
	var p POM
	if err := xml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("invalid POM XML: %w", err)
	}
	if p.Packaging == "" {
		p.Packaging = "jar"
	}
	return &p, nil
}

// LoadEffectivePOM parses a POM and resolves the full parent chain (Maven
// allows arbitrarily deep parent POMs), merging properties, dependencies and
// dependencyManagement from ancestors before applying property interpolation.
// Ancestor-defined values are overridden by descendant values, matching Maven.
func LoadEffectivePOM(data []byte, extraProps map[string]string, fetch PomFetcher) (*POM, error) {
	pom, err := ParsePOM(data)
	if err != nil {
		return nil, err
	}

	chain := []*POM{pom}
	seen := map[string]bool{pom.GroupID + ":" + pom.ArtifactID + ":" + pom.Version: true}
	cur := pom
	for cur.Parent != nil {
		pr := cur.Parent
		key := pr.GroupID + ":" + pr.ArtifactID + ":" + pr.Version
		if seen[key] {
			return nil, fmt.Errorf("cyclic parent POM chain detected at %s", key)
		}
		seen[key] = true

		if fetch == nil {
			break
		}
		raw, ferr := fetch(pr.GroupID, pr.ArtifactID, pr.Version)
		if ferr != nil {
			return nil, fmt.Errorf("fetching parent POM %s: %w", key, ferr)
		}
		parent, perr := ParsePOM(raw)
		if perr != nil {
			return nil, fmt.Errorf("parsing parent POM %s: %w", key, perr)
		}
		chain = append(chain, parent)
		cur = parent
	}

	// Merge root-to-leaf: start from the most distant ancestor, apply each
	// descendant's overrides on top, so the original pom's own values win.
	merged := &POM{Properties: Properties{Entries: map[string]string{}}}
	dmSeen := map[string]bool{}
	depManaged := []RawDependency{}
	for i := len(chain) - 1; i >= 0; i-- {
		p := chain[i]
		for k, v := range p.Properties.Entries {
			merged.Properties.Entries[k] = v
		}
		for _, d := range p.DependencyManagement.Dependencies {
			key := d.GroupID + ":" + d.ArtifactID + ":" + d.Classifier + ":" + d.Type
			if dmSeen[key] {
				// closer (more-derived) definition already recorded; skip ancestor's
				continue
			}
			dmSeen[key] = true
			depManaged = append(depManaged, d)
		}
	}
	merged.GroupID = pom.GroupID
	if merged.GroupID == "" && pom.Parent != nil {
		merged.GroupID = pom.Parent.GroupID
	}
	merged.ArtifactID = pom.ArtifactID
	merged.Version = pom.Version
	if merged.Version == "" && pom.Parent != nil {
		merged.Version = pom.Parent.Version
	}
	merged.Packaging = pom.Packaging
	merged.Dependencies = pom.Dependencies
	merged.DependencyManagement.Dependencies = depManaged
	merged.Repositories = pom.Repositories

	interp := NewInterpolator(merged, extraProps)
	interp.ResolvePOM(merged)

	return merged, nil
}

// IsOptional parses the Maven "optional" string field ("true"/"false", default false).
func (d RawDependency) IsOptional() bool { return d.Optional == "true" }

// EffectiveScope returns the dependency's scope, defaulting to "compile" per Maven rules.
func (d RawDependency) EffectiveScope() string {
	if d.Scope == "" {
		return "compile"
	}
	return d.Scope
}