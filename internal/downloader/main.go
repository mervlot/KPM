// Package downloader performs HTTP fetches with retry/backoff and drives a
// bounded worker pool for parallel artifact downloads.
package downloader

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// OnRetry, if set, is called before every retry sleep so callers (the CLI)
// can print live feedback instead of going silent while backing off.
var OnRetry func(url string, attempt, maxAttempts int, wait time.Duration, reason error)

// globalPace enforces a minimum spacing between the START of any two
// outgoing requests, across the whole process — regardless of how many
// goroutines are involved. This matters because Maven Central throttles on
// aggregate request RATE over time, not on concurrency: even a single
// goroutine firing requests back-to-back as fast as the network allows can
// (and, as reported, does) trip a 429. Backoff-after-failure alone isn't
// enough; the fix is to never send fast enough to get throttled in the
// first place. Default: one request in flight every 700ms (~1.4 req/s).
// This was raised from an initial 350ms after real-world testing still hit
// 429s at that rate — Central's actual tolerance is unpublished and can
// vary by time of day/network, so this errs slow rather than re-tuning by
// guesswork. Combined with skipping checksum sidecar requests for POMs
// (see resolver/fetch.go), this cuts both the rate AND the raw request
// count for the resolve phase, which is where the 429s were happening.
//
// Package-level and mutex-guarded (not per-Client) so the resolve phase
// (sequential) and the install phase (worker pool) share the same budget —
// otherwise pacing each independently would still let the two phases
// combine into a burst.
var globalPace = &pacer{minInterval: 700 * time.Millisecond}

type pacer struct {
	mu          sync.Mutex
	minInterval time.Duration
	last        time.Time
}

func (p *pacer) wait() {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	if elapsed := now.Sub(p.last); elapsed < p.minInterval {
		time.Sleep(p.minInterval - elapsed)
	}
	p.last = time.Now()
}

// SetGlobalPace overrides the minimum spacing between requests process-wide
// (e.g. the CLI can slow this down further with a --slow flag, or speed it
// up for a trusted private repository that isn't Maven Central).
func SetGlobalPace(d time.Duration) { globalPace.minInterval = d }

type Client struct {
	http       *http.Client
	MaxRetries int
	BaseDelay  time.Duration
}

func New() *Client {
	return &Client{
		http:       &http.Client{Timeout: 60 * time.Second},
		MaxRetries: 5,
		BaseDelay:  500 * time.Millisecond,
	}
}

// Get fetches a URL's body with exponential-backoff-plus-jitter retries on
// transient failures (network errors, 5xx, 429), and — before every attempt,
// success or retry alike — waits on the global pacer so requests never fire
// faster than globalPace.minInterval apart. 4xx other than 429 fails fast
// since retrying won't help (e.g. 404 = artifact genuinely doesn't exist
// there).
//
// On 429 specifically: Maven Central does not publish a fixed "N requests
// per second" limit — since 2024 it throttles based on aggregate traffic
// from your IP/org rather than a documented per-client rate. The fix isn't
// "retry harder," it's "never send fast enough to get flagged": every
// request, whether it's the sequential resolve phase or the concurrent
// install phase, passes through the same global pacer below.
func (c *Client) Get(url string) ([]byte, error) {
	return c.GetAuth(url, "", "")
}

func (c *Client) GetAuth(url, username, password string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.MaxRetries; attempt++ {
		if attempt > 0 {
			wait := backoffWithJitter(c.BaseDelay, attempt)
			if OnRetry != nil {
				OnRetry(url, attempt, c.MaxRetries, wait, lastErr)
			}
			time.Sleep(wait)
		}

		globalPace.wait() // pace EVERY attempt, not just the first

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		if username != "" {
			req.SetBasicAuth(username, password)
		}
		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("network error fetching %s: %w", url, err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			data, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				lastErr = fmt.Errorf("reading response body from %s: %w", url, err)
				continue
			}
			return data, nil
		}

		resp.Body.Close()
		if resp.StatusCode == 404 {
			return nil, &NotFoundError{URL: url}
		}
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("repository returned %d for %s", resp.StatusCode, url)
			continue // retryable
		}
		// Other 4xx: not retryable (bad auth, forbidden, etc.)
		return nil, fmt.Errorf("repository returned %d for %s (not retrying: client error)", resp.StatusCode, url)
	}
	return nil, fmt.Errorf("giving up after %d attempts: %w", c.MaxRetries+1, lastErr)
}

// backoffWithJitter grows 500ms, 1s, 2s, 4s, ... and adds up to 300ms of
// random jitter so a burst of parallel requests that all got 429'd at once
// don't all retry at exactly the same instant and re-trigger the limit.
func backoffWithJitter(base time.Duration, attempt int) time.Duration {
	capped := attempt
	if capped > 6 {
		capped = 6 // avoid overflow / absurdly long waits
	}
	wait := base * time.Duration(1<<uint(capped-1))
	wait += time.Duration(rand.Intn(300)) * time.Millisecond
	return wait
}

type NotFoundError struct{ URL string }

func (e *NotFoundError) Error() string { return fmt.Sprintf("not found: %s", e.URL) }

// Job is one unit of parallel download work.
type Job struct {
	Name string
	Run  func() error
}

// RunPool executes jobs with up to `concurrency` running at once, collecting
// all errors rather than stopping at the first (so one broken artifact
// doesn't hide unrelated failures elsewhere in a large resolution).
func RunPool(jobs []Job, concurrency int) []error {
	if concurrency < 1 {
		concurrency = 1
	}
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	for _, j := range jobs {
		wg.Add(1)
		sem <- struct{}{}
		go func(job Job) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := job.Run(); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("%s: %w", job.Name, err))
				mu.Unlock()
			}
		}(j)
	}
	wg.Wait()
	return errs
}