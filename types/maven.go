package types

import (
	"encoding/xml"
	"fmt"
	"kpm/env"
	"strings"

	"github.com/scagogogo/sonatype-central-sdk/pkg/api"
)

type Mavenurl struct {
	Group    string
	Artifact string
	Version  string
}
type Metadata struct {
	XMLName    xml.Name   `xml:"metadata"`
	GroupID    string     `xml:"groupId"`
	ArtifactID string     `xml:"artifactId"`
	Versioning Versioning `xml:"versioning"`
}

// BuildPath generates the repository path for the given package format (e.g., "jar", "pom")
func (m Mavenurl) BuildPath(pkgFormat string) string {
	fmt.Println(m.BuildPath(pkgFormat))
	return api.BuildArtifactPath(m.Group, m.Artifact, m.Version, pkgFormat)
}

// BuildLatestPath generates the path using a specified version
func (m Mavenurl) BuildLatestPath(version, pkgFormat string) string {
	return api.BuildArtifactPath(m.Group, m.Artifact, version, pkgFormat)
}
func (m Mavenurl) GlobalPath(version, pkgFormat string) string {
	if env.DetectOS() == "windows"{
	return fmt.Sprintf("%s\\%s\\%s\\%s\\%s-%s.%s",env.Env.Windows, m.Group,m.Artifact,version,m.Artifact,version,pkgFormat)
	}else if env.DetectOS()== "linux"{
			return fmt.Sprintf("%s/%s/%s/%s/%s-%s.%s",env.Env.Linux, m.Group,m.Artifact,version,m.Artifact,version,pkgFormat)
	}else if env.DetectOS()== "darwin"{
			return fmt.Sprintf("%s/%s/%s/%s/%s-%s.%s", env.Env.Darwin, m.Group,m.Artifact,version,m.Artifact,version,pkgFormat)
	}
	
	return ""
}
func (m Mavenurl) LocalMavenMetaData() string {
	if env.DetectOS() == "windows"{
	return fmt.Sprintf("%s\\%s\\%s\\maven-metadata.xml",env.Env.Windows, m.Group,m.Artifact)
	}else if env.DetectOS()== "linux"{
			return fmt.Sprintf("%s/%s/%s/maven-metadata.xml",env.Env.Linux, m.Group,m.Artifact)
	}else if env.DetectOS()== "darwin"{
			return fmt.Sprintf("%s/%s/%s/maven-metadata.xml", env.Env.Darwin, m.Group,m.Artifact)
	}
	
	return ""
}
func (m Mavenurl) LocalPath(version, pkgFormat string) string {
	if env.DetectOS() == "windows"{
	return fmt.Sprintf(".\\libs\\%s\\%s\\%s\\%s-%s.%s", m.Group,m.Artifact,version,m.Artifact,version,pkgFormat)
	}else {return fmt.Sprintf("./libs/%s/%s/%s/%s-%s.%s/", m.Group,m.Artifact,version,m.Artifact,version,pkgFormat)
	}
	
}
type Versioning struct {
	Latest      string   `xml:"latest"`
	Release     string   `xml:"release"`
	Versions    []string `xml:"versions>version"`
	LastUpdated string   `xml:"lastUpdated"`
}

func (m Mavenurl) MetadataUrl() string {
	groupPath := strings.ReplaceAll(m.Group, ".", "/")
	return fmt.Sprintf("https://repo1.maven.org/maven2/%s/%s/maven-metadata.xml", groupPath, m.Artifact)
}
func (m Mavenurl) PomUrl() string {
	groupPath := strings.ReplaceAll(m.Group, ".", "/")
	return fmt.Sprintf(
		"https://repo1.maven.org/maven2/%s/%s/%s/%s-%s.pom",
		groupPath,
		m.Artifact,
		m.Version,
		m.Artifact,
		m.Version,
	)
}