package parser

import "strings"

// Interpolator resolves ${propertyName} placeholders the way Maven does:
// project.* / pom.* self-references, then explicit <properties>, then a
// fallback for env.* / java system properties (best-effort — KPM doesn't
// run a JVM, so these resolve to empty unless explicitly supplied).
type Interpolator struct {
	props map[string]string
}

func NewInterpolator(pom *POM, extra map[string]string) *Interpolator {
	props := map[string]string{}
	if pom.Properties.Entries != nil {
		for k, v := range pom.Properties.Entries {
			props[k] = v
		}
	}
	// Self-referential built-ins.
	props["project.groupId"] = pom.GroupID
	props["project.artifactId"] = pom.ArtifactID
	props["project.version"] = pom.Version
	props["pom.groupId"] = pom.GroupID
	props["pom.artifactId"] = pom.ArtifactID
	props["pom.version"] = pom.Version
	props["version"] = pom.Version

	// User-supplied overrides (e.g. from kpm.json "properties" block) win.
	for k, v := range extra {
		props[k] = v
	}
	return &Interpolator{props: props}
}

// Resolve expands ${a.b.c} placeholders in s, recursively (bounded to avoid
// infinite loops on self-referential properties), leaving unresolvable
// placeholders untouched so callers can detect and warn about them.
func (in *Interpolator) Resolve(s string) string {
	if !strings.Contains(s, "${") {
		return s
	}
	const maxPasses = 8
	for pass := 0; pass < maxPasses; pass++ {
		changed := false
		var b strings.Builder
		i := 0
		for i < len(s) {
			start := strings.Index(s[i:], "${")
			if start == -1 {
				b.WriteString(s[i:])
				break
			}
			start += i
			end := strings.Index(s[start:], "}")
			if end == -1 {
				b.WriteString(s[i:])
				break
			}
			end += start
			key := s[start+2 : end]
			b.WriteString(s[i:start])
			if val, ok := in.props[key]; ok {
				b.WriteString(val)
				changed = true
			} else {
				b.WriteString(s[start : end+1]) // leave unresolved placeholder as-is
			}
			i = end + 1
		}
		s = b.String()
		if !changed {
			break
		}
	}
	return s
}

// ResolvePOM applies interpolation in-place to the fields that commonly
// contain property references: dependency versions/scopes and parent version.
func (in *Interpolator) ResolvePOM(pom *POM) {
	for i := range pom.Dependencies {
		pom.Dependencies[i].Version = in.Resolve(pom.Dependencies[i].Version)
		pom.Dependencies[i].GroupID = in.Resolve(pom.Dependencies[i].GroupID)
		pom.Dependencies[i].ArtifactID = in.Resolve(pom.Dependencies[i].ArtifactID)
	}
	for i := range pom.DependencyManagement.Dependencies {
		d := &pom.DependencyManagement.Dependencies[i]
		d.Version = in.Resolve(d.Version)
		d.GroupID = in.Resolve(d.GroupID)
		d.ArtifactID = in.Resolve(d.ArtifactID)
	}
	pom.Version = in.Resolve(pom.Version)
}
