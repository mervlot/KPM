package cli

import (
	"errors"
	"fmt"

	"kpm/internal/classpath"
	"kpm/internal/compiler"
	"kpm/internal/config"
	"kpm/internal/lockfile"
	"kpm/internal/logger"
	"kpm/internal/project"
)

// compileMode selects which of Manager's three entry points cmdCompile
// calls — "auto" (kpm compile / @compile), "java" (@compile-java), or
// "kotlin" (@compile-kotlin). This is the only place that decides which
// Manager method to call; Manager itself does the actual strategy logic.
type compileMode string

const (
	modeAuto   compileMode = "auto"
	modeJava   compileMode = "java"
	modeKotlin compileMode = "kotlin"
)

// cmdCompile loads the project model, builds the compile classpath from
// what's already installed (never triggering a download — see
// internal/classpath's package doc), and hands both to compiler.Manager.
// This function and everything it calls stays entirely inside
// internal/cli, internal/project, internal/classpath, and
// internal/compiler — javac/kotlinc are only ever invoked from inside
// internal/compiler.
func cmdCompile(args []string, mode compileMode) int {
	proj, err := project.Load(config.ManifestFile)
	if err != nil {
		return fail(err)
	}

	cp, err := classpath.Build(lockfile.FileName, proj.InstalledJarsDir, classpath.Compile)
	if err != nil {
		return fail(err)
	}
	proj.CompileClasspath = cp.Entries

	mgr := compiler.NewManager()

	switch mode {
	case modeJava:
		err = mgr.CompileJava(proj, cp.String())
	case modeKotlin:
		err = mgr.CompileKotlin(proj, cp.String())
	default:
		err = mgr.CompileAuto(proj, cp.String())
	}
	if err != nil {
		return fail(err)
	}

	fmt.Println("Compiled sources to", proj.ClassesDir)
	return 0
}

// diagnosticForCompile adds titled, actionable diagnostics for the
// compiler/classpath error types. diagnosticFor (main.go) tries this first
// and falls through to its own generic cases if it returns ok=false —
// kept in this file since it's compile-specific, rather than growing one
// giant switch in main.go.
func diagnosticForCompile(err error) (logger.Diagnostic, bool) {
	var missingCompiler *compiler.MissingCompilerError
	if errors.As(err, &missingCompiler) {
		fixes := []string{
			fmt.Sprintf("Install a JDK/Kotlin toolchain that provides %s", missingCompiler.Name),
			fmt.Sprintf("Make sure it's on PATH, or set %s", missingCompiler.HomeVar),
		}
		return logger.Diagnostic{
			Title:  fmt.Sprintf("%s not found", missingCompiler.Name),
			Detail: missingCompiler.Error(),
			Fixes:  fixes,
		}, true
	}

	var compileErr *compiler.CompileError
	if errors.As(err, &compileErr) {
		return logger.Diagnostic{
			Title:  fmt.Sprintf("%s compilation failed", compileErr.Compiler),
			Detail: compileErr.Error(),
			Fixes: []string{
				"Fix the error(s) reported above and re-run `kpm compile` (or your compile task)",
			},
		}, true
	}

	var emptySrc *compiler.EmptySourceError
	if errors.As(err, &emptySrc) {
		return logger.Diagnostic{
			Title:  "No source to compile",
			Detail: emptySrc.Error(),
			Fixes: []string{
				"Add .java/.kt files under src/main/java or src/main/kotlin (per package.kpm's \"sourceDir\")",
			},
		}, true
	}

	var missingLock *classpath.MissingLockfileError
	if errors.As(err, &missingLock) {
		return logger.Diagnostic{
			Title:  "No resolved dependencies",
			Detail: missingLock.Error(),
			Fixes:  []string{"Run `kpm install`"},
		}, true
	}

	var missingJar *classpath.MissingJarError
	if errors.As(err, &missingJar) {
		return logger.Diagnostic{
			Title:  "Missing dependency jar",
			Detail: missingJar.Error(),
			Fixes:  []string{"Run `kpm install` to re-download it"},
		}, true
	}

	var conflict *classpath.ConflictError
	if errors.As(err, &conflict) {
		return logger.Diagnostic{
			Title:  "Classpath conflict",
			Detail: conflict.Error(),
			Fixes:  []string{"Run `kpm sync` to regenerate kpm.lock"},
		}, true
	}

	return logger.Diagnostic{}, false
}