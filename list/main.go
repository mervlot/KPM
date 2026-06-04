package list

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/tabwriter"
)

type kpmFile struct {
	KpmVersion string     `json:"kpm version"`
	Resources  []resource `json:"resources"`
}

type resource struct {
	Group   string `json:"group"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Path    string `json:"path"`
	Lpath   string `json:"lpath"`
	Gpath   string `json:"gpath"`
}

// Main lists packages from resource.kpm in a table. If a resource's
// version is empty but a path/URL is present, attempt to infer the
// version from the path or filename so the VERSION column is not blank.
func Main() {
	// Try common locations (cwd and parent dir)
	candidates := []string{"resource.kpm", filepath.Join("..", "resource.kpm")}
	var f *os.File
	var err error
	for _, p := range candidates {
		f, err = os.Open(p)
		if err == nil {
			defer f.Close()
			break
		}
	}
	if f == nil {
		fmt.Fprintln(os.Stderr, "could not find resource.kpm in ./ or ../")
		return
	}

	var k kpmFile
	if err := json.NewDecoder(f).Decode(&k); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse resource.kpm: %v\n", err)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "GROUP\tNAME\tVERSION\tPATH")

	for _, r := range k.Resources {
		ver := strings.TrimSpace(r.Version)
		if ver == "" {
			ver = inferVersionFrom(r.Path)
			if ver == "" {
				ver = inferVersionFrom(r.Lpath)
			}
			if ver == "" {
				ver = inferVersionFrom(r.Gpath)
			}
			if ver == "" {
				ver = "(unknown)"
			}
		}

		path := r.Path
		if path == "" {
			if r.Gpath != "" {
				path = r.Gpath
			} else {
				path = r.Lpath
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Group, r.Name, ver, r.Lpath)
	}

	w.Flush()
}

var versionFileRe = regexp.MustCompile(`[-_]?([0-9]+[A-Za-z0-9._+\-]*)\.jar$`)
var versionSegmentRe = regexp.MustCompile(`^[0-9]+[A-Za-z0-9._\-+]*$`)

// inferVersionFrom tries multiple heuristics to extract a version string
// from a path or URL. It looks at the filename first, then path segments
// (from end to front) and finally any segment containing a digit and a dot.
func inferVersionFrom(p string) string {
	if p == "" {
		return ""
	}
	// strip query params if it's a URL
	if idx := strings.Index(p, "?"); idx >= 0 {
		p = p[:idx]
	}

	// try filename pattern: name-1.2.3.jar or name_1.2.3.jar
	base := filepath.Base(p)
	if m := versionFileRe.FindStringSubmatch(base); m != nil && len(m) >= 2 {
		return m[1]
	}

	// split path and examine segments from the end
	parts := strings.Split(p, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		seg := parts[i]
		if seg == "" {
			continue
		}
		seg = strings.TrimSuffix(seg, ".jar")
		if versionSegmentRe.MatchString(seg) {
			// return segments like 1.2.3, 32.1.2-jre, 3.13.0
			return seg
		}
		// also accept segments containing both digit and dot
		if strings.Contains(seg, ".") && regexp.MustCompile(`[0-9]`).MatchString(seg) {
			return seg
		}
	}

	// nothing found
	return ""
}
