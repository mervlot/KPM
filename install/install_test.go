package install

import (
	"fmt"
	"kpm/types"
	"testing"
)

func TestGetMavenMetadata(t *testing.T) {
	maven := types.Mavenurl{
		Group:    "org/jetbrains/kotlinx",
		Artifact: "kotlinx-coroutines-core",
		Version:  "1.10.2",
	}

	// Build URL
	urls := fmt.Sprintf("https://repo1.maven.org/maven2/%s/%s/%s/%s-%s.jar",
		maven.Group, maven.Artifact, maven.Version, maven.Artifact, maven.Version)

	t.Logf("URL: %s", urls)
	url := "https://repo1.maven.org/maven2/org/jetbrains/kotlinx/kotlinx-coroutines-core/maven-metadata.xml"
	meta, err := GetMavenMetadata(url)
	if err != nil {
		t.Fatalf("failed to fetch metadata: %v", err)
	}

	DownloadJar(
		"https://repo1.mn.org/maven2/org/jetbrains/kotlinx/kotlinx-coroutines-core/1.10.2/kotlinx-coroutines-core-1.10.2.jar",
		"./libs/kotlinx-coroutines-core-1.10.2.jar",
		false,
	)

	t.Logf("GroupID: %s", meta.GroupID)
	t.Logf("ArtifactID: %s", meta.ArtifactID)
	t.Logf("Latest: %s", meta.Versioning.Latest)
	t.Logf("Release: %s", meta.Versioning.Release)
	t.Logf("All versions count: %d", len(meta.Versioning.Versions))
}
