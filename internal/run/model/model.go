// Package model defines the parsed representation of a kpm.run file: an
// ordered list of tasks, each a sequence of steps. It intentionally contains
// no execution logic — parsing and running are separate concerns (see
// internal/run/parser and internal/run/executor) so the graph can be built
// once and reused for every nested task invocation without re-parsing.
package model

// StepKind distinguishes how a step's text should be interpreted. The
// distinction between "invoke another task" and "run a shell command" is
// NOT made here — a bare line is ambiguous until execution time, when the
// executor checks it against the set of known task names (see the package
// doc on executor for why). This mirrors the spec: "If a line matches
// another task name exactly, execute that task. Otherwise execute the line
// as an external shell command."
type StepKind int

const (
	// StepCommand is a bare line: resolved at execution time to either a
	// task invocation (if it matches a declared task name) or a shell
	// command (otherwise).
	StepCommand StepKind = iota
	// StepBuiltin is a line beginning with "@": always a built-in KPM
	// action, never a task or shell command, regardless of naming overlap.
	StepBuiltin
)

// Step is one line within a task body.
type Step struct {
	Kind StepKind
	// Raw is the step's text with the leading "@" (for builtins) already
	// stripped and whitespace trimmed. For StepCommand this is either a
	// task name or a full shell command line. For StepBuiltin this is the
	// builtin name plus any space-separated arguments.
	Raw string
	// Line is the 1-indexed source line number, kept for error messages.
	Line int
}

// Name returns the builtin/task name portion of Raw (the first
// whitespace-delimited token); Args returns everything after it.
func (s Step) Name() string {
	name, _ := splitFirst(s.Raw)
	return name
}

func (s Step) Args() []string {
	_, rest := splitFirst(s.Raw)
	if rest == "" {
		return nil
	}
	return splitFields(rest)
}

func splitFirst(s string) (first, rest string) {
	i := 0
	for i < len(s) && s[i] != ' ' && s[i] != '\t' {
		i++
	}
	first = s[:i]
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	rest = s[i:]
	return
}

func splitFields(s string) []string {
	var fields []string
	start := -1
	for i, r := range s {
		if r == ' ' || r == '\t' {
			if start >= 0 {
				fields = append(fields, s[start:i])
				start = -1
			}
		} else if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		fields = append(fields, s[start:])
	}
	return fields
}

// Task is one named, ordered sequence of steps.
type Task struct {
	Name  string
	Steps []Step
	// Line is the 1-indexed source line of the "name:" declaration.
	Line int
}

// File is a fully parsed kpm.run: an ordered task list plus a name index.
// Order is preserved (e.g. for a future `kpm run` with no argument that
// might run the first-declared task) even though lookups go through the
// map.
type File struct {
	Tasks  []*Task
	byName map[string]*Task
}

func NewFile() *File {
	return &File{byName: map[string]*Task{}}
}

// Add appends a task, returning false if a task with this name already exists
// (the parser is responsible for turning that into a DuplicateTaskError).
func (f *File) Add(t *Task) bool {
	if _, exists := f.byName[t.Name]; exists {
		return false
	}
	if f.byName == nil {
		f.byName = map[string]*Task{}
	}
	f.byName[t.Name] = t
	f.Tasks = append(f.Tasks, t)
	return true
}

// Task looks up a task by name.
func (f *File) Task(name string) (*Task, bool) {
	t, ok := f.byName[name]
	return t, ok
}