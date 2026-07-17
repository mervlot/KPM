package compiler

import (
	"fmt"
	"os"

	"kpm/internal/project"
)

// Manager decides which compiler(s) a project needs and runs them in the
// right order. This is the only exported entry point internal/cli and
// internal/run/executor should call — neither ever constructs a
// JavaCompiler/KotlinCompiler or touches os/exec directly.
type Manager struct {
	Java   Compiler
	Kotlin Compiler
}

func NewManager() *Manager {
	return &Manager{Java: JavaCompiler{}, Kotlin: KotlinCompiler{}}
}

// CompileAuto inspects the project and picks the strategy automatically:
//
//   - Java only            -> javac
//   - Kotlin only          -> kotlinc
//   - Neither              -> EmptySourceError
//   - Both (mixed project) -> kotlinc first, then javac (see below)
//
// Why kotlinc-first for mixed projects, not the reverse or a "smart"
// dependency-order analysis: javac has no idea what a .kt file is and
// cannot be pointed at Kotlin source at all, so "javac first" isn't a
// choice that exists. kotlinc, on the other hand, DOES understand .java
// files well enough to resolve symbols Kotlin code references from Java —
// it just won't emit .class files for them. So the only working order is:
//  1. kotlinc compiles the Kotlin sources, with the Java sources also
//     listed on its command line purely so it can resolve any
//     Kotlin-calls-Java references, writing .class output for the Kotlin
//     side to ClassesDir.
//  2. javac compiles the Java sources, with ClassesDir appended to its
//     classpath so it can resolve any Java-calls-Kotlin references against
//     the .class files kotlinc just produced, writing its own .class
//     output to the same ClassesDir.
//
// This is the same strategy Gradle's kotlin-jvm plugin uses for mixed
// modules, and it's the only direction that actually works given each
// compiler's capabilities — it's not a "safest of several options" choice,
// it's the only one that isn't a dead end.
func (m *Manager) CompileAuto(p *project.Project, cp string) error {
	javaFiles, err := p.JavaFiles()
	if err != nil {
		return err
	}
	kotlinFiles, err := p.KotlinFiles()
	if err != nil {
		return err
	}

	if len(javaFiles) == 0 && len(kotlinFiles) == 0 {
		return &EmptySourceError{}
	}
	if err := p.EnsureBuildDirs(); err != nil {
		return err
	}

	switch {
	case len(kotlinFiles) > 0 && len(javaFiles) == 0:
		return m.Kotlin.Compile(Input{KotlinSources: kotlinFiles, Classpath: cp, OutDir: p.ClassesDir})

	case len(javaFiles) > 0 && len(kotlinFiles) == 0:
		return m.Java.Compile(Input{JavaSources: javaFiles, Classpath: cp, OutDir: p.ClassesDir})

	default:
		return m.compileMixed(p, javaFiles, kotlinFiles, cp)
	}
}

func (m *Manager) compileMixed(p *project.Project, javaFiles, kotlinFiles []string, cp string) error {
	// Step 1: kotlinc, with Java sources included for cross-reference only.
	if err := m.Kotlin.Compile(Input{
		KotlinSources: kotlinFiles,
		JavaSources:   javaFiles,
		Classpath:     cp,
		OutDir:        p.ClassesDir,
	}); err != nil {
		return fmt.Errorf("compiling Kotlin sources: %w", err)
	}

	// Step 2: javac, with the Kotlin output dir appended to its classpath so
	// it can resolve Java-calls-Kotlin references.
	javaClasspath := cp
	if javaClasspath != "" {
		javaClasspath += string(os.PathListSeparator) + p.ClassesDir
	} else {
		javaClasspath = p.ClassesDir
	}
	if err := m.Java.Compile(Input{
		JavaSources: javaFiles,
		Classpath:   javaClasspath,
		OutDir:      p.ClassesDir,
	}); err != nil {
		return fmt.Errorf("compiling Java sources: %w", err)
	}
	return nil
}

// CompileJava forces Java-only compilation (used by the "@compile-java"
// kpm.run builtin), regardless of whether Kotlin sources also exist.
func (m *Manager) CompileJava(p *project.Project, cp string) error {
	javaFiles, err := p.JavaFiles()
	if err != nil {
		return err
	}
	if len(javaFiles) == 0 {
		return &EmptySourceError{Language: "Java", Dir: p.JavaSourceDir}
	}
	if err := p.EnsureBuildDirs(); err != nil {
		return err
	}
	return m.Java.Compile(Input{JavaSources: javaFiles, Classpath: cp, OutDir: p.ClassesDir})
}

// CompileKotlin forces Kotlin-only compilation (used by "@compile-kotlin"),
// regardless of whether Java sources also exist. Java sources are still
// passed through for cross-reference resolution (see KotlinCompiler.Compile)
// but are not compiled themselves — call CompileJava separately if both
// outputs are needed without CompileAuto's ordering.
func (m *Manager) CompileKotlin(p *project.Project, cp string) error {
	kotlinFiles, err := p.KotlinFiles()
	if err != nil {
		return err
	}
	if len(kotlinFiles) == 0 {
		return &EmptySourceError{Language: "Kotlin", Dir: p.KotlinSourceDir}
	}
	javaFiles, _ := p.JavaFiles()
	if err := p.EnsureBuildDirs(); err != nil {
		return err
	}
	return m.Kotlin.Compile(Input{KotlinSources: kotlinFiles, JavaSources: javaFiles, Classpath: cp, OutDir: p.ClassesDir})
}

// CheckToolchain verifies that whichever compiler(s) the project actually
// needs are locatable, WITHOUT requiring a compiler for a language the
// project doesn't use (a pure-Java project should never be blocked on a
// missing kotlinc). Intended for a future `kpm doctor`-style pre-flight
// check; not required for CompileAuto itself, which naturally fails with
// MissingCompilerError from whichever Locate() it actually needs.
func (m *Manager) CheckToolchain(p *project.Project) error {
	if p.HasJavaSources() {
		if _, err := m.Java.Locate(); err != nil {
			return err
		}
	}
	if p.HasKotlinSources() {
		if _, err := m.Kotlin.Locate(); err != nil {
			return err
		}
	}
	return nil
}