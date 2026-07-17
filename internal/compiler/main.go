// Package compiler implements KPM's compilation pipeline: locating javac
// and kotlinc, and invoking them against a project's sources and resolved
// classpath. This is the ONLY package that ever shells out to a compiler —
// internal/cli and internal/run/executor call into CompilerManager and
// never touch os/exec for javac/kotlinc themselves, so if KPM ever needs to
// change how compilation happens (a compiler daemon, incremental
// compilation, a different Kotlin compiler backend) there's exactly one
// place that changes.
package compiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
) // Input is everything a single compiler invocation needs. Both JavaCompiler
// and KotlinCompiler take the same shape, even though Kotlin's invocation
// additionally accepts Java sources for cross-reference resolution (see
// KotlinCompiler.Compile) — keeping the input shape uniform is what lets
// CompilerManager treat both compilers through one interface.
type Input struct {
	// JavaSources / KotlinSources are absolute or relative file paths.
	JavaSources   []string
	KotlinSources []string
	Classpath     string // pre-built via internal/classpath; compilers never resolve their own
	OutDir        string // class output directory
}

// Compiler is implemented by JavaCompiler and KotlinCompiler.
type Compiler interface {
	// Name identifies the compiler for error messages ("javac", "kotlinc").
	Name() string
	// Locate finds the compiler executable, checking PATH first, then the
	// relevant *_HOME environment variable. Returns MissingCompilerError
	// (not a bare error) on failure so callers can render a friendly
	// diagnostic without string-matching.
	Locate() (string, error)
	// Compile runs the compiler against Input. CompileError wraps any
	// non-zero exit with the compiler's own output attached.
	Compile(in Input) error
}

// MissingCompilerError is returned when a required compiler can't be found
// anywhere searched.
type MissingCompilerError struct {
	Name          string   // "javac" or "kotlinc"
	HomeVar       string   // "JAVA_HOME" or "KOTLIN_HOME"
	SearchedPaths []string // every location actually checked, for the error message
}

func (e *MissingCompilerError) Error() string {
	return fmt.Sprintf(
		"%s not found (checked: %s) — install it and make sure it's on PATH, or set %s to point at its installation",
		e.Name, strings.Join(e.SearchedPaths, ", "), e.HomeVar,
	)
}

// CompileError wraps a failed compiler invocation with its captured output,
// since a bare exit-status error ("exit status 1") tells the person
// nothing — the actual compiler diagnostics are what they need to see.
type CompileError struct {
	Compiler string
	Output   string
	Err      error
}

func (e *CompileError) Error() string {
	out := strings.TrimSpace(e.Output)
	if out == "" {
		return fmt.Sprintf("%s failed: %v", e.Compiler, e.Err)
	}
	return fmt.Sprintf("%s failed:\n%s", e.Compiler, out)
}
func (e *CompileError) Unwrap() error { return e.Err }

// EmptySourceError means compilation was requested but there is no matching
// source to compile — an existing-but-empty src/main/java, or the
// directory not existing at all, are both treated identically here (from
// the compiler's point of view, "nothing to compile" is the same problem
// either way).
type EmptySourceError struct {
	Language string // "Java", "Kotlin", or "" for "no source at all"
	Dir      string
}

func (e *EmptySourceError) Error() string {
	if e.Language == "" {
		return "no Java or Kotlin source found (expected sourceDir/main/java and/or sourceDir/main/kotlin)"
	}
	if e.Dir == "" {
		return fmt.Sprintf("no %s source directory found (expected src/main/%s)", e.Language, strings.ToLower(e.Language))
	}
	return fmt.Sprintf("%s source directory %s exists but contains no .%s files", e.Language, e.Dir, strings.ToLower(e.Language))
}

// LocateJavaRuntime finds the `java` launcher itself (as opposed to javac),
// using the same PATH -> $JAVA_HOME/bin search order. Exported for the CLI's
// "@run" builtin, which needs to launch compiled classes — running code is
// a different concern from compiling it, but discovery is the same, since
// `java` and `javac` live side by side in every JDK install.
func LocateJavaRuntime() (string, error) {
	return locate("java", "JAVA_HOME")
}

// locate implements the shared PATH -> $HOME/bin/<name> search order used
// by both JavaCompiler and KotlinCompiler. On macOS, if homeVar is
// JAVA_HOME and isn't set, it additionally tries `/usr/libexec/java_home`
// — the standard macOS way to find an installed JDK without relying on the
// user having set JAVA_HOME themselves (common with Homebrew/SDKMAN
// installs that put java_home-discoverable JDKs on disk but don't always
// export the env var, and don't always put javac on PATH either).
func locate(name, homeVar string) (string, error) {
	var searched []string

	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}
	searched = append(searched, "PATH")

	if home := os.Getenv(homeVar); home != "" {
		candidate := filepath.Join(home, "bin", binaryName(name))
		searched = append(searched, candidate)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	} else {
		searched = append(searched, homeVar+" (not set)")
		if runtime.GOOS == "darwin" && homeVar == "JAVA_HOME" {
			if home, err := macJavaHome(); err == nil && home != "" {
				candidate := filepath.Join(home, "bin", binaryName(name))
				searched = append(searched, candidate+" (via /usr/libexec/java_home)")
				if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
					return candidate, nil
				}
			}
		}
	}

	return "", &MissingCompilerError{Name: name, HomeVar: homeVar, SearchedPaths: searched}
}

// macJavaHome shells out to macOS's own JDK-discovery tool. Best-effort:
// any failure (tool missing, no JDK registered with it, non-macOS) just
// means this fallback contributes nothing, not an error — the caller
// already has PATH and JAVA_HOME as its primary search paths.
func macJavaHome() (string, error) {
	out, err := exec.Command("/usr/libexec/java_home").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// binaryName returns the filename a tool actually has on disk on the
// current OS. This isn't a uniform "+.exe on Windows" rule: javac and java
// are native executables (javac.exe, java.exe), but kotlinc ships as a
// batch script on Windows (kotlinc.bat) — its real distribution has no
// kotlinc.exe at all. Getting this wrong only breaks the $KOTLIN_HOME/bin
// fallback (PATH-based lookup via exec.LookPath already handles this
// correctly, since Windows' PATHEXT resolves ".bat" for a bare "kotlinc"
// automatically) — but someone relying solely on KOTLIN_HOME without it
// also being on PATH would otherwise get a false "not found".
func binaryName(name string) string {
	if runtime.GOOS != "windows" {
		return name
	}
	if name == "kotlinc" {
		return name + ".bat"
	}
	return name + ".exe"
}

// writeArgFile writes tokens (one full javac/kotlinc argument each) to a
// temp file in @argfile format and returns its path plus a cleanup func.
//
// Why this exists: on Windows, both javac and kotlinc are launched via
// batch scripts (javac.exe is native, but kotlinc is literally kotlinc.bat).
// A single "-cp" argument whose VALUE contains semicolons — which is
// exactly what a Windows classpath separator is — can get re-tokenized by
// the batch script's own argument forwarding into multiple separate
// arguments, so kotlinc sees each classpath entry as its own positional
// argument instead of one "-cp <value>" pair, and misinterprets the jar
// paths as source files ("source entry is not a Kotlin file"). This is not
// a Go os/exec bug or a quoting mistake on our part — it's a known
// characteristic of .bat argument handling that no amount of shell-level
// quoting from the calling process reliably works around, because the
// mis-tokenization happens INSIDE the batch script, after our process has
// already handed off a perfectly well-formed single argument.
//
// The fix used by real build tools (Gradle included) is to sidestep
// command-line argument passing for this entirely: write every argument to
// a file, one token per line, and invoke the compiler with a single
// "@path/to/argfile" argument. Both javac and kotlinc read this format
// directly and split on newlines themselves — no shell or batch script
// gets a chance to re-tokenize anything. This is applied on every
// platform (not just Windows) for consistency and because it also sidesteps
// Windows' ~8K character command-line length limit on large classpaths.
func writeArgFile(tokens []string) (path string, cleanup func(), err error) {
	f, err := os.CreateTemp("", "kpm-compile-args-*.txt")
	if err != nil {
		return "", nil, fmt.Errorf("creating compiler argfile: %w", err)
	}
	defer f.Close()

	for _, tok := range tokens {
		line := tok
		if strings.ContainsAny(tok, " \t") {
			line = `"` + strings.ReplaceAll(tok, `"`, `\"`) + `"`
		}
		if _, err := f.WriteString(line + "\n"); err != nil {
			os.Remove(f.Name())
			return "", nil, fmt.Errorf("writing compiler argfile: %w", err)
		}
	}

	return f.Name(), func() { os.Remove(f.Name()) }, nil
}