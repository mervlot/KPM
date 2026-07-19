package cli

import (
	"fmt"
	"os"

	"kpm/internal/run/executor"
	"kpm/internal/run/parser"
)

const runFile = "kpm.run"

// cmdRun implements `kpm run <task>`: parse kpm.run once, then execute the
// named task via internal/run/executor. Built-ins ("@name" steps) are
// wired here to KPM's own existing commands — the executor itself has no
// idea what "@install" means, it just calls whatever function this map
// gives it.
//
// Only builtins that map to something KPM actually does are registered.
// A kpm.run written against a hypothetical future builtin (e.g. "@compile",
// "@jar", "@publish" — KPM doesn't build or publish Kotlin yet) will fail
// with a clear "unknown built-in command" error rather than silently doing
// nothing, which would be far more confusing.
func cmdRun(args []string) int {
	if len(args) == 0 {
		return listRunTasks()
	}
	taskName := args[0]

	data, err := os.ReadFile(runFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fail(fmt.Errorf("no %s found in this directory — create one to define tasks (see the project README)", runFile))
		}
		return fail(err)
	}

	file, err := parser.Parse(data)
	if err != nil {
		return fail(err)
	}
	if _, ok := file.Task(taskName); !ok {
		return fail(fmt.Errorf("no task named %q in %s", taskName, runFile))
	}

	exec := executor.New(file, builtinRegistry())
	if err := exec.Run(taskName); err != nil {
		return fail(err)
	}
	return 0
}

func listRunTasks() int {
	data, err := os.ReadFile(runFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("no %s found in this directory\n", runFile)
			return 1
		}
		return fail(err)
	}
	file, err := parser.Parse(data)
	if err != nil {
		return fail(err)
	}
	if len(file.Tasks) == 0 {
		fmt.Println("no tasks defined in", runFile)
		return 0
	}
	fmt.Println("usage: kpm run <task>\n\navailable tasks:")
	for _, t := range file.Tasks {
		fmt.Println(" ", t.Name)
	}
	return 0
}

// builtinRegistry wraps KPM's existing int-returning command functions as
// executor.BuiltinFunc (error-returning), so @-steps in kpm.run reuse the
// exact same code path as running `kpm <command>` directly — no
// reimplementation, no drift between the two.
func builtinRegistry() map[string]executor.BuiltinFunc {
	toErr := func(code int) error {
		if code != 0 {
			return fmt.Errorf("command exited with status %d", code)
		}
		return nil
	}
	return map[string]executor.BuiltinFunc{
		"resolve":        func(args []string) error { return toErr(cmdGraph(nil, false)) }, // resolves + prints the tree; no separate resolve-only command exists yet
		"install":        func(args []string) error { return toErr(cmdInstall(args, false, false)) },
		"sync":           func(args []string) error { return toErr(cmdInstall(args, false, true)) },
		"update":         func(args []string) error { return toErr(cmdUpdate(args, false)) },
		"build":          func(args []string) error { return toErr(cmdBuild(args, false)) },
		"doctor":         func(args []string) error { return toErr(cmdDoctor(args)) },
		"clean":          func(args []string) error { return toErr(cmdClean(args)) },
		"outdated":       func(args []string) error { return toErr(cmdOutdated(args, false)) },
		"cache-clean":    func(args []string) error { return toErr(cmdCache(append([]string{"clean"}, args...))) },
		"why":            func(args []string) error { return toErr(cmdWhy(args, false)) },
		"compile":        func(args []string) error { return toErr(cmdCompile(args, modeAuto)) },
		"run":            func(args []string) error { return toErr(cmdRunApp(args)) },
		"compile-java":   func(args []string) error { return toErr(cmdCompile(args, modeJava)) },
		"compile-kotlin": func(args []string) error { return toErr(cmdCompile(args, modeKotlin)) },
		"jar":            func(args []string) error { return toErr(cmdJar()) },
	}
}
