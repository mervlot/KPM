// Package downloader performs HTTP fetches with retry/backoff and drives a
// bounded worker pool for parallel artifact downloads.
package downloader

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"time"
)

// OnRetry, if set, is called before every retry sleep so callers (the CLI)
// can print live feedback instead of going silent while backing off.
var OnRetry func(url string, attempt, maxAttempts int, wait time.Duration, reason error)

// globalLimit is a hard sliding-window rate limiter shared by every
// goroutine and every phase (sequential resolve, concurrent install pool
// alike): no more than maxPerWin requests may go out in any rolling
// `window`. This is deliberately a hard cap, not just an average spacing —
// a fixed-interval pacer still lets a burst of retries or concurrent
// goroutines land within the same window as long as each individual gap is
// respected; a sliding window actually refuses the (N+1)th request until
// the window has room. Default: at most 3 requests per 2-second window
// (~1.5 req/s sustained) — conservative because this host has been
// observed 429'ing even under a fixed 700ms pace.
var globalLimit = newRateLimiter(3, 2*time.Second)

type rateLimiter struct {
	mu         sync.Mutex
	maxPerWin  int
	window     time.Duration
	timestamps []time.Time // start times of requests currently inside the window
}

func newRateLimiter(maxPerWin int, window time.Duration) *rateLimiter {
	return &rateLimiter{maxPerWin: maxPerWin, window: window}
}

// wait blocks until issuing another request would not exceed maxPerWin
// requests in the trailing `window`, then records this request's timestamp.
func (r *rateLimiter) wait() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for {
		now := time.Now()
		cutoff := now.Add(-r.window)

		kept := r.timestamps[:0]
		for _, t := range r.timestamps {
			if t.After(cutoff) {
				kept = append(kept, t)
			}
		}
		r.timestamps = kept

		if len(r.timestamps) < r.maxPerWin {
			r.timestamps = append(r.timestamps, now)
			return
		}

		// Window is full: sleep until the oldest entry ages out, then re-check
		// (another goroutine may have taken the freed slot in the meantime).
		sleepFor := r.timestamps[0].Add(r.window).Sub(now)
		r.mu.Unlock()
		if sleepFor > 0 {
			time.Sleep(sleepFor)
		}
		r.mu.Lock()
	}
}

// SetRateLimit reconfigures the global request budget process-wide (used by
// `kpm --slow`, or for a trusted private repository that can take more load).
func SetRateLimit(maxPerWin int, window time.Duration) {
	globalLimit = newRateLimiter(maxPerWin, window)
}

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
// success or retry alike — blocks on the global sliding-window rate limiter
// so the total number of requests in flight anywhere in the process never
// exceeds the configured budget. 4xx other than 429 fails fast since
// retrying won't help (e.g. 404 = artifact genuinely doesn't exist there —
// see resolver.NotFoundInAnyRepoError for how that becomes a human-readable
// message upstream).
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

		globalLimit.wait() // hard cap on EVERY attempt, not just the first

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
	wait := base * time.Duration(1<<capped)
	if wait > 30*time.Second {
		wait = 30 * time.Second
	}
	jitter := time.Duration(rand.Intn(300)) * time.Millisecond
	return wait + jitter
}

// Job represents a unit of work for the downloader pool.
type Job struct {
	Name string
	Run  func() error
}

// RunPool executes jobs with bounded concurrency, collecting all errors.
// It ensures that transient failures are handled per-job (via the Job.Run func),
// and aggregates any final errors for the caller to handle.
func RunPool(jobs []Job, concurrency int) []error {
	if concurrency <= 0 {
		concurrency = 1
	}

	type result struct {
		name string
		err  error
	}

	jobChan := make(chan Job, len(jobs))
	resChan := make(chan result, len(jobs))

	for _, j := range jobs {
		jobChan <- j
	}
	close(jobChan)

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobChan {
				err := job.Run()
				resChan <- result{name: job.Name, err: err}
			}
		}()
	}

	wg.Wait()
	close(resChan)

	var errs []error
	for r := range resChan {
		if r.err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", r.name, r.err))
		}
	}

	// Sort errors for deterministic output
	sort.Slice(errs, func(i, j int) bool {
		return errs[i].Error() < errs[j].Error()
	})

	return errs
}

// NotFoundError is returned when a repository returns a 404.
type NotFoundError struct {
	URL string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("not found: %s", e.URL)
}