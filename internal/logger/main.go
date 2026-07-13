// Package logger provides simple leveled console output and a helper for
// formatting the "actionable diagnostics" the CLI shows on failure
// (what went wrong, why, and what to try next).
package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

var Verbose = false

func Info(format string, a ...any) { fmt.Fprintf(os.Stdout, format+"\n", a...) }
func Warn(format string, a ...any) { fmt.Fprintf(os.Stderr, "warning: "+format+"\n", a...) }
func Debug(format string, a ...any) {
	if Verbose {
		fmt.Fprintf(os.Stderr, "debug: "+format+"\n", a...)
	}
}

// Progress prints a single, continuously-updating status line (using \r)
// instead of scrolling output, so long resolutions/downloads give constant
// visual feedback without spamming the terminal. Safe for concurrent use.
type Progress struct {
	mu      sync.Mutex
	label   string
	done    int
	total   int
	started time.Time
}

var active struct {
	mu sync.Mutex
	p  *Progress
}

// FinishActiveProgress finishes whichever Progress was most recently created
// via NewProgress, printing summary. Safe to call even if nothing is active.
func FinishActiveProgress(summary string) {
	active.mu.Lock()
	p := active.p
	active.p = nil
	active.mu.Unlock()
	if p != nil {
		p.Finish(summary)
	} else {
		fmt.Println(summary)
	}
}

func NewProgress(label string, total int) *Progress {
	p := &Progress{label: label, total: total, started: time.Now()}
	active.mu.Lock()
	active.p = p
	active.mu.Unlock()
	return p
}

// Step reports one unit of work completed, with a short description of what
// just happened (e.g. an artifact coordinate), and redraws the status line.
func (p *Progress) Step(detail string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.done++
	elapsed := time.Since(p.started).Round(100 * time.Millisecond)
	var line string
	if p.total > 0 {
		line = fmt.Sprintf("\r%s [%d/%d] %s (%s)", p.label, p.done, p.total, truncate(detail, 55), elapsed)
	} else {
		line = fmt.Sprintf("\r%s [%d] %s (%s)", p.label, p.done, truncate(detail, 55), elapsed)
	}
	fmt.Fprint(os.Stderr, line+clearPad(line))
}

// Retrying prints a one-off status line for a backoff/retry event without
// disturbing the step counter, so 429s and network hiccups are visible
// instead of the tool going silent while it waits.
func (p *Progress) Retrying(url string, attempt, maxAttempts int, wait time.Duration, reason error) {
	line := fmt.Sprintf("\r⏳ rate-limited/retrying (%d/%d), waiting %s: %s",
		attempt, maxAttempts, wait.Round(100*time.Millisecond), truncate(shortURL(url), 60))
	fmt.Fprint(os.Stderr, line+clearPad(line))
}

// Finish prints a final newline-terminated summary so subsequent output
// starts on a clean line.
func (p *Progress) Finish(summary string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	elapsed := time.Since(p.started).Round(100 * time.Millisecond)
	line := fmt.Sprintf("\r%s (%s)", summary, elapsed)
	fmt.Fprintln(os.Stderr, line+clearPad(line))
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n-3] + "..."
	}
	return s
}

func shortURL(url string) string {
	if len(url) > 60 {
		return "..." + url[len(url)-57:]
	}
	return url
}

// clearPad returns enough trailing spaces to overwrite a previous, longer
// line before the cursor returns to column 0 on the next \r-prefixed write.
func clearPad(line string) string {
	const terminalWidthGuess = 120
	pad := terminalWidthGuess - len(line)
	if pad < 0 {
		pad = 0
	}
	return strings.Repeat(" ", pad)
}

// and concrete next steps. CLI commands should prefer returning/printing
// these over bare errors so failures are actionable rather than opaque.
type Diagnostic struct {
	Title  string   // one-line summary, e.g. "Version conflict"
	Detail string   // longer explanation of what happened
	Fixes  []string // ordered list of concrete things the user can try
}

func (d Diagnostic) Print() {
	fmt.Fprintf(os.Stderr, "\n✖ %s\n", d.Title)
	if d.Detail != "" {
		fmt.Fprintf(os.Stderr, "  %s\n", d.Detail)
	}
	if len(d.Fixes) > 0 {
		fmt.Fprintln(os.Stderr, "\n  Try:")
		for _, f := range d.Fixes {
			fmt.Fprintf(os.Stderr, "    • %s\n", f)
		}
	}
	fmt.Fprintln(os.Stderr)
}