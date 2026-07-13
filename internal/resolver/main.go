// Package resolver implements KPM's dependency resolution algorithm.
//
// Unlike the original prototype (which recursively downloaded whatever a
// POM listed, with no conflict resolution and no cycle protection), this is
// a proper graph algorithm:
//
//  1. Build the root project's effective dependencyManagement, merging its
//     own <dependencyManagement> with every imported BOM (nested BOMs
//     included), which are the versions that override anything found deeper
//     in the tree.
//  2. Breadth-first traverse the dependency graph starting from the root's
//     declared dependencies, fetching each POM at most once, tracking
//     depth and exclusion sets, and rejecting cycles as soon as they'd be
//     introduced (internal/graph.AddEdge).
//  3. For each coordinate (group:artifact) with more than one candidate
//     version in the graph, resolve the conflict via Maven's "nearest
//     definition wins" rule (see conflict.go).
//  4. Return the resolved Graph plus a Result summarizing winners and any
//     conflicts that were mediated, so the CLI can print `kpm graph` /
//     `kpm why` and write the lock file.
package resolver

import (
	"fmt"

	"kpm/internal/bom"
	"kpm/internal/config"
	"kpm/internal/graph"
	"kpm/internal/parser"
	"kpm/internal/version"
)

// OnResolveStep, if set, is called once per unique coordinate right before
// its POM is fetched, so the CLI can print live "resolving X/Y" feedback
// during what can otherwise be a long, silent network-bound phase.
var OnResolveStep func(coordinate string)

// consumers of a dependency (test and provided dependencies of a library
// you depend on are that library's own concern, not yours).
var scopesToSkipTransitively = map[string]bool{"test": true, "provided": true, "system": true}

type requirement struct {
	Group, Artifact, Classifier, Type string
	VersionSpec                       string
	Scope                             string
	Optional                          bool
	Exclusions                        map[string]bool // "group:artifact"
	Depth                             int
	ParentKey                         string
}

// Result is the outcome of a full resolution pass.
type Result struct {
	Graph     *graph.Graph
	Winners   map[string]*graph.Node // coordinate key -> winning node
	Conflicts []Conflict
	Warnings  []string
}

type Resolver struct {
	fetcher  *Fetcher
	pomCache map[string]*parser.POM // "group:artifact:version" -> parsed effective POM
}

func New(f *Fetcher) *Resolver {
	return &Resolver{fetcher: f, pomCache: map[string]*parser.POM{}}
}

// Resolve computes the full transitive dependency graph for a project manifest.
func (r *Resolver) Resolve(manifest *config.Manifest) (*Result, error) {
	g := graph.New()
	result := &Result{Graph: g, Winners: map[string]*graph.Node{}}

	rootMgmt, err := r.rootManagement(manifest)
	if err != nil {
		return nil, fmt.Errorf("resolving BOMs/dependencyManagement: %w", err)
	}

	queue := r.rootRequirements(manifest)
	processed := map[string]bool{} // node keys whose own dependencies we've already expanded

	for len(queue) > 0 {
		req := queue[0]
		queue = queue[1:]

		if req.Group == "" || req.Artifact == "" {
			result.Warnings = append(result.Warnings, fmt.Sprintf("skipping malformed dependency (missing group/artifact) requested by %s", req.ParentKey))
			continue
		}

		resolvedVersion, err := r.resolveVersion(req, rootMgmt)
		if err != nil {
			return nil, fmt.Errorf("resolving version for %s:%s (requested by %s): %w", req.Group, req.Artifact, req.ParentKey, err)
		}

		node := &graph.Node{
			Coordinate: graph.Coordinate{Group: req.Group, Artifact: req.Artifact, Classifier: req.Classifier, Type: req.Type},
			Version:    resolvedVersion,
			Scope:      req.Scope,
			Optional:   req.Optional,
			Depth:      req.Depth,
		}
		if req.ParentKey != "" {
			node.Parents = []string{req.ParentKey}
		}
		node = g.AddNode(node)
		nodeKey := node.Key()

		if req.ParentKey != "" {
			if err := g.AddEdge(req.ParentKey, nodeKey); err != nil {
				return nil, err // *graph.CycleError — surfaced to the CLI's diagnostic printer
			}
		}

		if processed[nodeKey] {
			continue
		}
		processed[nodeKey] = true
		if OnResolveStep != nil {
			OnResolveStep(node.Coordinate.String() + "@" + resolvedVersion)
		}

		// Optional transitive dependencies are recorded (so `why`/`graph` can
		// show them) but not expanded further, matching Maven semantics.
		if req.Optional && req.Depth > 0 {
			continue
		}

		pom, err := r.effectivePOM(req.Group, req.Artifact, resolvedVersion)
		if err != nil {
			return nil, fmt.Errorf("loading POM for %s:%s:%s: %w", req.Group, req.Artifact, resolvedVersion, err)
		}

		for _, child := range r.childRequirements(pom, req, nodeKey) {
			queue = append(queue, child)
		}
	}

	conflicts, err := mediate(g)
	if err != nil {
		return nil, err
	}
	result.Conflicts = conflicts
	for _, n := range g.Nodes() {
		if existing, ok := result.Winners[n.Coordinate.Key()]; !ok || n.Depth < existing.Depth {
			result.Winners[n.Coordinate.Key()] = n
		}
	}
	return result, nil
}

func (r *Resolver) rootRequirements(manifest *config.Manifest) []requirement {
	var reqs []requirement
	add := func(name string, spec config.DependencySpec) {
		group, artifact := splitCoord(name)
		excl := map[string]bool{}
		for _, e := range spec.Exclusions {
			excl[e] = true
		}
		scope := spec.Scope
		if scope == "" {
			scope = "compile"
		}
		reqs = append(reqs, requirement{
			Group: group, Artifact: artifact, Classifier: spec.Classifier, Type: spec.Type,
			VersionSpec: spec.Version, Scope: scope, Optional: spec.Optional,
			Exclusions: excl, Depth: 0, ParentKey: "root",
		})
	}
	for name, spec := range manifest.Dependencies {
		add(name, spec)
	}
	for name, spec := range manifest.DevDeps {
		if spec.Scope == "" {
			spec.Scope = "test"
		}
		add(name, spec)
	}
	return reqs
}

func (r *Resolver) childRequirements(pom *parser.POM, parentReq requirement, parentNodeKey string) []requirement {
	if scopesToSkipTransitively[parentReq.Scope] {
		return nil
	}
	var out []requirement
	for _, d := range pom.Dependencies {
		if d.GroupID == "" || d.ArtifactID == "" {
			continue
		}
		if parentReq.Exclusions[d.GroupID+":"+d.ArtifactID] || parentReq.Exclusions["*:"+d.ArtifactID] {
			continue
		}
		scope := d.EffectiveScope()
		if scopesToSkipTransitively[scope] {
			continue
		}

		childExcl := map[string]bool{}
		for k := range parentReq.Exclusions {
			childExcl[k] = true
		}
		for _, e := range d.Exclusions {
			childExcl[e.GroupID+":"+e.ArtifactID] = true
		}

		out = append(out, requirement{
			Group: d.GroupID, Artifact: d.ArtifactID, Classifier: d.Classifier, Type: d.Type,
			VersionSpec: d.Version, Scope: scope, Optional: d.IsOptional(),
			Exclusions: childExcl, Depth: parentReq.Depth + 1, ParentKey: parentNodeKey,
		})
	}
	return out
}

// rootManagement merges the root project's own dependencyManagement with
// every BOM listed in package.kpm's "boms" array (document order = priority).
func (r *Resolver) rootManagement(manifest *config.Manifest) (*bom.ManagementSet, error) {
	ms := bom.NewManagementSet()
	for _, coord := range manifest.BOMs {
		group, artifact, ver := splitGAV(coord)
		if group == "" || artifact == "" || ver == "" {
			return nil, fmt.Errorf("invalid BOM coordinate %q, expected group:artifact:version", coord)
		}
		raw, err := r.fetcher.FetchPOM(group, artifact, ver)
		if err != nil {
			return nil, fmt.Errorf("fetching BOM %s: %w", coord, err)
		}
		pom, err := parser.LoadEffectivePOM(raw, manifest.Properties, r.pomFetchFn())
		if err != nil {
			return nil, fmt.Errorf("parsing BOM %s: %w", coord, err)
		}
		bomMgmt, err := bom.Resolve(pom, r.bomFetchFn())
		if err != nil {
			return nil, err
		}
		for k, v := range bomMgmt.Versions {
			if _, exists := ms.Versions[k]; !exists {
				ms.Versions[k] = v
			}
		}
	}
	return ms, nil
}

func (r *Resolver) pomFetchFn() parser.PomFetcher {
	return func(g, a, v string) ([]byte, error) { return r.fetcher.FetchPOM(g, a, v) }
}
func (r *Resolver) bomFetchFn() bom.PomFetcher {
	return func(g, a, v string) ([]byte, error) { return r.fetcher.FetchPOM(g, a, v) }
}

func (r *Resolver) effectivePOM(group, artifact, ver string) (*parser.POM, error) {
	key := group + ":" + artifact + ":" + ver
	if p, ok := r.pomCache[key]; ok {
		return p, nil
	}
	raw, err := r.fetcher.FetchPOM(group, artifact, ver)
	if err != nil {
		return nil, err
	}
	pom, err := parser.LoadEffectivePOM(raw, nil, r.pomFetchFn())
	if err != nil {
		return nil, err
	}
	r.pomCache[key] = pom
	return pom, nil
}

// resolveVersion applies Maven-like version selection precedence:
//  1. root dependencyManagement / BOM pin (always wins — that's the point of a BOM)
//  2. the version declared directly on the dependency
//  3. a Maven version RANGE, resolved against published metadata
//  4. "LATEST"/"RELEASE" keywords or an empty spec, resolved via metadata
func (r *Resolver) resolveVersion(req requirement, rootMgmt *bom.ManagementSet) (string, error) {
	if v, ok := rootMgmt.Lookup(req.Group, req.Artifact, req.Classifier, req.Type); ok {
		return v, nil
	}
	spec := req.VersionSpec

	switch {
	case spec == "" || spec == "LATEST":
		m, err := r.fetcher.GetMetadata(req.Group, req.Artifact)
		if err != nil {
			return "", fmt.Errorf("no version specified and metadata lookup failed: %w", err)
		}
		if m.Versioning.Latest == "" {
			return "", fmt.Errorf("no version specified for %s:%s and repository metadata has no 'latest'", req.Group, req.Artifact)
		}
		return m.Versioning.Latest, nil

	case spec == "RELEASE":
		m, err := r.fetcher.GetMetadata(req.Group, req.Artifact)
		if err != nil {
			return "", err
		}
		return m.Versioning.Release, nil

	case len(spec) > 0 && (spec[0] == '[' || spec[0] == '('):
		rng, err := version.ParseRange(spec)
		if err != nil {
			return "", err
		}
		m, err := r.fetcher.GetMetadata(req.Group, req.Artifact)
		if err != nil {
			return "", err
		}
		var candidates []version.Version
		for _, vs := range m.Versioning.Versions {
			candidates = append(candidates, version.Parse(vs))
		}
		best, ok := rng.PickHighest(candidates)
		if !ok {
			return "", fmt.Errorf("no published version of %s:%s satisfies range %s", req.Group, req.Artifact, spec)
		}
		return best.String(), nil

	default:
		return spec, nil
	}
}

func splitCoord(name string) (group, artifact string) {
	for i := 0; i < len(name); i++ {
		if name[i] == ':' {
			return name[:i], name[i+1:]
		}
	}
	return "", name
}

func splitGAV(coord string) (group, artifact, ver string) {
	parts := make([]string, 0, 3)
	start := 0
	for i := 0; i <= len(coord); i++ {
		if i == len(coord) || coord[i] == ':' {
			parts = append(parts, coord[start:i])
			start = i + 1
		}
	}
	if len(parts) != 3 {
		return "", "", ""
	}
	return parts[0], parts[1], parts[2]
}