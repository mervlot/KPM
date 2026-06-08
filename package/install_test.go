package install

import (
	"fmt"
	"kpm/types"
	"testing"
)
var mavenMeta types.Metadata
func TestXmlParsing(t *testing.T) {
	url := "https://repo1.maven.org/maven2/com/google/guava/guava/maven-metadata.xml"
	mavenMeta,_ := GetMavenMetadata(url)
	DownloadFile(url, "./s.xml")
	fmt.Println(mavenMeta.GroupID)
}