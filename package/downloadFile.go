package install

import (
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
    if url == "" || outFile == "" {
        return fmt.Errorf("invalid url or output path")
    }

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return err
    }

    resp, err := client.Do(req)
    if err != nil {
        return err
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
    return err
}
