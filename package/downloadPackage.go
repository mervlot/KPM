package install

import (
	"fmt"
	"kpm/types"
	"kpm/utils"
	"os"
)

// Declare the map, but initialize it per-install to avoid crossover
var visited map[string]bool

// DownloadMaven downloads a Maven artifact and updates both resource.kpm and package.kpm
func DownloadMaven(group, artifact, version string, update bool, resourcedata *types.ResourceFile, index int) {
	fmt.Printf("🚀 Starting resolution for: %s:%s:%s\n", group, artifact, version)
	visited = make(map[string]bool)
	DownloadMavenInternal(group, artifact, version, update, resourcedata, index, true)
	fmt.Println("✨ Resolution complete.")
}

// DownloadMavenInternal is the internal implementation of DownloadMaven
func DownloadMavenInternal(group, artifact, version string, update bool, resourcedata *types.ResourceFile, index int, updatePackageKpm bool) {
	if group == "" || artifact == "" {
		fmt.Printf("❌ Error: missing GroupID or ArtifactID in download request\n")
		return
	}

	var maven types.Mavenurl = types.Mavenurl{
		Group:    group,
		Artifact: artifact,
		Version:  version,
	}

	// 1. Metadata Resolution
	if !utils.IsFile(maven.LocalMavenMetaData()) {
		fmt.Printf("🌐 Metadata missing locally. Fetching online for %s:%s\n", group, artifact)
		err := DownloadFile(maven.MetadataUrl(), maven.LocalMavenMetaData())
		if err != nil {
			fmt.Printf("❌ Failed downloading metadata for %s:%s - %v\n", group, artifact, err)
			return
		}
	} else {
		fmt.Printf("🔍 Metadata found locally for %s:%s\n", group, artifact)
	}

	data, err := os.ReadFile(maven.LocalMavenMetaData())
	if err != nil {
		fmt.Printf("❌ Failed reading metadata for %s:%s - %v\n", group, artifact, err)
		return
	}

	meta, err := UnmarshalMavenXml(data)
	if err != nil {
		fmt.Printf("❌ Failed parsing XML metadata for %s:%s - %v\n", group, artifact, err)
		return
	}

	if version == "" {
		version = meta.Versioning.Latest
		maven.Version = version
		fmt.Printf("📌 Resolved latest version: %s\n", version)
	}

	key := group + ":" + artifact + ":" + version
	if visited[key] {
		return // Already processed in this run
	}
	visited[key] = true
	fmt.Printf("\n📦 Processing package: %s\n", key)

	// 2. POM Resolution
	if !utils.IsFile(maven.GlobalPath(version, "pom")) {
		fmt.Printf("🌐 POM missing locally. Fetching online: %s\n", key)
		err := DownloadFile(maven.PomUrl(), maven.GlobalPath(version, "pom"))
		if err != nil {
			fmt.Printf("❌ Failed downloading POM for %s - %v\n", key, err)
			return
		}
	} else {
		fmt.Printf("🔍 POM already local: %s\n", key)
	}

	fmt.Printf("📄 Reading and parsing POM: %s\n", key)
	pomData, err := os.ReadFile(maven.GlobalPath(version, "pom"))
	if err != nil {
		fmt.Printf("❌ Failed reading POM for %s - %v\n", key, err)
		return
	}

	pom, err := UnmarshalMavenPom(pomData)
	if err != nil {
		fmt.Printf("❌ Failed parsing POM for %s - %v\n", key, err)
		return
	}

	if pom.Packaging == "" {
    pom.Packaging = "jar"
}

	// 3. Artifact Download
	if pom.Packaging != "pom" {
		if !utils.IsFile(maven.GlobalPath(version, pom.Packaging)) {
			fmt.Printf("🌐 Artifact missing locally. Fetching %s file for %s\n", pom.Packaging, key)
			err := DownloadFile(maven.BuildPath(pom.Packaging), maven.GlobalPath(version, pom.Packaging))
			if err != nil {
				fmt.Printf("❌ Failed downloading %s artifact for %s - %v\n", pom.Packaging, key, err)
				return
			}
		} else {
			fmt.Printf("✅ Artifact already local: %s.%s\n", key, pom.Packaging)
		}
	}

	// 4. Transitive Dependencies
	if len(pom.Dependencies) > 0 {
		fmt.Printf("🔀 Walking %d dependencies for %s\n", len(pom.Dependencies), key)
	}
	for _, v := range pom.Dependencies {
		if v.GroupID == "" || v.ArtifactID == "" {
			continue
		}
		if v.Version == "" {
			fmt.Printf("⚠️  Skipping dependency %s:%s (missing version - property resolution needed)\n", v.GroupID, v.ArtifactID)
			continue
		}
		if v.GroupID == group && v.ArtifactID == artifact && v.Version == version {
			continue
		}

		DownloadMavenInternal(
			v.GroupID,
			v.ArtifactID,
			v.Version,
			false,
			resourcedata,
			-1,
			false,
		)
	}

	// 5. Copy JAR to local project directory
	if !utils.IsFile(maven.LocalPath(version, pom.Packaging)) {
		utils.CopyFile(maven.GlobalPath(version, pom.Packaging), maven.LocalPath(version, pom.Packaging))
	}
}