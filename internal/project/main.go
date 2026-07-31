// Package project defines the shared project model every build-related
// subsystem (classpath, compiler, and later @jar/@run/@test/publishing)
// builds from, so none of them re-read kpm.json or re-scan the
// filesystem independently. Load it once per command, pass the *Project
// around.
//
// Deliberate deviation from a generic "SourceDirs []string": this version
// only supports exactly two fixed, known source roots — src/main/java and
// src/main/kotlin — per the current milestone's scope (tests, multi-module,
// and arbitrary source sets are explicitly out of scope for now). Explicit
// JavaSourceDir/KotlinSourceDir fields say plainly what's supported today;
// a caller can't accidentally treat a mystery slice entry as "the Kotlin
// dir" via string-suffix guessing. When multi-module/custom source sets
// are added later, this can grow into `SourceSets []SourceSet` alongside
// (not instead of) these two fields, without breaking anything that reads
// them now — see the "Future-proofing" note below.
package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"kpm/internal/config"
)

// Project is the resolved, on-disk shape of a KPM project: what source
// exists, where resources live, where already-installed dependency jars
// are, and where build output should go. Classpath and compiler resolution
// both take a *Project rather than re-deriving any of this themselves.
type Project struct {
	Name    string
	Version string
	// SourceDir is kpm.json's "sourceDir" (e.g. "src"); JavaSourceDir/
	// KotlinSourceDir below are resolved as SourceDir/main/java and
	// SourceDir/main/kotlin.
	SourceDir string

	// JavaSourceDir / KotlinSourceDir are "" if that source root doesn't
	// exist on disk. Both may be set (mixed-language project), one, or
	// neither (see HasJavaSources/HasKotlinSources).
	JavaSourceDir   string
	KotlinSourceDir string
	ResourceDir     string // src/main/resources, "" if absent

	// Test source roots — "" if the directory doesn't exist.
	TestJavaSourceDir   string // src/test/java
	TestKotlinSourceDir string // src/test/kotlin

	// InstalledJarsDir is where `kpm install` already placed dependency
	// jars ("./libs" today). This is a READ-ONLY input to the compiler —
	// compilation must never trigger a download, so this is the one and
	// only place compile/runtime classpath resolution is allowed to look.
	InstalledJarsDir string

	// CompileClasspath / RuntimeClasspath are populated by
	// internal/classpath.Build, not by Load — a fresh *Project has these
	// empty. Kept on the struct (rather than passed around separately) so
	// every downstream consumer (compiler today; @jar/@run tomorrow) reads
	// classpath off the same object it reads source dirs off.
	CompileClasspath []string
	RuntimeClasspath []string

	// Build output layout. Fixed, not configurable yet (see
	// "Future-proofing" in the package doc) — a project-level override can
	// be added to config.Manifest later without touching any code that
	// reads these fields, since it would just change what Load populates
	// them with.
	BuildDir       string
	ClassesDir     string // build/classes — main compiled output
	TestClassesDir string // build/test-classes — test compiled output
	GeneratedDir   string
	TmpDir         string
	LibsDir        string
}

const defaultInstalledJarsDir = "./libs"

// Load reads kpm.json and discovers which of the two supported source
// roots (src/main/java, src/main/kotlin) actually exist, plus resources.
// It does not require any source to exist — an empty project is valid
// (the compiler layer decides whether "no source" is an error for what
// it's being asked to do).
func Load(manifestPath string) (*Project, error) {
	manifest, err := config.Load(manifestPath)
	if err != nil {
		return nil, err
	}

	sourceDir := manifest.SourceDir
	if sourceDir == "" {
		sourceDir = "src"
	}
	buildDir := manifest.BuildDir
	if buildDir == "" {
		buildDir = "build"
	}

	p := &Project{
		Name:             manifest.Name,
		Version:          manifest.Version,
		SourceDir:        sourceDir,
		InstalledJarsDir: defaultInstalledJarsDir,
		BuildDir:         buildDir,
		ClassesDir:       filepath.Join(buildDir, "classes"),
		TestClassesDir:   filepath.Join(buildDir, "test-classes"),
		GeneratedDir:     filepath.Join(buildDir, "generated"),
		TmpDir:           filepath.Join(buildDir, "tmp"),
		LibsDir:          filepath.Join(buildDir, "libs"),
	}

	javaDir := filepath.Join(sourceDir, "main", "java")
	if isDir(javaDir) {
		p.JavaSourceDir = javaDir
	}
	kotlinDir := filepath.Join(sourceDir, "main", "kotlin")
	if isDir(kotlinDir) {
		p.KotlinSourceDir = kotlinDir
	}
	resourceDir := filepath.Join(sourceDir, "main", "resources")
	if isDir(resourceDir) {
		p.ResourceDir = resourceDir
	}
	if isDir(filepath.Join(sourceDir, "test", "java")) {
		p.TestJavaSourceDir = filepath.Join(sourceDir, "test", "java")
	}
	if isDir(filepath.Join(sourceDir, "test", "kotlin")) {
		p.TestKotlinSourceDir = filepath.Join(sourceDir, "test", "kotlin")
	}

	return p, nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// JavaFiles returns every *.java file under JavaSourceDir, recursively.
// Returns an empty (non-nil) slice, not an error, if JavaSourceDir is ""
// or contains no .java files — "no Java source" is a valid state the
// compiler layer interprets, not a project-loading failure.
func (p *Project) JavaFiles() ([]string, error) {
	return findSourceFiles(p.JavaSourceDir, ".java")
}

// KotlinFiles returns every *.kt file under KotlinSourceDir, recursively.
func (p *Project) KotlinFiles() ([]string, error) {
	return findSourceFiles(p.KotlinSourceDir, ".kt")
}

func findSourceFiles(dir, ext string) ([]string, error) {
	if dir == "" {
		return nil, nil
	}
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ext) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scanning %s for %s sources: %w", dir, ext, err)
	}
	return files, nil
}

// HasJavaSources / HasKotlinSources report whether the corresponding
// source root exists AND contains at least one matching file — an existing
// but empty src/main/java counts as "no Java sources" (see
// compiler.EmptySourceError for how that's surfaced to the user).
func (p *Project) HasJavaSources() bool {
	files, _ := p.JavaFiles()
	return len(files) > 0
}

func (p *Project) HasKotlinSources() bool {
	files, _ := p.KotlinFiles()
	return len(files) > 0
}

// TestJavaFiles returns every *.java file under TestJavaSourceDir.
func (p *Project) TestJavaFiles() ([]string, error) {
	return findSourceFiles(p.TestJavaSourceDir, ".java")
}

// TestKotlinFiles returns every *.kt file under TestKotlinSourceDir.
func (p *Project) TestKotlinFiles() ([]string, error) {
	return findSourceFiles(p.TestKotlinSourceDir, ".kt")
}

// EnsureTestBuildDirs creates build/test-classes if it doesn't exist.
func (p *Project) EnsureTestBuildDirs() error {
	return os.MkdirAll(p.TestClassesDir, 0o755)
}

// EnsureBuildDirs creates the full build/ output layout if it doesn't
// already exist. Idempotent — safe to call on every compile.
func (p *Project) EnsureBuildDirs() error {
	for _, dir := range []string{p.ClassesDir, p.GeneratedDir, p.TmpDir, p.LibsDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating build directory %s: %w", dir, err)
		}
	}
	return nil
}