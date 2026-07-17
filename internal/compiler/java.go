package compiler

import (
	"bytes"
	"os/exec"
)

// JavaCompiler shells out to javac. Discovery order: PATH, then
// $JAVA_HOME/bin/javac.
type JavaCompiler struct{}

func (JavaCompiler) Name() string { return "javac" }

func (JavaCompiler) Locate() (string, error) {
	return locate("javac", "JAVA_HOME")
}

// Compile runs javac against in.JavaSources only — Kotlin sources in Input
// are ignored here (mixed-language orchestration lives in CompilerManager,
// which decides whether javac needs the Kotlin output dir appended to
// in.Classpath before calling this).
func (c JavaCompiler) Compile(in Input) error {
	if len(in.JavaSources) == 0 {
		return &EmptySourceError{Language: "Java"}
	}

	javac, err := c.Locate()
	if err != nil {
		return err
	}

	args := []string{"-d", in.OutDir}
	if in.Classpath != "" {
		args = append(args, "-cp", in.Classpath)
	}
	args = append(args, in.JavaSources...)

	argFile, cleanup, err := writeArgFile(args)
	if err != nil {
		return err
	}
	defer cleanup()

	cmd := exec.Command(javac, "@"+argFile)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return &CompileError{Compiler: "javac", Output: out.String(), Err: err}
	}
	return nil
}

var _ Compiler = JavaCompiler{}