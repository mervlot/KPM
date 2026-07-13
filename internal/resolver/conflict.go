package resolver

import (
	"sort"

	"kpm/internal/graph"
)

// Conflict records that more than one version of a coordinate was requested
// somewhere in the graph, and which one KPM chose to resolve it to.
type Conflict struct {
	Coordinate string
	Chosen     string
	Candidates []CandidateVersion
}

type CandidateVersion struct {
	Version string
	Depth   int
	Via     string // parent node key that requested this version
}

// mediate implements Maven's default conflict resolution strategy: the
// version found at the SHALLOWEST depth wins ("nearest definition"); ties
// at equal depth are broken deterministically by version string so repeated
// runs of the same input always produce the same result.
//
// This does not mutate the graph (multiple candidate nodes remain, e.g. for
// `kpm why`); it only decides which single version a build should actually
// use for each coordinate, recorded in Result.Winners and the lock file.
func mediate(g *graph.Graph) ([]Conflict, error) {
	byCoord := map[string][]*graph.Node{}
	for _, n := range g.Nodes() {
		byCoord[n.Coordinate.Key()] = append(byCoord[n.Coordinate.Key()], n)
	}

	var conflicts []Conflict
	for coordKey, nodes := range byCoord {
		if len(nodes) <= 1 {
			continue
		}
		sort.Slice(nodes, func(i, j int) bool {
			if nodes[i].Depth != nodes[j].Depth {
				return nodes[i].Depth < nodes[j].Depth
			}
			return nodes[i].Version < nodes[j].Version
		})
		winner := nodes[0]

		c := Conflict{Coordinate: coordKey, Chosen: winner.Version}
		for _, n := range nodes {
			via := "root"
			if len(n.Parents) > 0 {
				via = n.Parents[0]
			}
			c.Candidates = append(c.Candidates, CandidateVersion{Version: n.Version, Depth: n.Depth, Via: via})
		}
		conflicts = append(conflicts, c)
	}

	sort.Slice(conflicts, func(i, j int) bool { return conflicts[i].Coordinate < conflicts[j].Coordinate })
	return conflicts, nil
}