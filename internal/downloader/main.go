package downloader

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// OnRetry is called before every retry sleep so callers can print live feedback.
var OnRetry func(url string, attempt, maxAttempts int, wait time.Duration, reason error)

// globalPace enforces a minimum spacing between the START of any two
// outgoing requests across the whole process. 
// 500ms = 2 requests per second (safely within Maven Central's 1-4 req/s tolerance).
var globalPace = &pacer{minInterval: 500 * time.Millisecond}

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

// SetGlobalPace overrides the minimum spacing between requests process-wide.
func SetGlobalPace(d time.Duration) {
	globalPace.mu.Lock()
	defer globalPace.mu.Unlock()
	globalPace.minInterval = d
}

// Client handles HTTP requests with built-in retries and rate limiting.
type Client struct {
	http       *http.Client
	MaxRetries int
	BaseDelay  time.Duration
}

// New creates a new downloader client.
func New() *Client {
	return &Client{
		http:       &http.Client{Timeout: 60 * time.Second},
		MaxRetries: 5,
		BaseDelay:  1000 * time.Millisecond,
	}
}

// Get fetches a URL, applying rate limiting and retries.
func (c *Client) Get(url string) ([]byte, error) {
	return c.GetAuth(url, "", "")
}

// GetAuth fetches a URL with optional basic auth, applying rate limiting and retries.
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

		// Pace EVERY attempt to strictly enforce the global rate limit
		globalPace.wait()

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		if username != "" {
			req.SetBasicAuth(username, password)
		}
		
		// Standard User-Agent to avoid aggressive bot filtering by Maven Central
		req.Header.Set("User-Agent", "KPM-Dependency-Resolver/1.0")

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
		return nil, fmt.Errorf("repository returned %d for %s (client error, not retrying)", resp.StatusCode, url)
	}
	return nil, fmt.Errorf("giving up after %d attempts: %w", c.MaxRetries+1, lastErr)
}

// backoffWithJitter grows 1s, 2s, 4s, 8s... and adds up to 500ms of random jitter.
func backoffWithJitter(base time.Duration, attempt int) time.Duration {
	capped := attempt
	if capped > 6 {
		capped = 6 // avoid overflow / absurdly long waits
	}
	wait := base * time.Duration(1<<capped)
	if wait > 30*time.Second {
		wait = 30 * time.Second
	}
	jitter := time.Duration(rand.Intn(500)) * time.Millisecond
	return wait + jitter
}

// NotFoundError indicates a 404 response.
type NotFoundError struct {
	URL string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("not found: %s", e.URL)
}

// Job represents a unit of work for the sequential runner.
type Job struct {
	Name string
	Run  func() error
}

// RunPool executes jobs STRICTLY SEQUENTIALLY to prevent concurrent bursts 
// that trigger Maven Central 429s. 
func RunPool(jobs []Job, concurrency int) []error {
	// Note: 'concurrency' parameter is kept for API compatibility, but ignored 
	// to guarantee sequential execution and prevent rate-limit bans.
	var errs []error
	for _, job := range jobs {
		if err := job.Run(); err != nil {
			errs = append(errs, fmt.Errorf("job %s failed: %w", job.Name, err))
		}
	}
	return errs
}