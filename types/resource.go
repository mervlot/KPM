package types

import "encoding/xml"

type Resource struct {
	Group   *string `json:"group,omitempty"`
	Name    string  `json:"name"`
	Version *string `json:"version,omitempty"`
	Source  string  `json:"source"`
	Type    string  `json:"type"`
	Domain  *string `json:"domain,omitempty"`
	Path    *string `json:"path,omitempty"`
	LPath   string  `json:"lpath"`
	GPath   *string `json:"gpath,omitempty"`
	URL     *string `json:"url,omitempty"`
	Hash    string  `json:"hash"`
}

type ResourceFile struct {
	Version   string     `json:"kpm version"`
	Resources []Resource `json:"resources"`
}



type PomDependency struct {
	XMLName      xml.Name     `xml:"project"`
	ModelVersion string       `xml:"modelVersion"`
	GroupID      string       `xml:"groupId"`
	ArtifactID   string       `xml:"artifactId"`
	Packaging    string       `xml:"packaging"`
	Description  string       `xml:"description"`
	URL          string       `xml:"url"`
	Version      string       `xml:"version"`
	Name         string       `xml:"name"`

	Licenses     []License     `xml:"licenses>license"`
	Organization Organization  `xml:"organization"`
	SCM          SCM           `xml:"scm"`
	Developers   []Developer   `xml:"developers>developer"`
	Dependencies []Dependency  `xml:"dependencies>dependency"`
}

type License struct {
	Name         string `xml:"name"`
	URL          string `xml:"url"`
	Distribution string `xml:"distribution"`
}

type Organization struct {
	Name string `xml:"name"`
	URL  string `xml:"url"`
}

type SCM struct {
	URL                 string `xml:"url"`
	Connection          string `xml:"connection"`
	DeveloperConnection string `xml:"developerConnection"`
}

type Developer struct {
	ID    string `xml:"id"`
	Name  string `xml:"name"`
	URL   string `xml:"url"`
	Email string `xml:"email"`
}

type Dependency struct {
	GroupID   string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version   string `xml:"version"`
	Scope     string `xml:"scope,omitempty"`
}