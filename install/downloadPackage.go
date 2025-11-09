package install

import (
	"fmt"
	"kpm/types"
	"os"
	"path/filepath"
)

// DownloadMaven downloads a Maven artifact and updates the resource file
func DownloadMaven(group, artifact, version string, update bool, resourcedata *types.ResourceFile, index int) {
	maven := types.Mavenurl{
		Group:    group,
		Artifact: artifact,
		Version:  version,
	}

	// Fetch latest metadata if needed
	mavenMeta, err := GetMavenMetadata(maven.MetadataUrl())
	if err != nil {
		fmt.Println("Failed to fetch metadata:", err)
		return
	}

	if update || version == "" || version == "latest" {
		version = mavenMeta.Versioning.Latest
	}
	maven.Version = version

	url := maven.LatestUrl(version)

	// Build file path as ./libs/<group>/<artifact>/<version>/<artifact>-<version>.jar
	fileName := fmt.Sprintf("%s-%s.jar", artifact, version)
	// Try to use absolute path based on current working dir so resource LPath is consistent
	cwd, errCwd := os.Getwd()
	var file string
	if errCwd == nil {
		file = filepath.Join(cwd, "libs", group, artifact, version, fileName)
	} else {
		// fallback to relative path
		file = fmt.Sprintf("./libs/%s/%s/%s/%s-%s.jar", group, artifact, version, artifact, version)
	}

	// If the target file already exists, do NOT download. Ensure resource metadata
	// is present and up-to-date (append or update as needed), then return.
	if _, err := os.Stat(file); err == nil {
		fmt.Println("File already exists, checking resource metadata:", file)

		// If caller provided an index (update flow), prefer updating that entry.
		if index >= 0 && index < len(resourcedata.Resources) {
			existing := resourcedata.Resources[index]
			existingVersion := ""
			if existing.Version != nil {
				existingVersion = *existing.Version
			}
			if existingVersion == version {
				fmt.Println("Resource already at required version:", version)
				return
			}
			// Update the existing resource entry at provided index
			UpdateResource(resourcedata, group, artifact, version, "maven", "jar",
				"https://repo1.maven.org/maven2/", url,
				fmt.Sprintf("%s/%s/%s/%s-%s.jar", group, artifact, version, artifact, version),
				fmt.Sprintf("~/.kpm/repo/%s/%s/%s/%s-%s.jar", group, artifact, version, artifact, version),
				"", "sha256:82af21f1c0e8ce74c5...", index)
			return
		}

		// Try to find an exact resource entry for this version
		idx := findResourceIndex(resourcedata, group, artifact, version)
		if idx >= 0 {
			fmt.Println("Resource metadata already present for", artifact, "version", version)
			return
		}

		// If there's an entry for the same group+name but different version, update it
		idxAny := findResourceIndex(resourcedata, group, artifact, "")
		if idxAny >= 0 {
			fmt.Println("Updating existing resource entry to version", version)
			UpdateResource(resourcedata, group, artifact, version, "maven", "jar",
				"https://repo1.maven.org/maven2/", url,
				fmt.Sprintf("%s/%s/%s/%s-%s.jar", group, artifact, version, artifact, version),
				fmt.Sprintf("~/.kpm/repo/%s/%s/%s/%s-%s.jar", group, artifact, version, artifact, version),
				"", "sha256:82af21f1c0e8ce74c5...", idxAny)
			return
		}

		// No resource entry exists; append one
		fmt.Println("Resource metadata missing; appending resource entry for", artifact)
		AppendResource(resourcedata, group, artifact, version, "maven", "jar",
			"https://repo1.maven.org/maven2/", url,
			fmt.Sprintf("%s/%s/%s/%s-%s.jar", group, artifact, version, artifact, version),
			fmt.Sprintf("~/.kpm/repo/%s/%s/%s/%s-%s.jar", group, artifact, version, artifact, version),
			"", "sha256:82af21f1c0e8ce74c5...")
		return
	}

	// Ensure folder exists


	// Download jar
	if err := DownloadJar(url, file, false); err != nil {
		fmt.Println("Download failed:", err)
		return
	}

	// Update or append resource entry
	if update && index >= 0 {
		UpdateResource(resourcedata, group, artifact, version, "maven", "jar",
			"https://repo1.maven.org/maven2/", url,
			fmt.Sprintf("%s/%s/%s/%s-%s.jar", group, artifact, version, artifact, version),
			fmt.Sprintf("%s", file),
			"", "sha256:82af21f1c0e8ce74c5c...", index)
	} else {
		AppendResource(resourcedata, group, artifact, version, "maven", "jar",
			"https://repo1.maven.org/maven2/", url,
			fmt.Sprintf("%s/%s/%s/%s-%s.jar", group, artifact, version, artifact, version),
			fmt.Sprintf("%s", file),
			"", "sha256:82af21f1c0e8ce74c5c...")
	}
}

// DownloadUrl downloads a URL-based jar and updates the resource file
func DownloadUrl(artifact string, resourcedata *types.ResourceFile, url string, update bool, index int) {
	// Build absolute path for URL-based jar: <cwd>/libs/<artifact>/<artifact>.jar
	fileName := fmt.Sprintf("%s.jar", artifact)
	cwd, errCwd := os.Getwd()
	var file string
	if errCwd == nil {
		file = filepath.Join(cwd, "libs", artifact, fileName)
	} else {
		file = fmt.Sprintf("./libs/%s/%s.jar", artifact, artifact)
	}

	// Skip everything if file exists and update is false
	if _, err := os.Stat(file); err == nil && !update {
		fmt.Println("File already exists, checking resource metadata:", file)
		idx := findResourceIndex(resourcedata, "", artifact, "")
		if idx >= 0 {
			fmt.Println("Resource metadata already present for", artifact)
			return
		}
		fmt.Println("Resource metadata missing; appending resource entry for", artifact)
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
		)
		return
	}



	// Download jar
	if err := DownloadJar(url, file, false); err != nil {
		fmt.Println("Download failed:", err)
		return
	}

	// Update or append resource entry
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
		)
	}
}
