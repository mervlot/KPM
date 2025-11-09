package types

import (
	"encoding/xml"
	"fmt"
	"strings"
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

func (m Mavenurl) LatestUrl(version string) string {
    groupPath := strings.ReplaceAll(m.Group, ".", "/")
    return fmt.Sprintf("https://repo1.maven.org/maven2/%s/%s/%s/%s-%s.jar", groupPath, m.Artifact, version, m.Artifact, version)
}


func (m Mavenurl) Url() string {
	return fmt.Sprintf("https://repo1.maven.org/maven2/%s/%s/%s/%s-%s.jar",
		m.Group, m.Artifact, m.Version, m.Artifact, m.Version)
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
