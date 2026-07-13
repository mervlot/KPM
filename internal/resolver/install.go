package resolver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"kpm/internal/downloader" // Adjust this import path to match your actual module name
)

// fetchArtifact is the core "offline-first" function.
// It checks the local cache first. If missing, it downloads from the network and saves to cache.
func fetchArtifact(groupID, artifactID, version, ext string, client *downloader.Client) ([]byte, string, error) {
	// 1. OFFLINE-FIRST: Check local cache first
	groupPath := filepath.Join("libs", filepath.FromSlash(groupID))
	artifactPath := filepath.Join(groupPath, artifactID, version)
	fileName := fmt.Sprintf("%s-%s.%s", artifactID, version, ext)
	localPath := filepath.Join(artifactPath, fileName)

	// Check if file exists and is not empty (size > 0)
	info, err := os.Stat(localPath)
	if err == nil && !info.IsDir() && info.Size() > 0 {
		// Found in cache! Read from disk. Zero network requests.
		data, err := os.ReadFile(localPath)
		if err == nil {
			return data, localPath, nil
		}
		// If reading fails (e.g., corrupted file), fall through to network download
	}

	// 2. NETWORK FALLBACK: Only hit the network if not cached
	mavenGroupPath := strings.ReplaceAll(groupID, ".", "/")
	url := fmt.Sprintf("https://repo1.maven.org/maven2/%s/%s/%s/%s-%s.%s",
		mavenGroupPath, artifactID, version, artifactID, version, ext)

	data, err := client.Get(url)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch %s: %w", url, err)
	}

	// 3. SAVE TO CACHE: Persist to disk for future offline-first use
	if err := os.MkdirAll(artifactPath, 0755); err != nil {
		return nil, "", fmt.Errorf("failed to create cache dir %s: %w", artifactPath, err)
	}

	if err := os.WriteFile(localPath, data, 0644); err != nil {
		return nil, "", fmt.Errorf("failed to write to cache %s: %w", localPath, err)
	}

	return data, localPath, nil
}

// ResolveAndInstall handles the dependency resolution and installation process.
// (Adapt this to match your existing function signatures)
func ResolveAndInstall(dependencies []Dependency) error {
	client := downloader.New()
	
	// Optional: Set rate limit explicitly (500ms = 2 req/sec)
	downloader.SetGlobalPace(500 * time.Millisecond)

	var jobs []downloader.Job

	for _, dep := range dependencies {
		dep := dep // capture loop variable
		job := downloader.Job{
			Name: fmt.Sprintf("%s:%s:%s", dep.GroupID, dep.ArtifactID, dep.Version),
			Run: func() error {
				// Fetch POM first (to read transitive dependencies)
				_, pomPath, err := fetchArtifact(dep.GroupID, dep.ArtifactID, dep.Version, "pom", client)
				if err != nil {
					return fmt.Errorf("failed to get POM for %s: %w", job.Name, err)
				}
				fmt.Printf("✅ Resolved POM: %s\n", pomPath)

				// Fetch JAR
				_, jarPath, err := fetchArtifact(dep.GroupID, dep.ArtifactID, dep.Version, "jar", client)
				if err != nil {
					return fmt.Errorf("failed to get JAR for %s: %w", job.Name, err)
				}
				fmt.Printf("✅ Installed JAR: %s\n", jarPath)

				return nil
			},
		}
		jobs = append(jobs, job)
	}

	// Execute sequentially (no goroutines) to respect rate limits
	fmt.Println("⏳ Installing dependencies sequentially (offline-first enabled)...")
	errs := downloader.RunPool(jobs, 1) // concurrency param ignored, forces sequential
	
	if len(errs) > 0 {
		fmt.Println("✖ Resolution failed:")
		for _, err := range errs {
			fmt.Printf("  - %v\n", err)
		}
		return fmt.Errorf("installation failed with %d errors", len(errs))
	}

	fmt.Println("🎉 All dependencies installed successfully!")
	return nil
}

// Dependency represents a parsed dependency from package.kpm
type Dependency struct {
	GroupID    string
	ArtifactID string
	Version    string
}