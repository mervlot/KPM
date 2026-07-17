// Package classpath turns a resolved kpm.lock into a deterministic,
// deduplicated Java classpath, reading only what `kpm install` already put
// on disk. It NEVER re-resolves dependencies and NEVER downloads anything —
// if a jar the lock file expects isn't on disk, that's reported as
// MissingJarError telling the person to run `kpm install`, not silently
// fetched.
package classpath

import (
	"fmt"
	"os"
	"sort"

	"kpm/internal/lockfile"
	"kpm/internal/resolver"
)

// Scope selects which dependency scopes belong on the classpath, mirroring
// Maven's scope semantics:
//   - Compile: compile + provided (provided is compile-time-only — e.g. a
//     servlet container supplies it at runtime, so code can reference it
//     while compiling without KPM needing to have downloaded a runtime copy)
//   - Runtime: compile + runtime (provided is deliberately excluded — it's
//     supplied by the environment, not bundled)
//
// Neither scope includes "test" or "system": test deps have no place on a
// production compile/runtime classpath, and "system" (a local, unmanaged
// path) isn't implemented — see the "system scope" note in Build.
type Scope int

const (
	Compile Scope = iota
	Runtime
)

// Classpath is an ordered, deduplicated list of jar paths.
type Classpath struct {
	Entries []string
}

// String joins entries with the OS-appropriate classpath separator
// (":" on Unix, ";" on Windows), ready to hand to javac/kotlinc's -cp flag.
func (c *Classpath) String() string {
	sep := string(os.PathListSeparator)
	out := ""
	for i, e := range c.Entries {
		if i > 0 {
			out += sep
		}
		out += e
	}
	return out
}

// MissingLockfileError means kpm.lock doesn't exist yet — resolution has
// never been run, so there's nothing to build a classpath from.
type MissingLockfileError struct{ Path string }

func (e *MissingLockfileError) Error() string {
	return fmt.Sprintf("no %s found — run `kpm install` first so there's a resolved dependency set to compile against", e.Path)
}

// MissingJarError means kpm.lock references a coordinate whose jar isn't
// actually on disk under InstalledJarsDir — the lock and the local cache
// have drifted apart (e.g. `kpm cache clean` ran after the lock was
// written, or the lock file was hand-edited).
type MissingJarError struct {
	Coordinate string
	Path       string
}

func (e *MissingJarError) Error() string {
	return fmt.Sprintf("dependency %s is in kpm.lock but its jar is missing at %s — run `kpm install` to re-download it", e.Coordinate, e.Path)
}

// ConflictError means kpm.lock somehow contains two different versions of
// the same group:artifact (this shouldn't happen in practice — the
// resolver's conflict mediation is supposed to guarantee exactly one
// winner per coordinate — but the classpath layer checks anyway rather
// than silently picking one, since a build compiled against the wrong
// version of a library fails in far more confusing ways than a clear error
// up front).
type ConflictError struct {
	Coordinate         string
	VersionA, VersionB string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf(
		"classpath conflict: kpm.lock lists both %s:%s and %s:%s — this indicates a corrupted or hand-edited kpm.lock; run `kpm sync` to regenerate it",
		e.Coordinate, e.VersionA, e.Coordinate, e.VersionB,
	)
}

// Build reads lockPath (normally "kpm.lock") and returns the classpath for
// the given scope, resolving each entry's jar path under installedJarsDir
// via resolver.ArtifactPath — the exact same path convention `kpm install`
// used to put it there, so this never has to guess or duplicate that logic.
func Build(lockPath, installedJarsDir string, scope Scope) (*Classpath, error) {
	lf, err := lockfile.Load(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &MissingLockfileError{Path: lockPath}
		}
		return nil, err
	}

	// Deterministic order: map iteration order is random in Go, and a
	// classpath whose entry order changes between otherwise-identical runs
	// makes compiler errors (which can be order-sensitive when duplicate
	// classes exist across jars) non-reproducible. Sort by node key before
	// doing anything else.
	keys := make([]string, 0, len(lf.Entries))
	for k := range lf.Entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	seenGA := map[string]lockfile.Entry{} // "group:artifact" -> the entry already added
	var cp Classpath

	for _, k := range keys {
		e := lf.Entries[k]

		if e.Type == "pom" {
			continue // no jar to add
		}
		if e.Classifier != "" {
			continue // sources/javadoc/etc. classifier jars never belong on a compile/runtime classpath
		}
		if !scopeAllowed(scope, e.Scope) {
			continue
		}

		ga := e.Group + ":" + e.Artifact
		if existing, dup := seenGA[ga]; dup {
			if existing.Version != e.Version {
				return nil, &ConflictError{Coordinate: ga, VersionA: existing.Version, VersionB: e.Version}
			}
			continue // exact duplicate (e.g. reachable via two scopes); never add the same jar twice
		}
		seenGA[ga] = e

		ext := e.Type
		if ext == "" {
			ext = "jar"
		}
		path := resolver.ArtifactPath(installedJarsDir, e.Group, e.Artifact, e.Version, e.Classifier, ext)
		if _, statErr := os.Stat(path); statErr != nil {
			return nil, &MissingJarError{Coordinate: e.Group + ":" + e.Artifact + ":" + e.Version, Path: path}
		}
		cp.Entries = append(cp.Entries, path)
	}

	return &cp, nil
}

// scopeAllowed implements the Maven-ish scope table described on the Scope
// type. An empty Scope on the lockfile entry defaults to "compile" (same
// default the resolver itself uses).
func scopeAllowed(want Scope, entryScope string) bool {
	if entryScope == "" {
		entryScope = "compile"
	}
	switch entryScope {
	case "test", "system":
		return false // test deps never belong on a production classpath; system scope isn't implemented
	case "provided":
		return want == Compile
	case "runtime":
		return want == Runtime
	case "compile":
		return true
	default:
		return false
	}
}