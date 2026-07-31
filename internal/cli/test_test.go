package cli

import (
	"os"
	"path/filepath"
	"testing"

	"kpm/internal/resolver"
)

func TestMissingTestDepsAllMissing(t *testing.T) {
	missing := missingTestDeps(t.TempDir())
	if len(missing) != len(defaultTestDeps) {
		t.Fatalf("expected all %d default test deps missing, got %d", len(defaultTestDeps), len(missing))
	}
}

func TestMissingTestDepsSomeInstalled(t *testing.T) {
	dir := t.TempDir()
	first := defaultTestDeps[0]
	group, artifact, _ := splitGroupArtifact(first.Coord)
	path := resolver.ArtifactPath(dir, group, artifact, first.Version, "", "jar")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("fake"), 0o644); err != nil {
		t.Fatal(err)
	}

	missing := missingTestDeps(dir)
	if len(missing) != len(defaultTestDeps)-1 {
		t.Fatalf("expected %d missing, got %d: %v", len(defaultTestDeps)-1, len(missing), missing)
	}
	for _, d := range missing {
		if d.Coord == first.Coord {
			t.Errorf("expected %s to be reported as present, not missing", first.Coord)
		}
	}
}

func TestSplitGroupArtifact(t *testing.T) {
	g, a, ok := splitGroupArtifact("org.junit.jupiter:junit-jupiter-api")
	if !ok || g != "org.junit.jupiter" || a != "junit-jupiter-api" {
		t.Errorf("got group=%q artifact=%q ok=%v", g, a, ok)
	}
	if _, _, ok := splitGroupArtifact("no-colon-here"); ok {
		t.Error("expected ok=false for a coordinate with no colon")
	}
}