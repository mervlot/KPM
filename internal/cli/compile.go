package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"kpm/internal/classpath"
	"kpm/internal/compiler"
	"kpm/internal/config"
	"kpm/internal/lockfile"
	"kpm/internal/logger"
	"kpm/internal/project"
	"kpm/internal/utils"
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
func cmdCopyJarResources() int {
	proj, err := project.Load(config.ManifestFile)
	if err != nil {
		return fail(err)
	}
	src := fmt.Sprintf("%s/main/resources", proj.SourceDir)

	dst := fmt.Sprintf("%s/classes/resources", proj.BuildDir)

	if err := utils.CopyDir(src, dst); err != nil {
		fmt.Println("Error:", err)
		return fail(err)
	}
	return 0
}

// func cmdTest(args []string) int {
// 	manifest, _, _, err := setup(false)
// 	if err != nil {
// 		logger.PrintDiagnostic(diagnosticFor(err))
// 		return 1
// 	}

// 	// Compile the test sources.
// 	if err := compile(manifest.TestDir, "build/classes-test", modeAuto); err != nil {
// 		logger.PrintDiagnostic(diagnosticFor(err))
// 		return 1
// 	}

// 	cp, err := buildClasspath(
// 		"build/classes",
// 		"build/classes-test",
// 	)
// 	if err != nil {
// 		logger.PrintDiagnostic(diagnosticFor(err))
// 		return 1
// 	}

// 	junitJar, err := findJUnitConsoleJar()
// 	if err != nil {
// 		logger.PrintDiagnostic(logger.Diagnostic{
// 			Title:  "JUnit Console Launcher not found",
// 			Detail: "Install org.junit.platform:junit-platform-console-standalone",
// 			Fixes: []string{
// 				"kpm add org.junit.platform:junit-platform-console-standalone:<version>",
// 			},
// 		})
// 		return 1
// 	}

// 	cmd := exec.Command(
// 		"java",
// 		"-jar", junitJar,
// 		"--class-path", cp,
// 		"--scan-class-path",
// 	)

// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	cmd.Stdin = os.Stdin

// 	if err := cmd.Run(); err != nil {
// 		return 1
// 	}

//		return 0
//	}
func cmdJar() int {
	javaPath, err := compiler.LocateJar()
	if err != nil {
		return fail(err)
	}
	proj, err := project.Load(config.ManifestFile)
	if err != nil {
		return fail(err)
	}

	var cmdArgs []string = []string{
		"cf",
		fmt.Sprintf("./%s/libs/%s.jar", proj.BuildDir, proj.Name),
		"-C",
		fmt.Sprintf("./%s/classes", proj.BuildDir), ".",
	}
	src := fmt.Sprintf("%s/main/resources", proj.SourceDir)
	_, err = os.ReadDir(src)
	if err == nil {
		cmdCopyJarResources()
	}

	cmd := exec.Command(javaPath, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// The program ran and exited non-zero on its own — that's the
			// program's business, not a KPM failure, so just propagate its
			// exit code without wrapping it in a "resolution failed"-style
			// diagnostic.
			return exitErr.ExitCode()
		}
		return fail(fmt.Errorf("err %w", err))
	}
	return 0
}

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
				"Add .java/.kt files under src/main/java or src/main/kotlin (per kpm.json's \"sourceDir\")",
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
