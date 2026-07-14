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
//
// Why this previously printed a new line per step instead of overwriting,
// on some Windows terminals: the line was padded to a guessed fixed width
// (120 columns) and included multi-byte emoji, which many terminal fonts
// render as 2 display columns wide even though they're a small number of
// UTF-8 bytes. On any terminal narrower than the guess (very common — 80
// and 100-column windows are typical), that pushed the line past the real
// width and the terminal auto-wrapped it onto a second physical row. `\r`
// only returns the cursor to column 0 of whatever row it's currently on —
// it doesn't know or care that the row above is "the same logical line" —
// so the next update just overwrote the tail end of the wrapped row while
// the first half stayed put, and it looked like the tool was spamming a
// new line every step.
//
// Fix: the live line never contains emoji, its detail portion is capped
// short enough to comfortably fit even an 80-column terminal, and instead
// of guessing a fixed pad width, we remember exactly how long the
// previous line was and pad only by the difference — so there's never a
// wrapping risk regardless of terminal size, and one-shot lines (Retrying,
// Finish) that trail off don't leave garbage from a longer previous line.
type Progress struct {
	mu      sync.Mutex
	label   string
	done    int
	total   int
	started time.Time
	prevLen int // length (in runes) of the last line written, for exact-pad overwrite
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

// maxLiveLineWidth caps every in-place-updating line at a width that fits
// comfortably inside even a narrow (80-column) terminal, so it can never
// wrap onto a second row and break the \r-overwrite trick.
const maxLiveLineWidth = 76

// writeLine overwrites the previous live line with a new one: \r returns to
// column 0, the new content is written, and any leftover characters from a
// longer previous line are blanked with exactly enough spaces (no more, no
// guessing) before the cursor sits at the end of the new content.
func (p *Progress) writeLine(content string) {
	runes := []rune(content)
	if len(runes) > maxLiveLineWidth {
		runes = runes[:maxLiveLineWidth]
		content = string(runes)
	}
	pad := 0
	if p.prevLen > len(runes) {
		pad = p.prevLen - len(runes)
	}
	fmt.Fprint(os.Stderr, "\r"+content+strings.Repeat(" ", pad)+"\r"+content)
	p.prevLen = len(runes)
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
		line = fmt.Sprintf("%s [%d/%d] %s (%s)", p.label, p.done, p.total, truncate(detail, 30), elapsed)
	} else {
		line = fmt.Sprintf("%s [%d] %s (%s)", p.label, p.done, truncate(detail, 30), elapsed)
	}
	p.writeLine(line)
}

// Retrying prints a one-off status line for a backoff/retry event without
// disturbing the step counter, so 429s and network hiccups are visible
// instead of the tool going silent while it waits.
func (p *Progress) Retrying(url string, attempt, maxAttempts int, wait time.Duration, reason error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	line := fmt.Sprintf("waiting %s before retry %d/%d: %s",
		wait.Round(100*time.Millisecond), attempt, maxAttempts, truncate(shortURL(url), 40))
	p.writeLine(line)
}

// Finish prints a final newline-terminated summary so subsequent output
// starts on a clean line. This one IS allowed to use emoji / run long,
// since it's a one-shot line ending in \n rather than something meant to be
// overwritten in place.
func (p *Progress) Finish(summary string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	elapsed := time.Since(p.started).Round(100 * time.Millisecond)
	pad := 0
	line := fmt.Sprintf("%s (%s)", summary, elapsed)
	if p.prevLen > len([]rune(line)) {
		pad = p.prevLen - len([]rune(line))
	}
	fmt.Fprintln(os.Stderr, "\r"+line+strings.Repeat(" ", pad))
	p.prevLen = 0
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

// Diagnostic is a structured, user-facing error report: what happened, why,
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