package compiler

import (
	"os"
	"path/filepath"
	"testing"

	"kpm/internal/project"
)

func tempProject(t *testing.T) *project.Project {
	t.Helper()
	dir := t.TempDir()
	return &project.Project{
		BuildDir:     filepath.Join(dir, "build"),
		ClassesDir:   filepath.Join(dir, "build/classes"),
		GeneratedDir: filepath.Join(dir, "build/generated"),
		TmpDir:       filepath.Join(dir, "build/tmp"),
		LibsDir:      filepath.Join(dir, "build/libs"),
	}
}

func TestCompileAutoEmptySourceError(t *testing.T) {
	mgr := NewManager()
	p := tempProject(t)
	err := mgr.CompileAuto(p, "")
	if err == nil {
		t.Fatal("expected EmptySourceError for a project with no source at all")
	}
	if _, ok := err.(*EmptySourceError); !ok {
		t.Fatalf("expected *EmptySourceError, got %T: %v", err, err)
	}
}

// TestCompileAutoMixedEndToEnd exercises the real kotlinc-then-javac
// strategy against actual toolchains when both are available, skipping
// gracefully otherwise. This is the architecturally significant path —
// unit-testing the branch selection alone wouldn't catch a broken
// cross-reference (e.g. javac unable to see kotlinc's output), which is
// exactly the failure mode this design has to avoid.
func TestCompileAutoMixedEndToEnd(t *testing.T) {
	mgr := NewManager()
	if _, err := mgr.Java.Locate(); err != nil {
		t.Skip("javac not available, skipping mixed end-to-end test:", err)
	}
	if _, err := mgr.Kotlin.Locate(); err != nil {
		t.Skip("kotlinc not available, skipping mixed end-to-end test:", err)
	}

	dir := t.TempDir()
	p := &project.Project{
		BuildDir:        filepath.Join(dir, "build"),
		ClassesDir:      filepath.Join(dir, "build/classes"),
		GeneratedDir:    filepath.Join(dir, "build/generated"),
		TmpDir:          filepath.Join(dir, "build/tmp"),
		LibsDir:         filepath.Join(dir, "build/libs"),
		JavaSourceDir:   filepath.Join(dir, "src/main/java"),
		KotlinSourceDir: filepath.Join(dir, "src/main/kotlin"),
	}

	writeFile(t, filepath.Join(p.KotlinSourceDir, "Greeter.kt"), `
class Greeter {
    fun greet(): String = "hi"
}
`)
	writeFile(t, filepath.Join(p.JavaSourceDir, "Main.java"), `
public class Main {
    public static void main(String[] args) {
        Greeter g = new Greeter();
        System.out.println(g.greet());
    }
}
`)

	if err := mgr.CompileAuto(p, ""); err != nil {
		t.Fatalf("unexpected error compiling mixed sources: %v", err)
	}

	for _, class := range []string{"Main.class", "Greeter.class"} {
		if _, err := os.Stat(filepath.Join(p.ClassesDir, class)); err != nil {
			t.Errorf("expected %s to exist after mixed compile: %v", class, err)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}