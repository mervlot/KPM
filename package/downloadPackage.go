package install

import (
	"fmt"
	"kpm/types"
	"kpm/utils"
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
	if group =="" && artifact == "" {
		fmt.Println("oops this shouldnt happend a function handles this error in the parent")
	}
	var maven types.Mavenurl = types.Mavenurl{
		Group: group,
		Artifact: artifact,
		Version: version,

	}

	if !utils.IsFile(maven.LocalMavenMetaData()) {
		err := DownloadFile(maven.MetadataUrl(),maven.LocalMavenMetaData())
		if err != nil {
			fmt.Println(err)
			return
		}}
		data, err := os.ReadFile(maven.LocalMavenMetaData())
		if err != nil {
			fmt.Println("read maven",err)
			return
		}
		meta, err := UnmarshalMavenXml(data)
		if err != nil {
			fmt.Println("xml download",err)
			return
		}
		if version == ""{
			version = meta.Versioning.Latest
			maven.Version = meta.Versioning.Latest
		}
		if !utils.IsFile(maven.GlobalPath(version,"pom")){
			err:= DownloadFile(maven.PomUrl(),maven.GlobalPath(version,"pom"))
			if err != nil {
			fmt.Println("pom download",err,maven.PomUrl())
			return
		}
		}
		pomData, err := os.ReadFile(maven.GlobalPath(version,"pom"))

		pom, err :=UnmarshalMavenPom(pomData)
		if len(pom.Dependencies) > 0{
			for i,v := range pom.Dependencies{
				
			}
		}

	
	
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

