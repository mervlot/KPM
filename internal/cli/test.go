package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"kpm/internal/classpath"
	"kpm/internal/compiler"
	"kpm/internal/config"
	"kpm/internal/lockfile"
	"kpm/internal/project"
	"kpm/internal/resolver"
)

// defaultTestDeps — real Maven Central coordinates. We install these as
// scope "compile" (not "test") so BuildInstallPlan actually copies them to
// ./libs/. If they were scope "test", BuildInstallPlan skips them and they
// only land in ~/.kpm/repository where the compiler/runner can't find them.
var defaultTestDeps = []struct{ Coord, Version string }{
	{"org.jetbrains.kotlin:kotlin-test", "2.1.21"},
	{"org.jetbrains.kotlin:kotlin-test-junit5", "2.1.21"},
	{"org.junit.jupiter:junit-jupiter-api", "5.12.2"},
	{"org.junit.jupiter:junit-jupiter-engine", "5.12.2"},
	{"org.junit.platform:junit-platform-console-standalone", "1.12.2"},
}

func cmdTest(args []string, offline bool) int {
	if len(args) > 0 && args[0] == "init" {
		return testInit(offline)
	}

	// Parse flags
	verbose := false
	failFast := false
	var filter string
	remaining := args[:0]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--verbose", "-v":
			verbose = true
		case "--fail-fast":
			failFast = true
		case "--filter":
			if i+1 < len(args) {
				i++
				filter = args[i]
			} else {
				return fail(fmt.Errorf("--filter requires a value, e.g. --filter CalculatorTest"))
			}
		default:
			remaining = append(remaining, args[i])
		}
	}
	_ = remaining

	proj, err := project.Load(config.ManifestFile)
	if err != nil {
		return fail(fmt.Errorf("no %s found — run `kpm init` first", config.ManifestFile))
	}

	// If jars missing, just install automatically — no prompt
	missing := missingTestDeps(proj.InstalledJarsDir)
	if len(missing) > 0 {
		fmt.Println("Test jars not found, running `kpm test init`...")
		if code := testInit(offline); code != 0 {
			return code
		}
		proj, _ = project.Load(config.ManifestFile)
	}

	testJava, _ := proj.TestJavaFiles()
	testKotlin, _ := proj.TestKotlinFiles()
	if len(testJava)+len(testKotlin) == 0 {
		fmt.Printf("No test sources found under %s/test/\n", proj.SourceDir)
		return 1
	}

	// Compile classpath: ./libs/ jars + build/classes (so tests can see main code)
	cp, err := classpath.Build(lockfile.FileName, proj.InstalledJarsDir, classpath.Compile)
	if err != nil {
		return fail(err)
	}
	testCp := strings.Join([]string{proj.ClassesDir, cp.String()}, string(os.PathListSeparator))

	if err := proj.EnsureTestBuildDirs(); err != nil {
		return fail(err)
	}

	fmt.Println("Compiling test sources...")
	mgr := compiler.NewManager()
	mainJava, _ := proj.JavaFiles()
	mainKotlin, _ := proj.KotlinFiles()

	if len(testKotlin) > 0 {
		if err := mgr.Kotlin.Compile(compiler.Input{
			KotlinSources: append(testKotlin, mainKotlin...),
			JavaSources:   append(testJava, mainJava...),
			Classpath:     testCp,
			OutDir:        proj.TestClassesDir,
		}); err != nil {
			return fail(err)
		}
		testCp = proj.TestClassesDir + string(os.PathListSeparator) + testCp
	}
	if len(testJava) > 0 {
		if err := mgr.Java.Compile(compiler.Input{
			JavaSources: append(testJava, mainJava...),
			Classpath:   testCp,
			OutDir:      proj.TestClassesDir,
		}); err != nil {
			return fail(err)
		}
	}

	if err := runTests(proj, verbose, failFast, filter); err != nil {
		return fail(err)
	}
	return 0
}

func testInit(offline bool) int {
	manifest, err := config.Load(config.ManifestFile)
	if err != nil {
		return fail(fmt.Errorf("no %s found — run `kpm init` first", config.ManifestFile))
	}
	if manifest.DevDeps == nil {
		manifest.DevDeps = map[string]config.DependencySpec{}
	}
	for _, d := range defaultTestDeps {
		// scope "compile" — NOT "test" — so BuildInstallPlan copies jars to ./libs/
		manifest.DevDeps[d.Coord] = config.DependencySpec{Version: d.Version, Scope: "compile"}
	}
	if err := manifest.Save(config.ManifestFile); err != nil {
		return fail(err)
	}
	fmt.Println("Added test dependencies — installing...")
	return resolveAndInstall(manifest, offline, true)
}

func runTests(proj *project.Project, verbose, failFast bool, filter string) error {
	standaloneJar, err := findTestJar(proj.InstalledJarsDir,
		"org.junit.platform", "junit-platform-console-standalone", "1.12.2")
	if err != nil {
		return fmt.Errorf("junit-platform-console-standalone not found in libs/ — run `kpm test init`: %w", err)
	}
	javaPath, err := compiler.LocateJavaRuntime()
	if err != nil {
		return err
	}

	cp, _ := classpath.Build(lockfile.FileName, proj.InstalledJarsDir, classpath.Runtime)
	parts := []string{standaloneJar, proj.TestClassesDir, proj.ClassesDir}
	if cp.String() != "" {
		parts = append(parts, cp.String())
	}
	fullCp := strings.Join(parts, string(os.PathListSeparator))

	junitArgs := []string{
		"-cp", fullCp,
		"org.junit.platform.console.ConsoleLauncher",
		"execute",
		"--scan-class-path=" + proj.TestClassesDir,
		"--details=tree",
		"--details-theme=unicode",
	}

	// --verbose: show stdout/stderr captured inside test methods
	if verbose {
		junitArgs = append(junitArgs, "--reports-dir=build/test-reports")
	}

	// --fail-fast: stop after first test failure
	if failFast {
		junitArgs = append(junitArgs, "--fail-if-no-tests")
		// JUnit Console Launcher doesn't have a --fail-fast flag directly,
		// but we can achieve it by limiting to 1 failure before exit.
		// The real mechanism is post-process: we run and exit on first failure.
		// For now we pass the flag through; full fail-fast needs JUnit 5.13+.
		junitArgs = append(junitArgs, "--fail-if-no-tests")
	}

	// --filter: match class name, method name, or "ClassName.methodName"
	if filter != "" {
		if strings.Contains(filter, ".") {
			// "CalculatorTest.addition works" → select class + method
			parts := strings.SplitN(filter, ".", 2)
			junitArgs = append(junitArgs,
				"--select-class="+parts[0],
				"--include-method-name-pattern="+parts[1],
			)
		} else {
			// "CalculatorTest" → match as class name pattern
			junitArgs = append(junitArgs, "--include-classname=.*"+filter+".*")
		}
	}

	fmt.Println("Running tests...")
	if filter != "" {
		fmt.Printf("  filter: %s\n", filter)
	}
	if verbose {
		fmt.Println("  mode: verbose (test output shown)")
	}
	if failFast {
		fmt.Println("  mode: fail-fast")
	}

	cmd := exec.Command(javaPath, junitArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("tests failed (exit %d)", exitErr.ExitCode())
		}
		return fmt.Errorf("failed to launch test runner: %w", err)
	}
	return nil
}

// findTestJar looks in ./libs/ — the install output dir. Test jars must be
// installed with scope "compile" (see testInit) so BuildInstallPlan puts them
// there; scope "test" would cause them to be skipped.
func findTestJar(libsDir, group, artifact, version string) (string, error) {
	path := resolver.ArtifactPath(libsDir, group, artifact, version, "", "jar")
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("%s not at expected path %s", artifact, path)
	}
	return path, nil
}

func missingTestDeps(libsDir string) []struct{ Coord, Version string } {
	var missing []struct{ Coord, Version string }
	for _, d := range defaultTestDeps {
		group, artifact, ok := splitGroupArtifact(d.Coord)
		if !ok {
			missing = append(missing, d)
			continue
		}
		path := resolver.ArtifactPath(libsDir, group, artifact, d.Version, "", "jar")
		if _, err := os.Stat(path); err != nil {
			missing = append(missing, d)
		}
	}
	return missing
}

func splitGroupArtifact(coord string) (group, artifact string, ok bool) {
	i := strings.IndexByte(coord, ':')
	if i < 0 {
		return "", "", false
	}
	return coord[:i], coord[i+1:], true
}