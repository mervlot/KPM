package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// CopyFile copies a file from src to dst.
// It creates the destination directory if it doesn't exist.
func CopyFile(src, dst string) error {
	// 1. Ensure the destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// 2. Open source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer sourceFile.Close()

	// 3. Create destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer destFile.Close()

	// 4. Perform the copy
	bytesWritten, err := io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// 5. Ensure everything is written to disk
	if err := destFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync file to disk: %w", err)
	}

	fmt.Printf("📂 Copied %d bytes to %s\n", bytesWritten, dst)
	return nil
}