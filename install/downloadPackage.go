package install

import (
	"encoding/xml"
	"fmt"
	"io"
	"kpm/types"
	"net/http"
	"os"
	"path/filepath"
)

// DownloadMaven downloads a Maven artifact and updates both resource.kpm and package.kpm
func DownloadMaven(group, artifact, version string, update bool, resourcedata *types.ResourceFile, index int) {
	DownloadMavenInternal(group, artifact, version, update, resourcedata, index, true)
}

// DownloadMavenInternal is the internal implementation of DownloadMaven
// updatePackageKpm controls whether to add the dependency to package.kpm (true for direct, false for transitive)
func DownloadMavenInternal(group, artifact, version string, update bool, resourcedata *types.ResourceFile, index int, updatePackageKpm bool) {
	if group == "" || artifact == "" {
		fmt.Println("Error: group and artifact cannot be empty")
		return
	}

	maven := types.Mavenurl{
		Group:    group,
		Artifact: artifact,
		Version:  version,
	}

	fmt.Println("Fetching metadata:", maven.MetadataUrl())

	// Fetch latest version metadata from Maven repository
	mavenMeta, err := GetMavenMetadata(maven.MetadataUrl())
	if err != nil {
		fmt.Println("Error: failed to fetch metadata:", err)
		return
	}

	// Determine the actual version to download
	if update || version == "" || version == "latest" {
		if mavenMeta.Versioning.Latest == "" {
			fmt.Println("Error: no version information found in metadata")
			return
		}
		version = mavenMeta.Versioning.Latest
	}
	maven.Version = version

	// Fetch the POM file to extract packaging type and dependencies
	pomURL := maven.PomUrl()
	fmt.Println("Fetching POM:", pomURL)

	resp, err := http.Get(pomURL)
	if err != nil {
		fmt.Println("Error: failed to fetch POM:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("Error: failed to fetch POM (HTTP %d): %s\n", resp.StatusCode, resp.Status)
		return
	}

	pomBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error: failed to read POM response body:", err)
		return
	}

	// Parse POM to extract packaging type
	type Pom struct {
		Packaging string `xml:"packaging"`
	}

	var pom Pom
	if err := xml.Unmarshal(pomBytes, &pom); err != nil {
		fmt.Println("Warning: failed to parse POM packaging, defaulting to jar:", err)
	}

	packaging := pom.Packaging
	if packaging == "" {
		packaging = "jar"
	}

	fmt.Println("Packaging type:", packaging)

	// Validate packaging type - skip unsupported types
	if packaging == "pom" {
		fmt.Println("Info: skipping pom-only artifact (no binary to download)")
		return
	}

	if packaging != "jar" && packaging != "war" {
		fmt.Printf("Error: unsupported packaging type '%s' (only jar and war supported)\n", packaging)
		return
	}

	ext := packaging
	url := maven.BuildLatestPath(version, ext)

	// Build the local file path
	fileName := fmt.Sprintf("%s-%s.%s", artifact, version, ext)
	cwd, errCwd := os.Getwd()
	var file string
	if errCwd == nil {
		file = filepath.Join(cwd, "libs", group, artifact, version, fileName)
	} else {
		file = fmt.Sprintf("./libs/%s/%s/%s/%s-%s.%s", group, artifact, version, artifact, version, ext)
	}

	// Check if the artifact file already exists
	if _, err := os.Stat(file); err == nil {
		fmt.Println("Info: file already exists, checking resource metadata:", file)

		// Handle case where we're updating an existing resource entry
		if index >= 0 && index < len(resourcedata.Resources) {
			existing := resourcedata.Resources[index]
			existingVersion := ""
			if existing.Version != nil {
				existingVersion = *existing.Version
			}

			if existingVersion == version {
				fmt.Println("Info: resource already at required version:", version)
				// Update package.kpm for direct dependencies
				if updatePackageKpm {
					if err := UpdatePackageDependencyFromMaven(group, artifact, version); err != nil {
						fmt.Println("Warning: failed to update package.kpm:", err)
					}
				}
				return
			}

			// Update the resource with new version
			UpdateResource(resourcedata, group, artifact, version,
				"maven", ext,
				"https://repo1.maven.org/maven2/", url,
				file,
				fmtRelativePath(group, artifact, version, artifact, version, ext),
				"", "sha256...", index)
			if updatePackageKpm {
				if err := UpdatePackageDependencyFromMaven(group, artifact, version); err != nil {
					fmt.Println("Warning: failed to update package.kpm:", err)
				}
			}
			return
		}

		// Check if resource metadata already exists
		idx := findResourceIndex(resourcedata, group, artifact, version)
		if idx >= 0 {
			fmt.Println("Info: resource metadata already present")
			if updatePackageKpm {
				if err := UpdatePackageDependencyFromMaven(group, artifact, version); err != nil {
					fmt.Println("Warning: failed to update package.kpm:", err)
				}
			}
			return
		}

		// Append new resource entry
		AppendResource(resourcedata, group, artifact, version,
			"maven", ext,
			"https://repo1.maven.org/maven2/", url,
			file,
			fmtRelativePath(group, artifact, version, artifact, version, ext),
			"", "sha256...")

		if updatePackageKpm {
			if err := UpdatePackageDependencyFromMaven(group, artifact, version); err != nil {
				fmt.Println("Warning: failed to update package.kpm:", err)
			}
		}
		return
	}

	// Download the artifact from Maven repository
	if err := DownloadJar(url, file, false); err != nil {
		fmt.Printf("Error: download failed for %s:%s:%s: %v\n", group, artifact, version, err)
		return
	}

	fmt.Printf("Successfully downloaded: %s:%s:%s\n", group, artifact, version)

	// Update resource metadata
	if update && index >= 0 {
		UpdateResource(resourcedata, group, artifact, version,
			"maven", ext,
			"https://repo1.maven.org/maven2/", url,
			file,
			fmtRelativePath(group, artifact, version, artifact, version, ext),
			"", "sha256...", index)
	} else {
		AppendResource(resourcedata, group, artifact, version,
			"maven", ext,
			"https://repo1.maven.org/maven2/", url,
			file,
			fmtRelativePath(group, artifact, version, artifact, version, ext),
			"", "sha256...")
	}

	// Update package.kpm for direct dependencies only
	if updatePackageKpm {
		if err := UpdatePackageDependencyFromMaven(group, artifact, version); err != nil {
			fmt.Println("Warning: failed to update package.kpm:", err)
		}
	}

	// Download POM dependencies (will not be added to package.kpm)
	DownloadPomDependencies(pomBytes, resourcedata)
}

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
		)

		// Update package.kpm with the installed package
		if err := UpdatePackageDependencyFromUrl(artifact); err != nil {
			fmt.Println("Warning: failed to update package.kpm:", err)
		}
		return
	}

	// Download the JAR from the URL
	if err := DownloadJar(url, file, false); err != nil {
		fmt.Printf("Error: download failed for %s: %v\n", artifact, err)
		return
	}

	fmt.Printf("Successfully downloaded: %s from %s\n", artifact, url)

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

	// Update package.kpm with the installed package
	if err := UpdatePackageDependencyFromUrl(artifact); err != nil {
		fmt.Println("Warning: failed to update package.kpm:", err)
	}
}

// fmtRelativePath builds a project-relative local path for Maven artifacts.
// Returns a path like ./libs/group/artifact/version/artifact-version.ext
func fmtRelativePath(group, artifact, version, name, ver, ext string) string {
	return filepath.ToSlash(filepath.Join(".", "libs", group, artifact, version, fmt.Sprintf("%s-%s.%s", name, ver, ext)))
}

// PomDependency represents a single Maven dependency from a POM file
type PomDependency struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
	Scope      string `xml:"scope"`
}

// ParsePomDependencies extracts all dependencies from a POM XML byte array
// Returns a slice of PomDependency structs
func ParsePomDependencies(pomBytes []byte) ([]PomDependency, error) {
	if len(pomBytes) == 0 {
		return []PomDependency{}, fmt.Errorf("empty POM content")
	}

	type PomRoot struct {
		XMLName      xml.Name `xml:"project"`
		Dependencies struct {
			Dependency []PomDependency `xml:"dependency"`
		} `xml:"dependencies"`
	}

	var pom PomRoot
	err := xml.Unmarshal(pomBytes, &pom)
	if err != nil {
		return nil, fmt.Errorf("failed to parse POM XML: %w", err)
	}

	return pom.Dependencies.Dependency, nil
}

// DownloadPomDependencies automatically downloads all dependencies declared in a POM file
// These dependencies are added to resource.kpm only, not to package.kpm
func DownloadPomDependencies(pomBytes []byte, resourcedata *types.ResourceFile) {
	if resourcedata == nil {
		fmt.Println("Error: resourcedata cannot be nil")
		return
	}

	deps, err := ParsePomDependencies(pomBytes)
	if err != nil {
		fmt.Println("Warning: failed to parse POM dependencies:", err)
		return
	}

	if len(deps) == 0 {
		fmt.Println("Info: no dependencies found in POM")
		return
	}

	fmt.Printf("\nInfo: found %d POM dependencies, starting download...\n", len(deps))

	installedCount := 0
	skippedCount := 0

	for _, dep := range deps {
		// Skip test and provided scopes as they're not needed at runtime
		if dep.Scope == "test" || dep.Scope == "provided" {
			fmt.Printf("  Skipping %s:%s (scope: %s)\n", dep.GroupID, dep.ArtifactID, dep.Scope)
			skippedCount++
			continue
		}

		// Validate required fields
		if dep.GroupID == "" || dep.ArtifactID == "" {
			fmt.Println("  Skipping dependency with empty groupId or artifactId")
			skippedCount++
			continue
		}

		// Determine version to use
		version := dep.Version
		if version == "" {
			version = "latest"
		}

		fmt.Printf("  Installing dependency: %s:%s:%s\n", dep.GroupID, dep.ArtifactID, version)

		// Download as transitive dependency (updatePackageKpm = false, so it only goes to resource.kpm)
		DownloadMavenInternal(dep.GroupID, dep.ArtifactID, version, false, resourcedata, -1, false)
		installedCount++
	}

	fmt.Printf("Info: finished downloading POM dependencies (%d installed, %d skipped)\n", installedCount, skippedCount)
}


