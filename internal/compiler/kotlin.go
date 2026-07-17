package compiler

import (
	"bytes"
	"os/exec"
)

// KotlinCompiler shells out to kotlinc. Discovery order: PATH, then
// $KOTLIN_HOME/bin/kotlinc.
type KotlinCompiler struct{}

func (KotlinCompiler) Name() string { return "kotlinc" }

func (KotlinCompiler) Locate() (string, error) {
	return locate("kotlinc", "KOTLIN_HOME")
}

// Compile runs kotlinc against in.KotlinSources. If in.JavaSources is also
// set, those .java files are passed to kotlinc too — NOT to be compiled by
// it (kotlinc only ever emits .class files for Kotlin sources), but so
// kotlinc can resolve symbols a Kotlin file references from Java source in
// the same module (kotlinc's standard mechanism for Kotlin-calls-Java in a
// mixed module; see CompilerManager for why this direction runs first).
func (c KotlinCompiler) Compile(in Input) error {
	if len(in.KotlinSources) == 0 {
		return &EmptySourceError{Language: "Kotlin"}
	}

	kotlinc, err := c.Locate()
	if err != nil {
		return err
	}

	args := []string{"-d", in.OutDir}
	if in.Classpath != "" {
		args = append(args, "-cp", in.Classpath)
	}
	args = append(args, in.KotlinSources...)
	args = append(args, in.JavaSources...) // for cross-reference resolution only, see doc comment

	argFile, cleanup, err := writeArgFile(args)
	if err != nil {
		return err
	}
	defer cleanup()

	cmd := exec.Command(kotlinc, "@"+argFile)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return &CompileError{Compiler: "kotlinc", Output: out.String(), Err: err}
	}
	return nil
}

var _ Compiler = KotlinCompiler{}