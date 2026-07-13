package version

import (
	"fmt"
	"strings"
)

// Range represents a Maven version range, e.g. "[1.0,2.0)", "[1.5,)", "(,2.0]", "[1.0]".
// A Range with Exact set represents a soft "recommended" version (a bare
// version string like "1.2.3", which Maven treats as a suggestion, not a hard
// constraint) OR a hard single-version range "[1.2.3]".
type Range struct {
	raw      string
	Exact    *Version // set for bare recommended versions and [x] hard-pins
	LowIncl  bool
	Low      *Version
	High     *Version
	HighIncl bool
	isRange  bool
}

// ParseRange parses a Maven version or version-range specifier.
func ParseRange(spec string) (Range, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return Range{}, fmt.Errorf("empty version spec")
	}
	if !strings.HasPrefix(spec, "[") && !strings.HasPrefix(spec, "(") {
		v := Parse(spec)
		return Range{raw: spec, Exact: &v}, nil
	}

	if len(spec) < 3 {
		return Range{}, fmt.Errorf("invalid version range: %q", spec)
	}
	lowIncl := spec[0] == '['
	highIncl := spec[len(spec)-1] == ']'
	inner := spec[1 : len(spec)-1]

	parts := strings.SplitN(inner, ",", 2)
	r := Range{raw: spec, isRange: true, LowIncl: lowIncl, HighIncl: highIncl}

	if len(parts) == 1 {
		// Single hard-pinned version like [1.2.3]
		v := Parse(parts[0])
		r.Exact = &v
		r.isRange = false
		return r, nil
	}

	if strings.TrimSpace(parts[0]) != "" {
		v := Parse(strings.TrimSpace(parts[0]))
		r.Low = &v
	}
	if strings.TrimSpace(parts[1]) != "" {
		v := Parse(strings.TrimSpace(parts[1]))
		r.High = &v
	}
	return r, nil
}

// IsSoft reports whether this is a bare recommended version (not a hard range).
func (r Range) IsSoft() bool { return !r.isRange && r.Exact != nil && !strings.HasPrefix(r.raw, "[") }

// Matches reports whether v satisfies the range constraint.
func (r Range) Matches(v Version) bool {
	if r.Exact != nil && !r.isRange {
		return v.Equal(*r.Exact)
	}
	if r.Low != nil {
		c := v.Compare(*r.Low)
		if r.LowIncl {
			if c < 0 {
				return false
			}
		} else if c <= 0 {
			return false
		}
	}
	if r.High != nil {
		c := v.Compare(*r.High)
		if r.HighIncl {
			if c > 0 {
				return false
			}
		} else if c >= 0 {
			return false
		}
	}
	return true
}

func (r Range) String() string { return r.raw }

// PickHighest returns the highest candidate satisfying the range, or false
// if none match.
func (r Range) PickHighest(candidates []Version) (Version, bool) {
	var best *Version
	for i := range candidates {
		c := candidates[i]
		if !r.Matches(c) {
			continue
		}
		if best == nil || c.GreaterThan(*best) {
			cc := c
			best = &cc
		}
	}
	if best == nil {
		return Version{}, false
	}
	return *best, true
}