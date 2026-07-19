// Package config defines KPM's project manifest ("kpm.json") — the
// human-edited file analogous to package.json/Cargo.toml/pom.xml's
// top-level project block.
package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const ManifestFile = "kpm.json"

// DependencySpec is one entry under "dependencies". A bare string is
// shorthand for an exact version; the object form allows scope, classifier,
// exclusions, and optional flag, mirroring Maven's dependency element.
type DependencySpec struct {
	Version    string   `json:"version"`
	Scope      string   `json:"scope,omitempty"` // compile (default), runtime, test, provided
	Classifier string   `json:"classifier,omitempty"`
	Type       string   `json:"type,omitempty"` // jar (default), pom, war...
	Optional   bool     `json:"optional,omitempty"`
	Exclusions []string `json:"exclusions,omitempty"` // "group:artifact" entries
}

// Repository is a custom or mirrored Maven repository entry.
type Repository struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Username string `json:"username,omitempty"`
	// Password is never stored in plaintext in kpm.json; it's resolved at
	// runtime from KPM_REPO_<ID>_PASSWORD or the OS credential store.
	Priority int  `json:"priority,omitempty"` // lower = tried first; Maven Central defaults to lowest priority
	Mirrors  bool `json:"mirrorOf,omitempty"`
}

// Manifest is the full kpm.json document.
type Manifest struct {
	Name    string `json:"name"`
	Group   string `json:"group,omitempty"` // Maven groupId this project publishes as (future @jar/publish use)
	Version string `json:"version"`
	Private bool   `json:"private,omitempty"`

	Java   int    `json:"java,omitempty"`   // target/expected JDK major version, e.g. 21 (informational for now — not yet enforced by the compiler)
	Kotlin string `json:"kotlin,omitempty"` // expected Kotlin language version, e.g. "2.2.0" (informational for now — not yet enforced)

	MainClass string `json:"mainClass,omitempty"` // default class for the "@run" builtin when no argument is given

	SourceDir string `json:"sourceDir,omitempty"` // project source root, e.g. "src" (src/main/java, src/main/kotlin live under here)
	BuildDir  string `json:"buildDir,omitempty"`  // build output root, e.g. "build" (build/classes, build/libs, etc. live under here)
	TestDir string `json:"testDir,omitempty"` // test source root, e.g. "src/test"
	Dependencies map[string]DependencySpec `json:"dependencies"`
	DevDeps      map[string]DependencySpec `json:"devDependencies,omitempty"`
	BOMs         []string                  `json:"boms,omitempty"` // "group:artifact:version" imported BOMs
	Repositories []Repository              `json:"repositories,omitempty"`
	Scripts      map[string]string         `json:"scripts,omitempty"`
	Properties   map[string]string         `json:"properties,omitempty"` // user overrides, merged over POM properties
}

func New(name string) *Manifest {
	return &Manifest{
		Name:         name,
		Version:      "0.1.0",
		SourceDir:    "src",
		BuildDir:     "build",
		TestDir:      "src/test",
		Dependencies: map[string]DependencySpec{},
		Scripts:      map[string]string{},
	}
}

func Load(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	if m.Dependencies == nil {
		m.Dependencies = map[string]DependencySpec{}
	}
	return &m, nil
}

func (m *Manifest) Save(path string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// UnmarshalJSON allows a dependency to be written either as a bare version
// string ("1.2.3") or as a full object, like npm's package.json.
func (d *DependencySpec) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		d.Version = s
		return nil
	}
	type alias DependencySpec
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*d = DependencySpec(a)
	return nil
}

// MarshalJSON writes back the compact bare-string form when no extra fields are set.
func (d DependencySpec) MarshalJSON() ([]byte, error) {
	if d.Scope == "" && d.Classifier == "" && d.Type == "" && !d.Optional && len(d.Exclusions) == 0 {
		return json.Marshal(d.Version)
	}
	type alias DependencySpec
	return json.Marshal(alias(d))
}
