package install

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/scagogogo/sonatype-central-sdk/pkg/api"
)

// DownloadJar downloads a JAR using the Sonatype SDK client and saves it directly.
func DownloadJar(path string, outFile string, global bool) error {
	fmt.Println("Downloading:", path)

	// Note: All downloads go to project-local ./libs/ directory for portability.
	// The 'global' parameter is deprecated and ignored.

	// Use Sonatype API client to download bytes
	client := api.NewClient()
	data, err := client.Download(context.Background(),path)
	if err != nil {
		return fmt.Errorf("failed to download JAR: %w", err)
	}

	// Ensure destination folder exists
	dir := filepath.Dir(outFile)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}

	// Write bytes directly to file
	if err := os.WriteFile(outFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Println("Downloaded JAR to", outFile)
	return nil
}
