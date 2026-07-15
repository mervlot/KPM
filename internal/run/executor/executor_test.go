package executor

import (
	"errors"
	"testing"

	"kpm/internal/run/model"
)

// buildFile constructs a *model.File directly (bypassing the parser) so
// these tests exercise only executor behavior.
func buildFile(tasks ...*model.Task) *model.File {
	f := model.NewFile()
	for _, t := range tasks {
		f.Add(t)
	}
	return f
}

func TestDirectSelfRecursion(t *testing.T) {
	file := buildFile(&model.Task{
		Name:  "loop",
		Steps: []model.Step{{Kind: model.StepCommand, Raw: "loop"}},
	})
	exec := New(file, nil)
	err := exec.Run("loop")
	if err == nil {
		t.Fatal("expected recursion error, got nil")
	}
	var rec *RecursionError
	if !errors.As(err, &rec) {
		t.Fatalf("expected *RecursionError, got %T: %v", err, err)
	}
	if rec.Path[0] != "loop" || rec.Path[len(rec.Path)-1] != "loop" {
		t.Errorf("expected path to start and end at 'loop', got %v", rec.Path)
	}
}

func TestIndirectRecursion(t *testing.T) {
	file := buildFile(
		&model.Task{Name: "a", Steps: []model.Step{{Kind: model.StepCommand, Raw: "b"}}},
		&model.Task{Name: "b", Steps: []model.Step{{Kind: model.StepCommand, Raw: "a"}}},
	)
	exec := New(file, nil)
	err := exec.Run("a")
	var rec *RecursionError
	if !errors.As(err, &rec) {
		t.Fatalf("expected *RecursionError for a->b->a cycle, got %T: %v", err, err)
	}
}

func TestSameTaskTwiceNonCyclicIsAllowed(t *testing.T) {
	// "root" calls "shared" twice via two different steps — not a cycle,
	// since "shared" never (directly or indirectly) calls "root" or itself.
	var calls []string
	file := buildFile(
		&model.Task{Name: "root", Steps: []model.Step{
			{Kind: model.StepCommand, Raw: "shared"},
			{Kind: model.StepCommand, Raw: "shared"},
		}},
		&model.Task{Name: "shared", Steps: []model.Step{
			{Kind: model.StepBuiltin, Raw: "mark"},
		}},
	)
	exec := New(file, map[string]BuiltinFunc{
		"mark": func(args []string) error { calls = append(calls, "shared-ran"); return nil },
	})
	if err := exec.Run("root"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(calls) != 2 {
		t.Fatalf("expected 'shared' to run twice, got %d", len(calls))
	}
}

func TestUnknownBuiltin(t *testing.T) {
	file := buildFile(&model.Task{
		Name:  "build",
		Steps: []model.Step{{Kind: model.StepBuiltin, Raw: "nonexistent"}},
	})
	exec := New(file, map[string]BuiltinFunc{})
	err := exec.Run("build")
	var unk *UnknownBuiltinError
	if !errors.As(err, &unk) {
		t.Fatalf("expected *UnknownBuiltinError, got %T: %v", err, err)
	}
	if unk.Name != "nonexistent" {
		t.Errorf("expected Name = nonexistent, got %q", unk.Name)
	}
}

func TestBuiltinReceivesArgs(t *testing.T) {
	var gotArgs []string
	file := buildFile(&model.Task{
		Name:  "build",
		Steps: []model.Step{{Kind: model.StepBuiltin, Raw: "why some-lib"}},
	})
	exec := New(file, map[string]BuiltinFunc{
		"why": func(args []string) error { gotArgs = args; return nil },
	})
	if err := exec.Run("build"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gotArgs) != 1 || gotArgs[0] != "some-lib" {
		t.Errorf("expected args [some-lib], got %v", gotArgs)
	}
}

func TestExecutionOrderPreserved(t *testing.T) {
	var order []string
	file := buildFile(&model.Task{
		Name: "build",
		Steps: []model.Step{
			{Kind: model.StepBuiltin, Raw: "one"},
			{Kind: model.StepBuiltin, Raw: "two"},
			{Kind: model.StepBuiltin, Raw: "three"},
		},
	})
	exec := New(file, map[string]BuiltinFunc{
		"one":   func(args []string) error { order = append(order, "one"); return nil },
		"two":   func(args []string) error { order = append(order, "two"); return nil },
		"three": func(args []string) error { order = append(order, "three"); return nil },
	})
	if err := exec.Run("build"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"one", "two", "three"}
	for i, w := range want {
		if order[i] != w {
			t.Errorf("order[%d] = %q, want %q (got %v)", i, order[i], w, order)
		}
	}
}

func TestBareLineMatchingTaskNameInvokesTask(t *testing.T) {
	var ran bool
	file := buildFile(
		&model.Task{Name: "outer", Steps: []model.Step{{Kind: model.StepCommand, Raw: "inner"}}},
		&model.Task{Name: "inner", Steps: []model.Step{{Kind: model.StepBuiltin, Raw: "mark"}}},
	)
	exec := New(file, map[string]BuiltinFunc{
		"mark": func(args []string) error { ran = true; return nil },
	})
	if err := exec.Run("outer"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ran {
		t.Error("expected 'outer' invoking bare line 'inner' to run the inner task")
	}
}

func TestStopsAtFirstFailure(t *testing.T) {
	var order []string
	file := buildFile(&model.Task{
		Name: "build",
		Steps: []model.Step{
			{Kind: model.StepBuiltin, Raw: "one"},
			{Kind: model.StepBuiltin, Raw: "fails"},
			{Kind: model.StepBuiltin, Raw: "three"},
		},
	})
	exec := New(file, map[string]BuiltinFunc{
		"one":   func(args []string) error { order = append(order, "one"); return nil },
		"fails": func(args []string) error { return errBoom },
		"three": func(args []string) error { order = append(order, "three"); return nil },
	})
	if err := exec.Run("build"); err == nil {
		t.Fatal("expected error from failing step")
	}
	if len(order) != 1 {
		t.Errorf("expected execution to stop after the failing step, got order=%v", order)
	}
}

var errBoom = errors.New("boom")