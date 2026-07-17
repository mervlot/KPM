// Package cli wires the resolver, cache, repository, and lockfile packages
// together behind the `kpm` command-line interface.
package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"kpm/internal/cache"
	"kpm/internal/config"
	"kpm/internal/downloader"
	"kpm/internal/logger"
	"kpm/internal/repository"
	"kpm/internal/resolver"
	"kpm/internal/run/executor"
)

const Version = "0.2.0"

// Run dispatches argv (excluding the program name) to a subcommand. Returns
// the process exit code.
func Run(args []string) int {
	if len(args) == 0 {
		printHelp()
		return 0
	}

	offline := popFlag(&args, "--offline")
	if popFlag(&args, "--verbose") || popFlag(&args, "-v") {
		logger.Verbose = true
	}
	if popFlag(&args, "--slow") {
		downloader.SetRateLimit(1, 2500*time.Millisecond) // 1 request per 2.5s, hard cap
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "help", "-h", "--help":
		printHelp()
	case "version":
		fmt.Println("kpm version:", Version)

	case "init", "-i", "--init":
		return cmdInit(rest)
	case "add", "i", "install-add":
		return cmdAdd(rest, offline)
	case "remove", "rm":
		return cmdRemove(rest)
	case "update":
		return cmdUpdate(rest, offline)
	case "install", "get", "-g":
		return cmdInstall(rest, offline, false)
	case "sync":
		return cmdInstall(rest, offline, true)
	case "build":
		return cmdBuild(rest, offline)
	case "graph":
		return cmdGraph(rest, offline)
	case "why":
		return cmdWhy(rest, offline)
	case "outdated":
		return cmdOutdated(rest, offline)
	case "run":
		return cmdRun(rest)
	case "compile":
		return cmdCompile(rest, modeAuto)
	case "doctor":
		return cmdDoctor(rest)
	case "clean":
		return cmdClean(rest)
	case "cache":
		return cmdCache(rest)

	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printHelp()
		return 1
	}
	return 0
}

func popFlag(args *[]string, flag string) bool {
	out := (*args)[:0:0]
	found := false
	for _, a := range *args {
		if a == flag {
			found = true
			continue
		}
		out = append(out, a)
	}
	*args = out
	return found
}

func printHelp() {
	fmt.Print(`
kpm — a Maven-compatible dependency manager for Kotlin/JVM

Usage:
  kpm <command> [options]

Project:
  init                      Initialize a new package.kpm
  add <group:artifact[:version]>   Add a dependency and re-resolve
  remove <group:artifact>   Remove a dependency and re-resolve
  install                   Resolve + download everything in package.kpm (uses kpm.lock if present)
  sync                      Re-resolve ignoring kpm.lock and rewrite it
  update [group:artifact]   Update one or all dependencies to latest allowed versions
  build                     Resolve, install, and run the "build" script

Tasks (kpm.run):
  run <task>                Run a task defined in kpm.run
  run                       List available tasks

Build:
  compile                   Compile Java/Kotlin sources under src/main to build/classes
                            (mixed projects: kotlinc first, then javac — see internal/compiler docs)

Run (kpm.run builtin only, not a top-level command):
  @run <main-class> [args]  Launch a compiled class with build/classes + installed jars on -cp
                            (does not compile first — pair with @compile in the same task)

Inspection:
  graph                     Print the resolved dependency tree
  why <group:artifact>      Explain what pulled a dependency in
  outdated                  Show dependencies with newer versions available
  doctor                    Verify local cache integrity

Maintenance:
  clean                     Remove local build/libs output
  cache clean [--all]       Prune the local artifact cache

Flags:
  --offline                 Never hit the network; fail if the cache is incomplete
  --slow                    Space requests out further (~1 every 2.5s) if you're still seeing 429s
  --verbose                 Print debug diagnostics

  version, help             Show version / this message
`)
}

// setup loads package.kpm and constructs the shared resolver/fetcher stack
// every dependency-aware command needs.
func setup(offline bool) (*config.Manifest, *resolver.Resolver, *resolver.Fetcher, error) {
	manifest, err := config.Load(config.ManifestFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil, fmt.Errorf("no %s found in this directory — run `kpm init` first", config.ManifestFile)
		}
		return nil, nil, nil, err
	}

	repos := repository.BuildSet(manifest)
	c, err := cache.Open()
	if err != nil {
		return nil, nil, nil, err
	}
	http := downloader.New()

	// Constant feedback instead of going silent during network-bound work:
	// print a live line per resolved coordinate, and another whenever the
	// downloader is backing off after a 429/5xx rather than just hanging.
	progress := logger.NewProgress("resolving", 0)
	resolver.OnResolveStep = progress.Step
	downloader.OnRetry = progress.Retrying

	fetcher := resolver.NewFetcher(repos, c, http, offline)
	res := resolver.New(fetcher)
	return manifest, res, fetcher, nil
}

func diagnosticFor(err error) logger.Diagnostic {
	if d, ok := diagnosticForCompile(err); ok {
		return d
	}

	var rec *executor.RecursionError
	if errors.As(err, &rec) {
		return logger.Diagnostic{
			Title:  "Recursive task in kpm.run",
			Detail: err.Error(),
			Fixes: []string{
				"Check the task chain named in the error and remove the cycle",
			},
		}
	}
	var unkBuiltin *executor.UnknownBuiltinError
	if errors.As(err, &unkBuiltin) {
		return logger.Diagnostic{
			Title:  "Unknown built-in command in kpm.run",
			Detail: err.Error(),
			Fixes: []string{
				"Check the spelling of the \"@\" line",
				"Run `kpm help` to see which built-ins this version of kpm supports",
			},
		}
	}

	var notFound *resolver.NotFoundInAnyRepoError
	if errors.As(err, &notFound) {
		return logger.Diagnostic{
			Title:  fmt.Sprintf("%s not found", capitalize(notFound.Kind)),
			Detail: notFound.Error(),
			Fixes: []string{
				"Check the group/artifact name and version at https://search.maven.org",
				"If this is a private/internal library, make sure its repository is listed in package.kpm",
				"Version strings are case- and character-sensitive — e.g. \"3.2.3\" vs \"3.2.3.RELEASE\" are different artifacts",
			},
		}
	}

	var noConn *resolver.NoConnectionError
	if errors.As(err, &noConn) {
		return logger.Diagnostic{
			Title:  "No internet connection",
			Detail: fmt.Sprintf("couldn't reach the repository while fetching %s", noConn.Coordinate),
			Fixes: []string{
				"Check that you're connected to the internet",
				"If you're behind a VPN/proxy/firewall, make sure it allows access to repo1.maven.org",
				"Already downloaded this before? Try `kpm install --offline` to use what's cached locally",
			},
		}
	}

	msg := err.Error()
	switch {
	case contains(msg, "circular dependency"):
		return logger.Diagnostic{
			Title:  "Circular dependency detected",
			Detail: msg,
			Fixes: []string{
				"Run `kpm graph` to see the full dependency tree and locate the cycle",
				"Add an <exclusion> for one side of the cycle in package.kpm",
			},
		}
	case contains(msg, "checksum mismatch"):
		return logger.Diagnostic{
			Title:  "Checksum mismatch",
			Detail: msg,
			Fixes: []string{
				"Run `kpm cache clean` to remove the corrupted artifact, then retry",
				"Verify the repository URL in package.kpm is correct and trusted",
			},
		}
	case contains(msg, "offline mode"):
		return logger.Diagnostic{
			Title:  "Artifact not available offline",
			Detail: msg,
			Fixes: []string{
				"Run the command once without --offline to populate the cache",
			},
		}
	case contains(msg, "no %s found") || contains(msg, "package.kpm"):
		return logger.Diagnostic{
			Title:  "No project found",
			Detail: msg,
			Fixes:  []string{"Run `kpm init` in this directory"},
		}
	default:
		return logger.Diagnostic{Title: "Resolution failed", Detail: msg}
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}