package classpath

import (
	"os"
	"path/filepath"
	"testing"

	"kpm/internal/resolver"
)

func writeLockfile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeFakeJar(t *testing.T, libsRoot, group, artifact, version string) {
	t.Helper()
	path := resolver.ArtifactPath(libsRoot, group, artifact, version, "", "jar")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("fake jar bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestMissingLockfile(t *testing.T) {
	dir := t.TempDir()
	_, err := Build(filepath.Join(dir, "kpm.lock"), filepath.Join(dir, "libs"), Compile)
	if err == nil {
		t.Fatal("expected error for missing lock file")
	}
	if _, ok := err.(*MissingLockfileError); !ok {
		t.Fatalf("expected *MissingLockfileError, got %T: %v", err, err)
	}
}

func TestMissingJar(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "kpm.lock")
	writeLockfile(t, lockPath, `{"schemaVersion":1,"entries":{
		"com.example:foo::jar@1.0.0": {"group":"com.example","artifact":"foo","version":"1.0.0","type":"jar","scope":"compile"}
	}}`)
	// deliberately do NOT write the jar
	_, err := Build(lockPath, filepath.Join(dir, "libs"), Compile)
	if err == nil {
		t.Fatal("expected error for missing jar")
	}
	if _, ok := err.(*MissingJarError); !ok {
		t.Fatalf("expected *MissingJarError, got %T: %v", err, err)
	}
}

func TestScopeFiltering(t *testing.T) {
	dir := t.TempDir()
	libsRoot := filepath.Join(dir, "libs")
	lockPath := filepath.Join(dir, "kpm.lock")
	writeLockfile(t, lockPath, `{"schemaVersion":1,"entries":{
		"com.example:compiled::jar@1.0.0": {"group":"com.example","artifact":"compiled","version":"1.0.0","type":"jar","scope":"compile"},
		"com.example:provided-lib::jar@1.0.0": {"group":"com.example","artifact":"provided-lib","version":"1.0.0","type":"jar","scope":"provided"},
		"com.example:runtime-lib::jar@1.0.0": {"group":"com.example","artifact":"runtime-lib","version":"1.0.0","type":"jar","scope":"runtime"},
		"com.example:test-lib::jar@1.0.0": {"group":"com.example","artifact":"test-lib","version":"1.0.0","type":"jar","scope":"test"}
	}}`)
	for _, a := range []string{"compiled", "provided-lib", "runtime-lib", "test-lib"} {
		writeFakeJar(t, libsRoot, "com.example", a, "1.0.0")
	}

	compileCp, err := Build(lockPath, libsRoot, Compile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContainsArtifact(t, compileCp, "compiled")
	assertContainsArtifact(t, compileCp, "provided-lib")
	assertNotContainsArtifact(t, compileCp, "runtime-lib")
	assertNotContainsArtifact(t, compileCp, "test-lib")

	runtimeCp, err := Build(lockPath, libsRoot, Runtime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContainsArtifact(t, runtimeCp, "compiled")
	assertContainsArtifact(t, runtimeCp, "runtime-lib")
	assertNotContainsArtifact(t, runtimeCp, "provided-lib")
	assertNotContainsArtifact(t, runtimeCp, "test-lib")
}

func TestPomTypeAndClassifierSkipped(t *testing.T) {
	dir := t.TempDir()
	libsRoot := filepath.Join(dir, "libs")
	lockPath := filepath.Join(dir, "kpm.lock")
	writeLockfile(t, lockPath, `{"schemaVersion":1,"entries":{
		"com.example:parent::pom@1.0.0": {"group":"com.example","artifact":"parent","version":"1.0.0","type":"pom","scope":"compile"},
		"com.example:foo:sources:jar@1.0.0": {"group":"com.example","artifact":"foo","version":"1.0.0","type":"jar","classifier":"sources","scope":"compile"}
	}}`)
	cp, err := Build(lockPath, libsRoot, Compile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cp.Entries) != 0 {
		t.Errorf("expected no entries (pom has no jar, sources classifier excluded), got %v", cp.Entries)
	}
}

func TestClasspathStringSeparator(t *testing.T) {
	cp := &Classpath{Entries: []string{"a.jar", "b.jar"}}
	got := cp.String()
	want := "a.jar" + string(os.PathListSeparator) + "b.jar"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func assertContainsArtifact(t *testing.T, cp *Classpath, artifact string) {
	t.Helper()
	for _, e := range cp.Entries {
		if containsSubstr(e, artifact) {
			return
		}
	}
	t.Errorf("expected classpath to contain %q, got %v", artifact, cp.Entries)
}

func assertNotContainsArtifact(t *testing.T, cp *Classpath, artifact string) {
	t.Helper()
	for _, e := range cp.Entries {
		if containsSubstr(e, artifact) {
			t.Errorf("expected classpath NOT to contain %q, got %v", artifact, cp.Entries)
			return
		}
	}
}

func containsSubstr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}