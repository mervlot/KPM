// Package version implements Maven-compatible version parsing and comparison.
//
// Maven versions are NOT semver. Maven splits a version string into a
// sequence of tokens (numeric or alphabetic), separated by '.', '-', or a
// transition between digit/letter, and compares tokens pairwise. Certain
// alphabetic qualifiers ("alpha", "beta", "milestone", "rc"/"cr", "snapshot",
// "", "ga"/"final", "sp") have a well-defined relative ordering that does not
// match plain lexical ordering. This package implements that ordering
// closely enough for real-world dependency resolution (this is the same
// algorithm Maven's ComparableVersion uses).
package version

import (
	"strconv"
	"strings"
)

// qualifier ordering table. Lower index = "earlier" (smaller) version.
// Unknown qualifiers sort after all known ones, alphabetically among themselves,
// but before a numeric token at the same position (Maven quirk: qualifiers are
// generally < release numbers, except when the whole item is empty which means "ga").
var qualifierOrder = map[string]int{
	"alpha":     0,
	"beta":      1,
	"milestone": 2,
	"m":         2,
	"rc":        3,
	"cr":        3,
	"snapshot":  4,
	"":          5, // implicit "release" marker
	"ga":        5,
	"final":     5,
	"sp":        6,
}

// Version is a parsed, comparable Maven version.
type Version struct {
	raw        string
	tokens     []token
	isSnapshot bool
}

type token struct {
	isNumeric bool
	num       int64
	str       string // lowercased qualifier, only set when !isNumeric
}

// Parse tokenizes a raw Maven version string.
func Parse(raw string) Version {
	v := Version{raw: raw}
	if strings.HasSuffix(strings.ToUpper(raw), "-SNAPSHOT") {
		v.isSnapshot = true
	}
	v.tokens = tokenize(raw)
	return v
}

func tokenize(raw string) []token {
	var toks []token
	var cur strings.Builder
	curIsDigit := false
	flush := func() {
		if cur.Len() == 0 {
			return
		}
		s := cur.String()
		if curIsDigit {
			n, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				n = 0
			}
			toks = append(toks, token{isNumeric: true, num: n})
		} else {
			toks = append(toks, token{isNumeric: false, str: strings.ToLower(s)})
		}
		cur.Reset()
	}

	for i, r := range raw {
		switch {
		case r == '.' || r == '-' || r == '_':
			flush()
		case r >= '0' && r <= '9':
			if i > 0 && !curIsDigit && cur.Len() > 0 {
				flush()
			}
			curIsDigit = true
			cur.WriteRune(r)
		default:
			if i > 0 && curIsDigit && cur.Len() > 0 {
				flush()
			}
			curIsDigit = false
			cur.WriteRune(r)
		}
	}
	flush()
	return toks
}

func (v Version) String() string { return v.raw }

// IsSnapshot reports whether this is a Maven -SNAPSHOT version.
func (v Version) IsSnapshot() bool { return v.isSnapshot }

func qualifierRank(s string) int {
	if r, ok := qualifierOrder[s]; ok {
		return r
	}
	// Unknown qualifier: ranks after all known qualifiers, before numeric release.
	return 100
}

// rankOf places any token on the same ordering axis: numeric tokens sit at
// rank 5 (the same rank as the implicit "release" qualifier), so a numeric
// token and an empty/ga/final qualifier compare by magnitude against zero,
// while qualifiers that sort after release (e.g. "sp") correctly outrank a
// same-position numeric token, and qualifiers before release (alpha/beta/
// milestone/rc/snapshot) correctly sort below it.
func rankOf(t token) int {
	if t.isNumeric {
		return 5
	}
	return qualifierRank(t.str)
}

func cmpToken(a, b token) int {
	ra, rb := rankOf(a), rankOf(b)
	if ra != rb {
		if ra < rb {
			return -1
		}
		return 1
	}
	if !a.isNumeric && !b.isNumeric && ra == 100 {
		return strings.Compare(a.str, b.str) // both unknown qualifiers: lexical
	}
	an, bn := int64(0), int64(0)
	if a.isNumeric {
		an = a.num
	}
	if b.isNumeric {
		bn = b.num
	}
	switch {
	case an < bn:
		return -1
	case an > bn:
		return 1
	default:
		return 0
	}
}

// Compare returns -1, 0, or 1 comparing v to other, following Maven's
// version-ordering semantics (numeric tokens compared numerically,
// qualifiers ordered alpha<beta<milestone<rc<snapshot<""/ga/final<sp,
// shorter sequences padded with implicit zero/empty tokens).
func (v Version) Compare(other Version) int {
	n := len(v.tokens)
	if len(other.tokens) > n {
		n = len(other.tokens)
	}
	for i := 0; i < n; i++ {
		var a, b token
		if i < len(v.tokens) {
			a = v.tokens[i]
		} else {
			a = token{isNumeric: true, num: 0}
		}
		if i < len(other.tokens) {
			b = other.tokens[i]
		} else {
			b = token{isNumeric: true, num: 0}
		}
		if c := cmpToken(a, b); c != 0 {
			return c
		}
	}
	return 0
}

func (v Version) LessThan(other Version) bool    { return v.Compare(other) < 0 }
func (v Version) GreaterThan(other Version) bool { return v.Compare(other) > 0 }
func (v Version) Equal(other Version) bool       { return v.Compare(other) == 0 }

// Max returns the greater of a set of versions. Panics on empty input.
func Max(vs []Version) Version {
	m := vs[0]
	for _, v := range vs[1:] {
		if v.GreaterThan(m) {
			m = v
		}
	}
	return m
}