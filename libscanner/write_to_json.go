package libscanner

import (
	"fmt"
	"os"
)

// WriteToJson writes the given data to package.kpm
func WriteToJson(data []byte, file string) error {
	err := os.WriteFile(file, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing package.kpm: %w", err)
	}
	return nil
}
