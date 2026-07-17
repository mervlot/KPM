package compiler

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMissingCompilerErrorMessage(t *testing.T) {
	e := &MissingCompilerError{Name: "javac", HomeVar: "JAVA_HOME", SearchedPaths: []string{"PATH", "JAVA_HOME (not set)"}}
	msg := e.Error()
	if !contains(msg, "javac") || !contains(msg, "JAVA_HOME") {
		t.Errorf("unexpected message: %q", msg)
	}
}

func TestCompileErrorIncludesOutput(t *testing.T) {
	e := &CompileError{Compiler: "javac", Output: "Hello.java:3: error: ';' expected", Err: os.ErrInvalid}
	msg := e.Error()
	if !contains(msg, "';' expected") {
		t.Errorf("expected compiler output in error message, got: %q", msg)
	}
}

func TestEmptySourceErrorMessages(t *testing.T) {
	noSourceAtAll := &EmptySourceError{}
	if !contains(noSourceAtAll.Error(), "no Java or Kotlin source") {
		t.Errorf("unexpected message: %q", noSourceAtAll.Error())
	}

	noDirAtAll := &EmptySourceError{Language: "Java"}
	if !contains(noDirAtAll.Error(), "no Java source directory") {
		t.Errorf("unexpected message: %q", noDirAtAll.Error())
	}

	dirExistsButEmpty := &EmptySourceError{Language: "Kotlin", Dir: "src/main/kotlin"}
	msg := dirExistsButEmpty.Error()
	if !contains(msg, "src/main/kotlin") || !contains(msg, ".kotlin") {
		t.Errorf("unexpected message: %q", msg)
	}
}

func TestJavaCompileEndToEnd(t *testing.T) {
	jc := JavaCompiler{}
	if _, err := jc.Locate(); err != nil {
		t.Skip("javac not available in this environment, skipping end-to-end compile test:", err)
	}

	dir := t.TempDir()
	src := filepath.Join(dir, "Hello.java")
	os.WriteFile(src, []byte("public class Hello { public static void main(String[] a) {} }"), 0o644)
	out := filepath.Join(dir, "out")
	os.MkdirAll(out, 0o755)

	err := jc.Compile(Input{JavaSources: []string{src}, OutDir: out})
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "Hello.class")); err != nil {
		t.Errorf("expected Hello.class to exist: %v", err)
	}
}

func TestJavaCompileErrorSurfacesOutput(t *testing.T) {
	jc := JavaCompiler{}
	if _, err := jc.Locate(); err != nil {
		t.Skip("javac not available, skipping:", err)
	}

	dir := t.TempDir()
	src := filepath.Join(dir, "Bad.java")
	os.WriteFile(src, []byte("this is not valid java"), 0o644)
	out := filepath.Join(dir, "out")
	os.MkdirAll(out, 0o755)

	err := jc.Compile(Input{JavaSources: []string{src}, OutDir: out})
	if err == nil {
		t.Fatal("expected a compile error for invalid source")
	}
	ce, ok := err.(*CompileError)
	if !ok {
		t.Fatalf("expected *CompileError, got %T: %v", err, err)
	}
	if ce.Output == "" {
		t.Error("expected compiler output to be captured")
	}
}

func TestKotlinCompileEndToEnd(t *testing.T) {
	kc := KotlinCompiler{}
	if _, err := kc.Locate(); err != nil {
		t.Skip("kotlinc not available in this environment, skipping:", err)
	}

	dir := t.TempDir()
	src := filepath.Join(dir, "Hello.kt")
	os.WriteFile(src, []byte("fun main() {}"), 0o644)
	out := filepath.Join(dir, "out")
	os.MkdirAll(out, 0o755)

	err := kc.Compile(Input{KotlinSources: []string{src}, OutDir: out})
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "HelloKt.class")); err != nil {
		t.Errorf("expected HelloKt.class to exist: %v", err)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}