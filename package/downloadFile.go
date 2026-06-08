package install

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var client = &http.Client{
	Timeout: 30 * time.Second,
}

func DownloadFile(url, outFile string) error {
	req, _ := http.NewRequestWithContext(context.Background(), "GET", url, nil)

	// simple retry (handles 429 + transient failures)
	var resp *http.Response
	var err error

	for i := 0; i < 3; i++ {
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode < 500 && resp.StatusCode != 429 {
			break
		}
		time.Sleep(time.Duration(i+1) * 500 * time.Millisecond)
	}

	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(outFile), 0755); err != nil {
		return err
	}

	f, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

