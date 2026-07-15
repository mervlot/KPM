// Package executor runs a parsed kpm.run task graph (internal/run/model).
//
// The task-vs-shell-command ambiguity lives here rather than in the parser
// on purpose: whether a bare line like "build" means "run the task named
// build" or "run the shell command `build`" depends on the full set of
// declared task names, which only exists after the whole file is parsed —
// and per the spec, an exact task-name match always wins, so this has to be
// checked against the live graph at the moment the step runs, not baked in
// at parse time.
//
// Recursion detection: runTask tracks the current call stack as an ordered
// slice of task names. Before running a task, it checks whether that name
// is already on the stack; if so, that's a cycle (direct self-reference or
// an indirect a -> b -> a chain), and it returns a RecursionError carrying
// the full path for a precise error message rather than a stack overflow.
// This check happens on every invocation (not a global "visited" set), so
// the SAME task can still legitimately run multiple times from different,
// non-overlapping branches in one execution — only an actual cycle is
// rejected.
package executor

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"kpm/internal/run/model"
)

// BuiltinFunc implements one "@name" built-in action. Registered from
// outside (internal/cli), so the executor has zero direct dependency on
// what KPM's actual commands do — it only knows how to look one up and
// call it.
type BuiltinFunc func(args []string) error

// RecursionError reports a task cycle with the full call path, e.g.
// "build -> test -> build".
type RecursionError struct {
	Path []string
}

func (e *RecursionError) Error() string {
	return fmt.Sprintf("recursive task invocation: %s", strings.Join(e.Path, " -> "))
}

// UnknownBuiltinError is returned when a step references "@something" with
// no registered builtin of that name.
type UnknownBuiltinError struct {
	Name string
	Line int
}

func (e *UnknownBuiltinError) Error() string {
	return fmt.Sprintf("kpm.run:%d: unknown built-in command \"@%s\"", e.Line, e.Name)
}

// ShellError wraps a failed external command with the task/line that ran it.
type ShellError struct {
	Task    string
	Line    int
	Command string
	Err     error
}

func (e *ShellError) Error() string {
	return fmt.Sprintf("kpm.run:%d: command failed in task %q: %s: %v", e.Line, e.Task, e.Command, e.Err)
}
func (e *ShellError) Unwrap() error { return e.Err }

// Executor runs tasks from a single parsed *model.File. Constructed once
// per `kpm run` invocation and reused for every nested task call, so the
// file is never re-parsed mid-run.
type Executor struct {
	file     *model.File
	builtins map[string]BuiltinFunc
	// Stdout/Stderr, if set, redirect shell command output; default to the
	// process's own streams. Exposed so tests don't spam test output.
	Stdout, Stderr *os.File
}

// New constructs an Executor for a parsed file with the given builtin
// registry. A missing entry in builtins for a referenced "@name" step
// surfaces as UnknownBuiltinError, not a panic.
func New(file *model.File, builtins map[string]BuiltinFunc) *Executor {
	if builtins == nil {
		builtins = map[string]BuiltinFunc{}
	}
	return &Executor{file: file, builtins: builtins}
}

// Run starts execution at the named task with an empty call stack.
func (e *Executor) Run(taskName string) error {
	return e.runTask(taskName, nil)
}

func (e *Executor) runTask(name string, stack []string) error {
	for _, s := range stack {
		if s == name {
			return &RecursionError{Path: append(append([]string{}, stack...), name)}
		}
	}
	task, ok := e.file.Task(name)
	if !ok {
		return fmt.Errorf("no task named %q in kpm.run", name)
	}
	stack = append(stack, name)

	for _, step := range task.Steps {
		switch step.Kind {
		case model.StepBuiltin:
			fn, ok := e.builtins[step.Name()]
			if !ok {
				return &UnknownBuiltinError{Name: step.Name(), Line: step.Line}
			}
			if err := fn(step.Args()); err != nil {
				return fmt.Errorf("kpm.run:%d: builtin \"@%s\" failed: %w", step.Line, step.Name(), err)
			}

		case model.StepCommand:
			if _, isTask := e.file.Task(step.Raw); isTask {
				if err := e.runTask(step.Raw, stack); err != nil {
					return err
				}
				continue
			}
			if err := e.runShell(task.Name, step); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *Executor) runShell(taskName string, step model.Step) error {
	shell, flag := "sh", "-c"
	if runtime.GOOS == "windows" {
		shell, flag = "cmd", "/C"
	}
	cmd := exec.Command(shell, flag, step.Raw)
	cmd.Stdout = firstNonNil(e.Stdout, os.Stdout)
	cmd.Stderr = firstNonNil(e.Stderr, os.Stderr)
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return &ShellError{Task: taskName, Line: step.Line, Command: step.Raw, Err: err}
	}
	return nil
}

func firstNonNil(f *os.File, fallback *os.File) *os.File {
	if f != nil {
		return f
	}
	return fallback
}