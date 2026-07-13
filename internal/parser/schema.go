package parser

import "encoding/xml"

// POM mirrors the subset of the Maven POM schema KPM understands:
// coordinates, packaging, properties, parent inheritance, dependencyManagement
// (including <scope>import</scope> BOM entries), dependencies (with
// exclusions/optional/classifier), and repositories.
type POM struct {
	XMLName    xml.Name `xml:"project"`
	GroupID    string   `xml:"groupId"`
	ArtifactID string   `xml:"artifactId"`
	Version    string   `xml:"version"`
	Packaging  string   `xml:"packaging"`
	Name       string   `xml:"name"`

	Parent *ParentRef `xml:"parent"`

	Properties Properties `xml:"properties"`

	DependencyManagement struct {
		Dependencies []RawDependency `xml:"dependencies>dependency"`
	} `xml:"dependencyManagement"`

	Dependencies []RawDependency `xml:"dependencies>dependency"`

	Repositories []RawRepository `xml:"repositories>repository"`

	Modules []string `xml:"modules>module"` // multi-module aggregator, future use
}

type ParentRef struct {
	GroupID      string `xml:"groupId"`
	ArtifactID   string `xml:"artifactId"`
	Version      string `xml:"version"`
	RelativePath string `xml:"relativePath"`
}

// Properties captures arbitrary <properties><key>value</key>...</properties>
// as a map, since the element names are user-defined.
type Properties struct {
	Entries map[string]string
}

func (p *Properties) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	p.Entries = map[string]string{}
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			var val string
			if err := d.DecodeElement(&val, &t); err != nil {
				return err
			}
			p.Entries[t.Name.Local] = val
		case xml.EndElement:
			if t.Name == start.Name {
				return nil
			}
		}
	}
}

type RawExclusion struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
}

type RawDependency struct {
	GroupID    string         `xml:"groupId"`
	ArtifactID string         `xml:"artifactId"`
	Version    string         `xml:"version"`
	Scope      string         `xml:"scope"`
	Classifier string         `xml:"classifier"`
	Type       string         `xml:"type"`
	Optional   string         `xml:"optional"`
	Exclusions []RawExclusion `xml:"exclusions>exclusion"`
}

type RawRepository struct {
	ID  string `xml:"id"`
	URL string `xml:"url"`
}

// Metadata is maven-metadata.xml (moved here from the old types package;
// see internal/metadata for fetch/cache logic that produces one of these).
type Metadata struct {
	XMLName    xml.Name `xml:"metadata"`
	GroupID    string   `xml:"groupId"`
	ArtifactID string   `xml:"artifactId"`
	Versioning struct {
		Latest      string   `xml:"latest"`
		Release     string   `xml:"release"`
		Versions    []string `xml:"versions>version"`
		LastUpdated string   `xml:"lastUpdated"`
		Snapshot    struct {
			Timestamp   string `xml:"timestamp"`
			BuildNumber int    `xml:"buildNumber"`
		} `xml:"snapshot"`
	} `xml:"versioning"`
}

func ParseMetadata(data []byte) (*Metadata, error) {
	var m Metadata
	if err := xml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}