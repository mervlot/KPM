package install

import (
	"fmt"
	"os"
	"strings"

	"kpm/readline"
	"kpm/types"
	"kpm/utils"
)

// per-run visited cache
var visited map[string]bool

// DownloadMaven entry point
func DownloadMaven(group, artifact, version string, update bool, resourcedata *types.ResourceFile, index int) {
	visited = make(map[string]bool)
	DownloadMavenInternal(group, artifact, version, update, resourcedata, index, true, false, "", "")
	UpdatePackageDependencyFromMaven(group, artifact, version)
}

// internal resolver
func DownloadMavenInternal(
	group, artifact, version string,
	update bool,
	resourcedata *types.ResourceFile,
	index int,
	updatePackageKpm bool,
	dependency bool,
	scope, parent string,
) {
	if group == "" || artifact == "" {
		fmt.Println("error: missing GroupID or ArtifactID")
		return
	}

	maven := types.Mavenurl{
		Group:    group,
		Artifact: artifact,
		Version:  version,
	}

	// -----------------------------
	// METADATA
	// -----------------------------
	if !utils.IsFile(maven.LocalMavenMetaData()) {
		if err := DownloadFile(maven.MetadataUrl(), maven.LocalMavenMetaData()); err != nil {
			fmt.Printf("metadata download failed %s:%s\n", group, artifact)
			return
		}
	}

	data, err := os.ReadFile(maven.LocalMavenMetaData())
	if err != nil {
		fmt.Println("metadata read error:", err)
		return
	}

	meta, err := UnmarshalMavenXml(data)
	if err != nil {
		fmt.Println("metadata parse error:", err)
		return
	}

	// resolve latest if needed
	if version == "" {
		version = meta.Versioning.Latest
		maven.Version = version
	}

	key := group + ":" + artifact + ":" + version + ":" + parent
	if visited[key] {
		return
	}
	visited[key] = true

	fmt.Println("processing:", key)

	// -----------------------------
	// POM
	// -----------------------------
	pomPath := maven.GlobalPath(version, "pom")

	if !utils.IsFile(pomPath) {
		if err := DownloadFile(maven.PomUrl(), pomPath); err != nil {
			fmt.Println("POM download failed:", err)
			return
		}
	}

	pomData, err := os.ReadFile(pomPath)
	if err != nil {
		fmt.Println("POM read error:", err)
		return
	}

	pom, err := UnmarshalMavenPom(pomData)
	if err != nil {
		fmt.Println("POM parse error:", err)
		return
	}

	if pom.Packaging == "" {
		pom.Packaging = "jar"
	}

	// -----------------------------
	// RESOURCE TRACKING
	// -----------------------------
	if exists(resourcedata, group, artifact, version) {
		fmt.Println("skip duplicate:", group, artifact, version)
	} else {
		if idx, ok := findSameArtifact(resourcedata, group, artifact); ok {
			// same artifact exists but different version → update instead of duplicate
			fmt.Println("upgrading:", group, artifact, "to", version)

			resourcedata.Resources[idx].Version = ptr(version)

			resourcedata.Resources[idx].Path = ptr(fmt.Sprintf(
				"io/%s/%s/%s/%s-%s.%s",
				strings.ReplaceAll(group, ".", "/"),
				artifact,
				version,
				artifact,
				version,
				pom.Packaging,
			))

			resourcedata.Resources[idx].LPath = maven.LocalPath(version, pom.Packaging)
			resourcedata.Resources[idx].GPath = ptr(maven.GlobalPath(version, pom.Packaging))
			resourcedata.Resources[idx].URL = ptr(fmt.Sprintf(
				"https://repo1.maven.org/maven2/%s/%s/%s/%s-%s.%s",
				strings.ReplaceAll(group, ".", "/"),
				artifact,
				version,
				artifact,
				version,
				pom.Packaging,
			))
		} else {
			AppendResource(
				resourcedata,
				group,
				artifact,
				version,
				"maven",
				pom.Packaging,
				"https://repo1.maven.org/maven2/",
				fmt.Sprintf(
					"https://repo1.maven.org/maven2/%s/%s/%s/%s-%s.%s",
					strings.ReplaceAll(group, ".", "/"),
					artifact,
					version,
					artifact,
					version,
					pom.Packaging,
				),
				maven.LocalPath(version, pom.Packaging),
				maven.GlobalPath(version, pom.Packaging),
				maven.BuildPath(pom.Packaging),
				"xx",
				scope,
				parent,
			)
		}
	}

	// -----------------------------
	// DEPENDENCIES
	// -----------------------------
	for _, v := range pom.Dependencies {

		if v.GroupID == "" || v.ArtifactID == "" {
			continue
		}

		// -------------------------
		// VERSION RESOLUTION
		// -------------------------
		if v.Version == "" {

			fmt.Println("\nmissing version for:", v.GroupID, v.ArtifactID)
			fmt.Println("format: group:artifact:version")
			fmt.Println("l = latest | r = retry | q = quit")

			for {
				input, _ := readline.Main()
				input = strings.TrimSpace(strings.ToLower(input))

				// quit everything
				if input == "q" {
					fmt.Println("stopping install")
					os.Exit(0)
				}

				// latest
				if input == "l" {
					v.Version = meta.Versioning.Latest
					break
				}

				// retry prompt
				if input == "r" {
					continue
				}

				// manual input
				parts := strings.Split(input, ":")
				if len(parts) == 3 {
					v.GroupID = parts[0]
					v.ArtifactID = parts[1]
					v.Version = parts[2]
					break
				}

				fmt.Println("invalid input, try again")
			}
		}

		// FIX: ensure recursion uses resolved version
		if v.Version == "" {
			continue
		}

		// avoid self-loop
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
			true,
			v.Scope,
			group+":"+artifact,
		)
	}

	// -----------------------------
	// ARTIFACT DOWNLOAD
	// -----------------------------
	if pom.Packaging != "pom" && !utils.IsFile(maven.GlobalPath(version, pom.Packaging)) {
		if err := DownloadFile(maven.BuildPath(pom.Packaging), maven.GlobalPath(version, pom.Packaging)); err != nil {
			fmt.Println("artifact download failed:", err)
		}
	}
}

func safeString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func exists(res *types.ResourceFile, g, n, v string) bool {
	for _, r := range res.Resources {
		if r.Name != n {
			continue
		}
		if safeString(r.Group) == g && safeString(r.Version) == v {
			return true
		}
	}
	return false
}

func findSameArtifact(res *types.ResourceFile, g, n string) (int, bool) {
	for i, r := range res.Resources {
		if r.Group != nil && r.Name == n {
			if *r.Group == g {
				return i, true
			}
		}
	}
	return -1, false
}
