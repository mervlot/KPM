package install

import (
	"fmt"
	"kpm/types"
	"kpm/utils"
	"os"
)

func pomDep(visited map[string]bool, group, artifact, version string, resourcedata *types.ResourceFile) {

	key := group + ":" + artifact + ":" + version
	if visited[key] {
		return
	}
	visited[key] = true
	maven := types.Mavenurl{
		Group:    group,
		Artifact: artifact,
		Version:  version,
	}

	if !utils.IsFile(maven.GlobalPath(version, "pom")) {
		err := DownloadFile(maven.PomUrl(), maven.GlobalPath(version, "pom"))
		if err != nil {
			fmt.Println("pom download", err, maven.PomUrl())
			return
		}
	}
	pomData, err := os.ReadFile(maven.GlobalPath(version, "pom"))
	if err != nil {
		return
	}

	pom, err := UnmarshalMavenPom(pomData)
	if err != nil {
		return
	}
	for _, v := range pom.Dependencies {
		pomDep(visited, v.GroupID, v.ArtifactID, v.Version, resourcedata)
	}
	DownloadMavenInternal(group, artifact, version, false, resourcedata, -1, true)

}
