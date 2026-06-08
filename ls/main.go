package ls

import (
	"fmt"
	"log"
	"os"
)

func Ls(dirPath string) {
	// Directory to list (current directory ".")
	// dirPath := "."

	// Open the directory
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	}

	// Loop through entries
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			log.Printf("Could not get info for %s: %v", entry.Name(), err)
			continue
		}

		// Print file name and size
		fmt.Printf("%-20s %10d bytes\n", info.Name(), info.Size())
	}
}
