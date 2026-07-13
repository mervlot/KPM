// Package graph implements the dependency graph used by the resolver:
// nodes are GAV (group:artifact:version) coordinates plus scope/optional
// metadata, edges are "depends on" relationships annotated with the
// requested version and the path (parent chain) that introduced them, so
// conflict resolution and `kpm why` can explain provenance.
package graph

import (
	"fmt"
	"sort"
	"strings"
)

// Coordinate identifies an artifact without a version — the resolution unit
// Maven conflict rules operate on (only one version of group:artifact wins).
type Coordinate struct {
	Group      string
	Artifact   string
	Classifier string // "" = none
	Type       string // "jar", "pom", "war", ... ("" defaults to "jar")
}

func (c Coordinate) Key() string {
	t := c.Type
	if t == "" {
		t = "jar"
	}
	return fmt.Sprintf("%s:%s:%s:%s", c.Group, c.Artifact, c.Classifier, t)
}

func (c Coordinate) String() string {
	if c.Classifier != "" {
		return fmt.Sprintf("%s:%s:%s", c.Group, c.Artifact, c.Classifier)
	}
	return fmt.Sprintf("%s:%s", c.Group, c.Artifact)
}

// Node is one resolved-or-candidate artifact in the graph.
type Node struct {
	Coordinate
	Version  string
	Scope    string // compile, runtime, test, provided, system
	Optional bool
	Depth    int      // shortest known distance from a root, used for "nearest wins"
	Parents  []string // node keys (Coordinate.Key()+"@"+version) that requested this node, for `why`
}

func (n *Node) Key() string { return n.Coordinate.Key() + "@" + n.Version }

// Graph is a directed graph of dependency requirements. It may contain
// multiple candidate Versions per Coordinate simultaneously during
// resolution; conflict resolution later collapses each Coordinate down to
// exactly one winning Node.
type Graph struct {
	nodes map[string]*Node           // nodeKey -> node
	edges map[string]map[string]bool // fromNodeKey -> set of toNodeKey
	rev   map[string]map[string]bool // toNodeKey -> set of fromNodeKey (for `why`)
}

func New() *Graph {
	return &Graph{
		nodes: make(map[string]*Node),
		edges: make(map[string]map[string]bool),
		rev:   make(map[string]map[string]bool),
	}
}

// AddNode inserts (or returns the existing) node for this exact coordinate+version.
func (g *Graph) AddNode(n *Node) *Node {
	k := n.Key()
	if existing, ok := g.nodes[k]; ok {
		if n.Depth < existing.Depth {
			existing.Depth = n.Depth
		}
		existing.Parents = append(existing.Parents, n.Parents...)
		return existing
	}
	g.nodes[k] = n
	return n
}

// AddEdge records that `from` depends on `to`. Both nodes must already exist.
func (g *Graph) AddEdge(fromKey, toKey string) error {
	if fromKey == toKey {
		return fmt.Errorf("self-dependency: %s", fromKey)
	}
	if g.edges[fromKey] == nil {
		g.edges[fromKey] = make(map[string]bool)
	}
	g.edges[fromKey][toKey] = true
	if g.rev[toKey] == nil {
		g.rev[toKey] = make(map[string]bool)
	}
	g.rev[toKey][fromKey] = true

	if cyclePath, cyclic := g.detectCycleFrom(toKey, fromKey); cyclic {
		return &CycleError{Path: cyclePath}
	}
	return nil
}

func (g *Graph) Nodes() map[string]*Node { return g.nodes }

func (g *Graph) NodesByCoordinate(coordKey string) []*Node {
	var out []*Node
	for _, n := range g.nodes {
		if n.Coordinate.Key() == coordKey {
			out = append(out, n)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out
}

// CycleError describes a dependency cycle in human-readable "a -> b -> c -> a" form.
type CycleError struct{ Path []string }

func (e *CycleError) Error() string {
	return "circular dependency detected: " + strings.Join(e.Path, " -> ")
}

// detectCycleFrom does a DFS starting at `start` looking for a path back to `target`.
// Called after adding edge target->start conceptually reversed: we check whether
// `to` can already reach `from`, which combined with the new from->to edge forms a cycle.
func (g *Graph) detectCycleFrom(start, target string) ([]string, bool) {
	visited := make(map[string]bool)
	var path []string
	var dfs func(cur string) bool
	dfs = func(cur string) bool {
		if cur == target {
			path = append(path, cur)
			return true
		}
		if visited[cur] {
			return false
		}
		visited[cur] = true
		path = append(path, cur)
		for next := range g.edges[cur] {
			if dfs(next) {
				return true
			}
		}
		path = path[:len(path)-1]
		return false
	}
	if dfs(start) {
		full := append([]string{target}, path...)
		return full, true
	}
	return nil, false
}

// TopoSort returns nodes in dependency order (a node appears after all of
// its dependencies). Errors if the graph has a cycle (shouldn't happen since
// AddEdge rejects cycles eagerly, but kept as a defense-in-depth check).
func (g *Graph) TopoSort() ([]*Node, error) {
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := make(map[string]int)
	var order []*Node
	var visit func(k string) error
	visit = func(k string) error {
		switch color[k] {
		case black:
			return nil
		case gray:
			return &CycleError{Path: []string{k}}
		}
		color[k] = gray
		for next := range g.edges[k] {
			if err := visit(next); err != nil {
				return err
			}
		}
		color[k] = black
		if n, ok := g.nodes[k]; ok {
			order = append(order, n)
		}
		return nil
	}
	keys := make([]string, 0, len(g.nodes))
	for k := range g.nodes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if err := visit(k); err != nil {
			return nil, err
		}
	}
	return order, nil
}

// Why returns every root-to-node path that pulled `nodeKey` into the graph,
// used to implement `kpm why <artifact>`.
func (g *Graph) Why(nodeKey string) [][]string {
	var paths [][]string
	var walk func(k string, trail []string, seen map[string]bool)
	walk = func(k string, trail []string, seen map[string]bool) {
		trail = append([]string{k}, trail...)
		parents := g.rev[k]
		if len(parents) == 0 {
			paths = append(paths, trail)
			return
		}
		for p := range parents {
			if seen[p] {
				continue
			}
			seen2 := make(map[string]bool, len(seen)+1)
			for s := range seen {
				seen2[s] = true
			}
			seen2[p] = true
			walk(p, trail, seen2)
		}
	}
	walk(nodeKey, nil, map[string]bool{nodeKey: true})
	return paths
}