package install

import (
	"fmt"
	"os"
	"path/filepath"

	"kpm/types"
)

// DownloadUrl downloads a JAR from a URL and updates the resource file
func DownloadUrl(artifact string, resourcedata *types.ResourceFile, url string, update bool, index int) {
	if artifact == "" || url == "" {
		fmt.Println("Error: artifact and url cannot be empty")
		return
	}

	// Build the local file path for URL-based artifacts
	fileName := fmt.Sprintf("%s.jar", artifact)
	cwd, errCwd := os.Getwd()
	var file string
	if errCwd == nil {
		file = filepath.Join(cwd, "libs", artifact, fileName)
	} else {
		file = fmt.Sprintf("./libs/%s/%s.jar", artifact, artifact)
	}

	// If file already exists and not forcing update, just verify metadata
	if _, err := os.Stat(file); err == nil && !update {
		fmt.Println("Info: file already exists, checking resource metadata:", file)
		idx := findResourceIndex(resourcedata, "", artifact, "")
		if idx >= 0 {
			fmt.Println("Info: resource metadata already present for", artifact)
			return
		}

		fmt.Println("Info: resource metadata missing; appending resource entry for", artifact)
		AppendResource(
			resourcedata,
			"",
			artifact,
			"",      // Version
			"url",   // Source
			"jar",   // Type
			"",      // Domain
			"",      // Path
			file,    // LPath
			"",      // GPath
			url,     // URL
			"works", // Hash
			"",
			"",
		)

		// Update package.kpm with the installed package
		if err := UpdatePackageDependencyFromUrl(artifact); err != nil {
			fmt.Println("Warning: failed to update package.kpm:", err)
		}
		return
	}

	if update && index >= 0 {
		UpdateResource(
			resourcedata,
			"",
			artifact,
			"",      // Version
			"url",   // Source
			"jar",   // Type
			"",      // Domain
			"",      // Path
			file,    // LPath
			"",      // GPath
			url,     // URL
			"works", // Hash
			index,
		)
	} else {
		AppendResource(
			resourcedata,
			"",
			artifact,
			"",      // Version
			"url",   // Source
			"jar",   // Type
			"",      // Domain
			"",      // Path
			file,    // LPath
			"",      // GPath
			url,     // URL
			"works", // Hash
			"",
			"",
		)
	}

	// Update package.kpm with the installed package
	if err := UpdatePackageDependencyFromUrl(artifact); err != nil {
		fmt.Println("Warning: failed to update package.kpm:", err)
	}
}
