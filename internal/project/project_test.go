package project

import (
	"os"
	"path/filepath"
	"testing"
)

// writeManifest + writeFile are small test helpers to build a throwaway
// project directory on disk, since Load/JavaFiles/KotlinFiles all read
// from real paths (no filesystem abstraction — keeping this package simple
// was preferred over introducing an fs.FS layer for a milestone this
// scoped).
func writeManifest(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "kpm.json")
	content := `{"name":"demo","version":"0.1.0","maindir":"./src","dependencies":{}}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
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

func TestLoadJavaOnly(t *testing.T) {
	dir := t.TempDir()
	manifest := writeManifest(t, dir)
	writeFile(t, filepath.Join(dir, "src/main/java/com/example/Hello.java"), "package com.example; class Hello {}")

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir(dir)

	p, err := Load(filepath.Base(manifest))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.JavaSourceDir == "" {
		t.Error("expected JavaSourceDir to be set")
	}
	if p.KotlinSourceDir != "" {
		t.Error("expected KotlinSourceDir to be empty (no kotlin dir exists)")
	}
	if !p.HasJavaSources() {
		t.Error("expected HasJavaSources() == true")
	}
	if p.HasKotlinSources() {
		t.Error("expected HasKotlinSources() == false")
	}
	files, err := p.JavaFiles()
	if err != nil || len(files) != 1 {
		t.Fatalf("expected exactly 1 java file, got %v (err=%v)", files, err)
	}
}

func TestLoadEmptyProjectIsValid(t *testing.T) {
	dir := t.TempDir()
	manifest := writeManifest(t, dir)

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir(dir)

	p, err := Load(filepath.Base(manifest))
	if err != nil {
		t.Fatalf("unexpected error loading a project with no source at all: %v", err)
	}
	if p.HasJavaSources() || p.HasKotlinSources() {
		t.Error("expected no sources to be detected")
	}
}

func TestExistingButEmptySourceDirCountsAsNoSources(t *testing.T) {
	dir := t.TempDir()
	manifest := writeManifest(t, dir)
	// directory exists but has no .java files in it — just a stray file
	writeFile(t, filepath.Join(dir, "src/main/java/README.txt"), "not a source file")

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir(dir)

	p, err := Load(filepath.Base(manifest))
	if err != nil {
		t.Fatal(err)
	}
	if p.JavaSourceDir == "" {
		t.Fatal("expected the directory itself to be detected as present")
	}
	if p.HasJavaSources() {
		t.Error("expected HasJavaSources() == false: directory exists but contains no .java files")
	}
}

func TestEnsureBuildDirs(t *testing.T) {
	dir := t.TempDir()
	p := &Project{
		BuildDir:     filepath.Join(dir, "build"),
		ClassesDir:   filepath.Join(dir, "build/classes"),
		GeneratedDir: filepath.Join(dir, "build/generated"),
		TmpDir:       filepath.Join(dir, "build/tmp"),
		LibsDir:      filepath.Join(dir, "build/libs"),
	}
	if err := p.EnsureBuildDirs(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, d := range []string{p.ClassesDir, p.GeneratedDir, p.TmpDir, p.LibsDir} {
		if info, err := os.Stat(d); err != nil || !info.IsDir() {
			t.Errorf("expected %s to exist as a directory", d)
		}
	}
}
