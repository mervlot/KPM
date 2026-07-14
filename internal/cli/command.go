package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"kpm/internal/cache"
	"kpm/internal/config"
	"kpm/internal/downloader"
	"kpm/internal/lockfile"
	"kpm/internal/logger"
	"kpm/internal/repository"
	"kpm/internal/resolver"
)

const libsDir = "./libs"

// installConcurrency is deliberately low. The real protection against
// Maven Central's rate limiting is downloader.globalPace (a process-wide
// minimum spacing between requests, shared by every goroutine); this just
// bounds how many downloads can be in-flight/decompressing/writing to disk
// at once. A higher number here would NOT download faster — every one of
// them still has to wait its turn at the same global pacer — it would just
// mean more goroutines parked waiting on that lock at once.
const installConcurrency = 2

func fail(err error) int {
	diagnosticFor(err).Print()
	return 1
}

func cmdInit(args []string) int {
	if _, err := os.Stat(config.ManifestFile); err == nil {
		fmt.Println(config.ManifestFile, "already exists")
		return 1
	}
	name := "my-kpm-project"
	if len(args) > 0 {
		name = args[0]
	}
	m := config.New(name)
	if err := m.Save(config.ManifestFile); err != nil {
		return fail(err)
	}
	fmt.Println("Created", config.ManifestFile)
	return 0
}

// cmdAdd adds one or more "group:artifact[:version]" dependencies to
// package.kpm and re-resolves the whole project.
func cmdAdd(args []string, offline bool) int {
	if len(args) == 0 {
		fmt.Println("usage: kpm add <group:artifact[:version]> [more...]")
		return 1
	}
	manifest, err := config.Load(config.ManifestFile)
	if err != nil {
		return fail(fmt.Errorf("no %s found — run `kpm init` first", config.ManifestFile))
	}

	// Create a test manifest to validate resolution BEFORE saving to disk
	testManifest := *manifest
	testManifest.Dependencies = make(map[string]config.DependencySpec)
	for k, v := range manifest.Dependencies {
		testManifest.Dependencies[k] = v
	}

	for _, spec := range args {
		group, artifact, ver := splitAddSpec(spec)
		if group == "" || artifact == "" {
			fmt.Println("skipping invalid dependency spec:", spec, "(expected group:artifact[:version])")
			continue
		}
		testManifest.Dependencies[group+":"+artifact] = config.DependencySpec{Version: ver}
	}

	// CRITICAL FIX: Only save the manifest if resolution and installation succeed
	if code := runResolveAndInstallWithManifest(&testManifest, offline, true); code != 0 {
		return code // Failed, do NOT persist changes to package.kpm
	}

	manifest.Dependencies = testManifest.Dependencies
	if err := manifest.Save(config.ManifestFile); err != nil {
		return fail(err)
	}
	fmt.Println("Successfully added dependencies and updated", config.ManifestFile)
	return 0
}

func cmdRemove(args []string) int {
	if len(args) == 0 {
		fmt.Println("usage: kpm remove <group:artifact>")
		return 1
	}
	manifest, err := config.Load(config.ManifestFile)
	if err != nil {
		return fail(err)
	}

	// Create a test manifest to validate resolution BEFORE saving to disk
	testManifest := *manifest
	testManifest.Dependencies = make(map[string]config.DependencySpec)
	for k, v := range manifest.Dependencies {
		testManifest.Dependencies[k] = v
	}

	removed := false
	for _, name := range args {
		if _, ok := testManifest.Dependencies[name]; ok {
			delete(testManifest.Dependencies, name)
			removed = true
		} else {
			fmt.Println("not a direct dependency (nothing to remove):", name)
		}
	}
	if !removed {
		return 1
	}

	// CRITICAL FIX: Only save the manifest if resolution and installation succeed
	if code := runResolveAndInstallWithManifest(&testManifest, false, true); code != 0 {
		return code // Failed, do NOT persist changes to package.kpm
	}

	manifest.Dependencies = testManifest.Dependencies
	if err := manifest.Save(config.ManifestFile); err != nil {
		return fail(err)
	}
	fmt.Println("Successfully removed dependencies and updated", config.ManifestFile)
	return 0
}

// cmdUpdate clears the pinned version for one dependency (or all, if none
// named) so resolution picks up the latest version satisfying constraints.
func cmdUpdate(args []string, offline bool) int {
	manifest, err := config.Load(config.ManifestFile)
	if err != nil {
		return fail(err)
	}

	// Create a test manifest to validate resolution BEFORE saving to disk
	testManifest := *manifest
	testManifest.Dependencies = make(map[string]config.DependencySpec)
	for k, v := range manifest.Dependencies {
		testManifest.Dependencies[k] = v
	}

	targets := args
	if len(targets) == 0 {
		for name := range testManifest.Dependencies {
			targets = append(targets, name)
		}
	}

	for _, name := range targets {
		spec, ok := testManifest.Dependencies[name]
		if !ok {
			fmt.Println("not a direct dependency:", name)
			continue
		}
		spec.Version = "" // "" triggers metadata "latest" lookup during resolution
		testManifest.Dependencies[name] = spec
	}

	// CRITICAL FIX: Only save the manifest if resolution and installation succeed
	if code := runResolveAndInstallWithManifest(&testManifest, offline, true); code != 0 {
		return code // Failed, do NOT persist changes to package.kpm
	}

	manifest.Dependencies = testManifest.Dependencies
	if err := manifest.Save(config.ManifestFile); err != nil {
		return fail(err)
	}
	fmt.Println("Successfully updated dependencies and updated", config.ManifestFile)
	return 0
}

// cmdInstall resolves (reusing kpm.lock if present and untouched) and
// downloads every dependency. cmdInstall(sync=true) forces a fresh
// resolution and rewrites the lock file even if manifest looks unchanged.
func cmdInstall(args []string, offline, forceSync bool) int {
	return runResolveAndInstall(offline, forceSync)
}

func runResolveAndInstallWithManifest(manifest *config.Manifest, offline, forceSync bool) int {
	repos := repository.BuildSet(manifest)
	c, err := cache.Open()
	if err != nil {
		return fail(err)
	}
	http := downloader.New()

	progress := logger.NewProgress("🔍 resolving", 0)
	resolver.OnResolveStep = progress.Step
	downloader.OnRetry = progress.Retrying

	fetcher := resolver.NewFetcher(repos, c, http, offline)
	res := resolver.New(fetcher)

	_ = forceSync // kept for CLI symmetry; resolution is always freshly computed today,
	// matching "sync" semantics — a future revision can short-circuit to the
	// existing kpm.lock when the manifest hash is unchanged and forceSync is false.

	result, err := res.Resolve(manifest)
	if err != nil {
		return fail(err)
	}
	logger.FinishActiveProgress(fmt.Sprintf("✔ resolved %d coordinate(s)", len(result.Winners)))
	for _, w := range result.Warnings {
		logger.Warn("%s", w)
	}
	for _, c := range result.Conflicts {
		logger.Debug("conflict %s resolved to %s (candidates: %v)", c.Coordinate, c.Chosen, c.Candidates)
	}

	plans := resolver.BuildInstallPlan(result, libsDir)
	fmt.Printf("Resolved %d artifacts, downloading %d...\n", len(result.Winners), len(plans))

	installProgress := logger.NewProgress("⬇ downloading", len(plans))
	downloader.OnRetry = installProgress.Retrying
	if errs := res.InstallWithProgress(plans, installConcurrency, installProgress.Step); len(errs) > 0 {
		logger.FinishActiveProgress(fmt.Sprintf("✖ %d artifact(s) failed", len(errs)))
		for _, e := range errs {
			logger.Warn("%s", e)
		}
		return fail(fmt.Errorf("%d artifact(s) failed to install", len(errs)))
	}
	logger.FinishActiveProgress(fmt.Sprintf("✔ downloaded %d artifact(s)", len(plans)))

	lf := lockfile.FromResult(result, fetcher.RepositoryFor)
	if err := lf.Save(lockfile.FileName); err != nil {
		return fail(err)
	}
	fmt.Println("Wrote", lockfile.FileName)
	return 0
}

func runResolveAndInstall(offline, forceSync bool) int {
	manifest, res, fetcher, err := setup(offline)
	if err != nil {
		return fail(err)
	}

	_ = forceSync // kept for CLI symmetry; resolution is always freshly computed today,
	// matching "sync" semantics — a future revision can short-circuit to the
	// existing kpm.lock when the manifest hash is unchanged and forceSync is false.

	result, err := res.Resolve(manifest)
	if err != nil {
		return fail(err)
	}
	logger.FinishActiveProgress(fmt.Sprintf("✔ resolved %d coordinate(s)", len(result.Winners)))
	for _, w := range result.Warnings {
		logger.Warn("%s", w)
	}
	for _, c := range result.Conflicts {
		logger.Debug("conflict %s resolved to %s (candidates: %v)", c.Coordinate, c.Chosen, c.Candidates)
	}

	plans := resolver.BuildInstallPlan(result, libsDir)
	fmt.Printf("Resolved %d artifacts, downloading %d...\n", len(result.Winners), len(plans))

	installProgress := logger.NewProgress("⬇ downloading", len(plans))
	downloader.OnRetry = installProgress.Retrying
	if errs := res.InstallWithProgress(plans, installConcurrency, installProgress.Step); len(errs) > 0 {
		logger.FinishActiveProgress(fmt.Sprintf("✖ %d artifact(s) failed", len(errs)))
		for _, e := range errs {
			logger.Warn("%s", e)
		}
		return fail(fmt.Errorf("%d artifact(s) failed to install", len(errs)))
	}
	logger.FinishActiveProgress(fmt.Sprintf("✔ downloaded %d artifact(s)", len(plans)))

	lf := lockfile.FromResult(result, fetcher.RepositoryFor)
	if err := lf.Save(lockfile.FileName); err != nil {
		return fail(err)
	}
	fmt.Println("Wrote", lockfile.FileName)
	return 0
}

func cmdBuild(args []string, offline bool) int {
	if code := runResolveAndInstall(offline, false); code != 0 {
		return code
	}
	manifest, err := config.Load(config.ManifestFile)
	if err != nil {
		return fail(err)
	}
	script, ok := manifest.Scripts["build"]
	if !ok {
		fmt.Println("no \"build\" script defined in package.kpm scripts; install completed successfully")
		return 0
	}
	fmt.Println("running build script:", script)
	// Script execution intentionally kept simple/explicit (no shell
	// interpolation) to avoid the injection surface of exec'ing arbitrary
	// shell strings; a future revision can add a sandboxed script runner.
	fmt.Println("(script execution is not yet wired up in this build — run it manually):", script)
	return 0
}

func cmdGraph(args []string, offline bool) int {
	manifest, res, _, err := setup(offline)
	if err != nil {
		return fail(err)
	}
	result, err := res.Resolve(manifest)
	if err != nil {
		return fail(err)
	}
	order, err := result.Graph.TopoSort()
	if err != nil {
		return fail(err)
	}
	fmt.Println(manifest.Name + "@" + manifest.Version)
	for _, n := range order {
		marker := " "
		if w, ok := result.Winners[n.Coordinate.Key()]; ok && w.Version == n.Version {
			marker = "*"
		}
		fmt.Printf("%s%s %s:%s %s (%s)\n", strings.Repeat(" ", n.Depth), marker, n.Group, n.Artifact, n.Version, n.Scope)
	}
	fmt.Println("\n(* = winning version after conflict resolution)")
	return 0
}

func cmdWhy(args []string, offline bool) int {
	if len(args) == 0 {
		fmt.Println("usage: kpm why <group:artifact>")
		return 1
	}
	manifest, res, _, err := setup(offline)
	if err != nil {
		return fail(err)
	}
	result, err := res.Resolve(manifest)
	if err != nil {
		return fail(err)
	}
	target := args[0]
	found := false
	for _, n := range result.Graph.Nodes() {
		if n.Coordinate.String() != target {
			continue
		}
		found = true
		fmt.Printf("%s:%s@%s\n", n.Group, n.Artifact, n.Version)
		for _, path := range result.Graph.Why(n.Key()) {
			fmt.Println(" " + strings.Join(path, " -> "))
		}
	}
	if !found {
		fmt.Println(target, "is not in the resolved dependency graph")
		return 1
	}
	return 0
}

func cmdOutdated(args []string, offline bool) int {
	manifest, res, fetcher, err := setup(offline)
	if err != nil {
		return fail(err)
	}
	result, err := res.Resolve(manifest)
	if err != nil {
		return fail(err)
	}
	type row struct{ coord, current, latest string }
	var rows []row
	for _, n := range result.Winners {
		m, err := fetcher.GetMetadata(n.Group, n.Artifact)
		if err != nil || m.Versioning.Latest == "" {
			continue
		}
		if m.Versioning.Latest != n.Version {
			rows = append(rows, row{n.Coordinate.String(), n.Version, m.Versioning.Latest})
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].coord < rows[j].coord })
	if len(rows) == 0 {
		fmt.Println("everything is up to date")
		return 0
	}
	fmt.Printf("%-45s %-15s %-15s\n", "PACKAGE", "CURRENT", "LATEST")
	for _, r := range rows {
		fmt.Printf("%-45s %-15s %-15s\n", r.coord, r.current, r.latest)
	}
	return 0
}

func cmdDoctor(args []string) int {
	c, err := cache.Open()
	if err != nil {
		return fail(err)
	}
	problems, err := c.Doctor()
	if err != nil {
		return fail(err)
	}
	if _, err := os.Stat(config.ManifestFile); err != nil {
		fmt.Println("⚠ no", config.ManifestFile, "in current directory")
	}
	if len(problems) == 0 {
		fmt.Println("✓ local cache OK, no corrupted artifacts found")
		return 0
	}
	fmt.Println("Found", len(problems), "corrupted cache entries:")
	for _, p := range problems {
		fmt.Println(" -", p)
	}
	fmt.Println("\nrun `kpm cache clean` to remove them, then re-run `kpm install`")
	return 1
}

func cmdClean(args []string) int {
	if err := os.RemoveAll(libsDir); err != nil {
		return fail(err)
	}
	fmt.Println("removed", libsDir)
	return 0
}

func cmdCache(args []string) int {
	if len(args) == 0 || args[0] != "clean" {
		fmt.Println("usage: kpm cache clean [--all]")
		return 1
	}
	c, err := cache.Open()
	if err != nil {
		return fail(err)
	}
	maxAge := 30 * 24 * time.Hour
	for _, a := range args[1:] {
		if a == "--all" {
			maxAge = 0
		}
	}
	freed, n, err := c.Clean(maxAge)
	if err != nil {
		return fail(err)
	}
	fmt.Printf("removed %d file(s), freed %.2f MB\n", n, float64(freed)/1024/1024)
	return 0
}

func splitAddSpec(spec string) (group, artifact, version string) {
	parts := strings.Split(spec, ":")
	switch len(parts) {
	case 2:
		return parts[0], parts[1], ""
	case 3:
		return parts[0], parts[1], parts[2]
	default:
		return "", "", ""
	}
}