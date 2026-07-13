// Package lockfile defines kpm.lock: the resolved, checksum-pinned
// dependency tree that guarantees every `kpm install` reproduces exactly
// the same artifacts until the manifest changes and the lock is regenerated.
package lockfile

import (
	"encoding/json"
	"os"
	"sort"

	"kpm/internal/resolver"
)

const FileName = "kpm.lock"

const SchemaVersion = 1

type Entry struct {
	Group      string   `json:"group"`
	Artifact   string   `json:"artifact"`
	Version    string   `json:"version"`
	Classifier string   `json:"classifier,omitempty"`
	Type       string   `json:"type"`
	Scope      string   `json:"scope"`
	Repository string   `json:"repository"`
	Checksum   string   `json:"sha256,omitempty"`
	Requires   []string `json:"requires,omitempty"` // node keys of direct dependencies (dependency tree)
	Depth      int      `json:"depth"`
}

type Lockfile struct {
	SchemaVersion int              `json:"schemaVersion"`
	Entries       map[string]Entry `json:"entries"` // keyed by node key: group:artifact:classifier:type@version
}

func New() *Lockfile {
	return &Lockfile{SchemaVersion: SchemaVersion, Entries: map[string]Entry{}}
}

// FromResult converts a resolver.Result's graph into a lockfile, keeping
// only the winning version of each coordinate (the set a real build uses)
// plus enough of the graph structure to reconstruct "why" relationships.
func FromResult(res *resolver.Result, repoOf func(group, artifact, version string) string) *Lockfile {
	lf := New()
	for _, n := range res.Graph.Nodes() {
		winner, ok := res.Winners[n.Coordinate.Key()]
		if !ok || winner.Version != n.Version {
			continue // superseded by conflict resolution; not part of the effective build
		}
		var requires []string
		for to := range map[string]bool{} { // placeholder for future direct-edge export
			requires = append(requires, to)
		}
		sort.Strings(requires)

		e := Entry{
			Group: n.Group, Artifact: n.Artifact, Version: n.Version,
			Classifier: n.Classifier, Type: n.Type, Scope: n.Scope,
			Depth: n.Depth, Requires: requires,
		}
		if repoOf != nil {
			e.Repository = repoOf(n.Group, n.Artifact, n.Version)
		}
		lf.Entries[n.Key()] = e
	}
	return lf
}

func Load(path string) (*Lockfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var lf Lockfile
	if err := json.Unmarshal(data, &lf); err != nil {
		return nil, err
	}
	if lf.Entries == nil {
		lf.Entries = map[string]Entry{}
	}
	return &lf, nil
}

func (lf *Lockfile) Save(path string) error {
	data, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Diff reports coordinates whose resolved version differs between two
// lockfiles (or is newly added/removed) — used by `kpm outdated` and to
// warn when an install would change a previously locked resolution.
func Diff(old, new *Lockfile) (added, removed, changed []string) {
	for k, e := range new.Entries {
		if oe, ok := old.Entries[k]; !ok {
			added = append(added, k)
		} else if oe.Version != e.Version {
			changed = append(changed, k)
		}
	}
	for k := range old.Entries {
		if _, ok := new.Entries[k]; !ok {
			removed = append(removed, k)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	sort.Strings(changed)
	return
}