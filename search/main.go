package search

import (
	"context"
	"fmt"

	"github.com/scagogogo/sonatype-central-sdk/pkg/api"
)

func searchMavenCentral(group, artifact string, rows int) error {
	client := api.NewClient()

	// Use the SDK method — for example, search by group + artifact
	res, err := client.SearchByArtifactId(context.Background(), artifact, rows)
	if err != nil {
		return err
	}

	for _, a := range res {
		fmt.Printf("Group: %s, \nArtifact: %s, \nLatest Version: %s, \nTags: %v, \nEc: %v, \nID: %s, \nPackaging: %s, \nRepositoryID: %s, \nText: %s, \nVersionCount: %v, \nTimestamp: %v\n",
			a.GroupId, a.ArtifactId, a.LatestVersion, a.Tags, a.Ec, a.ID, a.Packaging, a.RepositoryID, a.Text, a.VersionCount, a.Timestamp)
		path := api.BuildArtifactPath(a.GroupId, a.ArtifactId, a.LatestVersion, "jar")
		client := api.NewClient()
		file, err := client.Download(context.Background(),path)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(file))
	}

	return nil
}

func Main() {
	if err := searchMavenCentral("org.jetbrains.kotlinx", "kotlinx-coroutines-core", 5); err != nil {
		print(err)
	}
}
