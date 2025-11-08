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
		Path:    Path,
		LPath:   LPath,
		GPath:   GPath,
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
