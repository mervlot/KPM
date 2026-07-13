// Package bom implements Maven's <dependencyManagement><dependencies>
// <dependency><scope>import</scope> mechanism — Bills of Materials such as
// the Spring, Kotlin, Jakarta, Android and Firebase BOMs — including BOMs
// that themselves import further BOMs.'
package bom

import (
	"fmt"

	"kpm/internal/parser"
)

// ManagedKey identifies a managed dependency by group:artifact:classifier:type.
type ManagedKey struct {
	Group, Artifact, Classifier, Type string
}

func keyFor(d parser.RawDependency) ManagedKey {
	t := d.Type
	if t == "" {
		t = "jar"
	}
	return ManagedKey{d.GroupID, d.ArtifactID, d.Classifier, t}
}

// ManagementSet is the fully-merged set of version/scope pins in effect,
// following Maven's "nearest declaration wins" rule: entries declared
// directly in the project's own dependencyManagement take priority over
// anything pulled in via a BOM import, and among BOM imports, the first
// one listed wins (matching Maven's document-order rule).
type ManagementSet struct {
	Versions map[ManagedKey]string
	Scopes   map[ManagedKey]string
}

func NewManagementSet() *ManagementSet {
	return &ManagementSet{Versions: map[ManagedKey]string{}, Scopes: map[ManagedKey]string{}}
}

func (m *ManagementSet) put(d parser.RawDependency, overwrite bool) {
	k := keyFor(d)
	if _, exists := m.Versions[k]; exists && !overwrite {
		return
	}
	if d.Version != "" {
		m.Versions[k] = d.Version
	}
	if d.Scope != "" && d.Scope != "import" {
		m.Scopes[k] = d.Scope
	}
}

func (m *ManagementSet) Lookup(group, artifact, classifier, typ string) (version string, ok bool) {
	if typ == "" {
		typ = "jar"
	}
	v, ok := m.Versions[ManagedKey{group, artifact, classifier, typ}]
	return v, ok
}

// PomFetcher retrieves a POM's bytes for group:artifact:version (same shape
// as parser.PomFetcher, kept separate to avoid a hard coupling here).
type PomFetcher func(group, artifact, version string) ([]byte, error)

// Resolve builds the ManagementSet for a POM's own dependencyManagement
// block: direct entries are recorded first (highest priority), then any
// <scope>import</scope> BOM entries are fetched and merged in document
// order, recursing into BOMs-of-BOMs. `visited` guards against import
// cycles between BOMs.
func Resolve(pom *parser.POM, fetch PomFetcher) (*ManagementSet, error) {
	return resolve(pom, fetch, map[string]bool{})
}

func resolve(pom *parser.POM, fetch PomFetcher, visited map[string]bool) (*ManagementSet, error) {
	ms := NewManagementSet()

	// Pass 1: direct (non-import) entries take precedence over anything a BOM brings in.
	var imports []parser.RawDependency
	for _, d := range pom.DependencyManagement.Dependencies {
		if d.Scope == "import" {
			imports = append(imports, d)
			continue
		}
		ms.put(d, true)
	}

	// Pass 2: imported BOMs, in document order; first import wins on conflicts,
	// and none of them override the project's own direct entries from pass 1.
	for _, imp := range imports {
		key := imp.GroupID + ":" + imp.ArtifactID + ":" + imp.Version
		if visited[key] {
			continue // nested BOM cycle; skip rather than fail the whole resolution
		}
		visited[key] = true

		if fetch == nil {
			continue
		}
		raw, err := fetch(imp.GroupID, imp.ArtifactID, imp.Version)
		if err != nil {
			return nil, fmt.Errorf("importing BOM %s: %w", key, err)
		}
		bomPom, err := parser.LoadEffectivePOM(raw, nil, func(g, a, v string) ([]byte, error) { return fetch(g, a, v) })
		if err != nil {
			return nil, fmt.Errorf("parsing BOM %s: %w", key, err)
		}
		nested, err := resolve(bomPom, fetch, visited)
		if err != nil {
			return nil, err
		}
		for k, v := range nested.Versions {
			if _, exists := ms.Versions[k]; !exists {
				ms.Versions[k] = v
			}
		}
		for k, v := range nested.Scopes {
			if _, exists := ms.Scopes[k]; !exists {
				ms.Scopes[k] = v
			}
		}
	}

	return ms, nil
}

// Merge layers `overlay` on top of `base` without mutating either,
// preferring `base` entries on conflict — used to combine the current
// project's own management set with each dependency's inherited one
// (nearest-definition-wins across the dependency tree, per Maven rules).
func Merge(base, overlay *ManagementSet) *ManagementSet {
	out := NewManagementSet()
	for k, v := range overlay.Versions {
		out.Versions[k] = v
	}
	for k, v := range overlay.Scopes {
		out.Scopes[k] = v
	}
	for k, v := range base.Versions {
		out.Versions[k] = v
	}
	for k, v := range base.Scopes {
		out.Scopes[k] = v
	}
	return out
}