// ...existing code...
package install

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// DownloadJar downloads a JAR using the shared httpClient and creates folders correctly.
func DownloadJar(url string, outFile string, global bool) error {
	// If user requested global path, ensure we expand to $HOME/.kpm/libs (not literal "~")
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to determine user home: %w", err)
		}
		// ensure base folder exists (not strictly required if outFile is absolute, but safe)
		if err := os.MkdirAll(filepath.Join(home, ".kpm", "libs"), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create global libs folder: %w", err)
		}
	}

	// Request using shared client
	resp, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download JAR, HTTP %d: %s", resp.StatusCode, url)
	}
	// Ensure parent folder for outFile exists
	dir := filepath.Dir(outFile)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}
	// Create file
	file, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Println("Downloaded JAR to", outFile)
	return nil
}

// ...existing code...
