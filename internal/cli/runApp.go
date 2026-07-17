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
	"kpm/internal/project"
)

// cmdRunApp implements the "@run" kpm.run builtin: `@run <main-class> [args...]`.
//
// It does NOT compile anything itself — pair it with "@compile" (or
// "@compile-java"/"@compile-kotlin") as an earlier step in the same task,
// same as you'd sequence any other build step. What it DOES do is spare you
// from hand-listing every dependency jar on -cp: it builds the exact same
// kind of classpath @compile does (via internal/classpath, reading
// kpm.lock — never downloading anything), just with Runtime scope instead
// of Compile scope, and adds the compiled classes directory in front of it.
func cmdRunApp(args []string) int {
	manifest, err := config.Load(config.ManifestFile)
	if err != nil {
		return fail(err)
	}

	var mainClass string
	var programArgs []string
	if len(args) > 0 {
		mainClass, programArgs = args[0], args[1:]
	} else {
		mainClass = manifest.MainClass
	}
	if mainClass == "" {
		return fail(fmt.Errorf("usage: @run <main-class> [program args...] (or set \"mainClass\" in package.kpm to omit it)"))
	}

	proj, err := project.Load(config.ManifestFile)
	if err != nil {
		return fail(err)
	}

	cp, err := classpath.Build(lockfile.FileName, proj.InstalledJarsDir, classpath.Runtime)
	if err != nil {
		return fail(err)
	}

	javaPath, err := compiler.LocateJavaRuntime()
	if err != nil {
		return fail(err)
	}

	fullClasspath := proj.ClassesDir
	if cp.String() != "" {
		fullClasspath += string(os.PathListSeparator) + cp.String()
	}

	cmdArgs := append([]string{"-cp", fullClasspath, mainClass}, programArgs...)
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
		return fail(fmt.Errorf("launching %s: %w", mainClass, err))
	}
	return 0
}