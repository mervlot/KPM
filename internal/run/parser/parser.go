// Package parser reads a kpm.run file and builds the task graph
// (internal/run/model), validating syntax and catching duplicate task
// names. It performs no execution and no task-vs-shell-command
// resolution — that ambiguity is inherently an execution-time question
// (see internal/run/executor), since it depends on the full set of
// declared task names, which the parser only assembles as a side effect.
//
// Grammar (deliberately tiny — this is not a general-purpose scripting
// language, per design constraint):
//
//	file       := (blank | comment | task)*
//	task       := taskheader step*
//	taskheader := IDENT ":"                 (column 0, no leading whitespace)
//	step       := WHITESPACE+ text          (any indentation counts, tabs or spaces)
//	comment    := WHITESPACE* "#" ...       (ignored entirely)
//	blank      := WHITESPACE*                (ignored entirely)
package parser

import (
	"fmt"
	"strings"

	"kpm/internal/run/model"
)

// ParseError is returned for any syntax problem; Line is 1-indexed.
type ParseError struct {
	Line   int
	Reason string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("kpm.run:%d: %s", e.Line, e.Reason)
}

// DuplicateTaskError is returned when the same task name is declared twice.
type DuplicateTaskError struct {
	Name       string
	FirstLine  int
	SecondLine int
}

func (e *DuplicateTaskError) Error() string {
	return fmt.Sprintf("kpm.run:%d: task %q is already defined at line %d — task names must be unique",
		e.SecondLine, e.Name, e.FirstLine)
}

// Parse builds a *model.File from raw kpm.run source.
func Parse(src []byte) (*model.File, error) {
	file := model.NewFile()
	firstLineOf := map[string]int{}

	var current *model.Task
	lines := strings.Split(string(src), "\n")

	for i, rawLine := range lines {
		lineNo := i + 1
		line := strings.TrimRight(rawLine, "\r")

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue // blank line
		}
		if strings.HasPrefix(trimmed, "#") {
			continue // full-line comment
		}

		indented := len(line) > 0 && (line[0] == ' ' || line[0] == '\t')

		if !indented {
			// Must be a task header: "name:"
			if !strings.HasSuffix(trimmed, ":") {
				return nil, &ParseError{Line: lineNo, Reason: fmt.Sprintf(
					"expected a task header ending in \":\" or an indented step, got %q", trimmed)}
			}
			name := strings.TrimSpace(strings.TrimSuffix(trimmed, ":"))
			if err := validateTaskName(name, lineNo); err != nil {
				return nil, err
			}

			t := &model.Task{Name: name, Line: lineNo}
			if !file.Add(t) {
				return nil, &DuplicateTaskError{Name: name, FirstLine: firstLineOf[name], SecondLine: lineNo}
			}
			firstLineOf[name] = lineNo
			current = t
			continue
		}

		// Indented line: a step belonging to the current task.
		if current == nil {
			return nil, &ParseError{Line: lineNo, Reason: "indented step found before any task declaration"}
		}
		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, "@") {
			name := strings.TrimSpace(strings.TrimPrefix(trimmed, "@"))
			if name == "" {
				return nil, &ParseError{Line: lineNo, Reason: "\"@\" with no builtin command name"}
			}
			current.Steps = append(current.Steps, model.Step{Kind: model.StepBuiltin, Raw: name, Line: lineNo})
		} else {
			current.Steps = append(current.Steps, model.Step{Kind: model.StepCommand, Raw: trimmed, Line: lineNo})
		}
	}

	return file, nil
}

// validateTaskName rejects task names that would be ambiguous or unsafe:
// empty, containing whitespace (would break the bare-word task-vs-shell
// resolution at execution time), or containing "@" or ":" (reserved
// syntax characters).
func validateTaskName(name string, line int) error {
	if name == "" {
		return &ParseError{Line: line, Reason: "empty task name"}
	}
	for _, r := range name {
		switch {
		case r == ' ' || r == '\t':
			return &ParseError{Line: line, Reason: fmt.Sprintf("task name %q must not contain whitespace", name)}
		case r == '@':
			return &ParseError{Line: line, Reason: fmt.Sprintf("task name %q must not contain \"@\" (reserved for builtins)", name)}
		case r == ':':
			return &ParseError{Line: line, Reason: fmt.Sprintf("task name %q must not contain \":\"", name)}
		}
	}
	return nil
}