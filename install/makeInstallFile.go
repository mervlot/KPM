package install

import (
	"fmt"
	"kpm/types"
)

// helper to get pointer from string
func ptr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// findResourceIndex returns the index of a resource matching group, name and (optionally) version.
// If version is empty, it matches any version for the group+name pair.
func findResourceIndex(resourcedata *types.ResourceFile, group, name, version string) int {
	for i, r := range resourcedata.Resources {
		if r.Name != name {
			continue
		}

		g := ""
		if r.Group != nil {
			g = *r.Group
		}

		v := ""
		if r.Version != nil {
			v = *r.Version
		}

		if g == group {
			if version == "" || v == version {
				return i
			}
		}
	}
	return -1
}

// AppendResource appends a resource to a new ResourceFile and saves it
func AppendResource(
	resourcedata *types.ResourceFile,
	Group string,
	Name string,
	Version string,
	Source string,
	Type string,
	Domain string,
	Path string,
	LPath string,
	GPath string,
	URL string,
	Hash string,
) {
	// Create a new ResourceFile

	// Append the resource
	resourcedata.Resources = append(resourcedata.Resources, types.Resource{
		Group:   ptr(Group),
		Name:    Name,
		Version: ptr(Version),
		Source:  Source,
		Type:    Type,
		Domain:  ptr(Domain),
		Path:    ptr(Path),
		LPath:   LPath,
		GPath:   ptr(GPath),
		URL:     ptr(URL),
		Hash:    Hash,
	})

	// Save the ResourceFile
	if err := SaveResourceFile(resourcedata); err != nil {
		fmt.Println("Error saving resource file:", err)
	} else {
		fmt.Println("Resource appended successfully!")
	}
}

func UpdateResource(
	resourcedata *types.ResourceFile,
	Group string,
	Name string,
	Version string,
	Source string,
	Type string,
	Domain string,
	Path string,
	LPath string,
	GPath string,
	URL string,
	Hash string,
	index int,
) {
	// Create a new ResourceFile

	// Append the resource

	resourcedata.Resources[index] = types.Resource{
		Group:   ptr(Group),
		Name:    Name,
		Version: ptr(Version),
		Source:  Source,
		Type:    Type,
		Domain:  ptr(Domain),
		Path:    ptr(Path),
		LPath:   LPath,
		GPath:   ptr(GPath),
		URL:     ptr(URL),
		Hash:    Hash,
	}

	// Save the ResourceFile
	if err := SaveResourceFile(resourcedata); err != nil {
		fmt.Println("Error saving resource file:", err)
	} else {
		fmt.Println("Resource appended successfully!")
	}
}
