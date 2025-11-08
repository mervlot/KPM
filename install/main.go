package install

import (
	"encoding/json"
	"fmt"
	"kpm/libscanner"
	"kpm/types"
)

// NewResourceFile creates an empty ResourceFile with default version
func NewResourceFile() *types.ResourceFile {
	return &types.ResourceFile{
		Version:   "0.0.1",
		Resources: []types.Resource{},
	}
}

// SaveResourceFile marshals the ResourceFile to JSON and writes it
func SaveResourceFile(rf *types.ResourceFile) error {
	resourceBytes, err := json.MarshalIndent(rf, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling resource file to JSON: %w", err)
	}

	err = libscanner.WriteToJson(resourceBytes, "resource.kpm")
	if err != nil {
		return fmt.Errorf("error writing resource file to package.kpm: %w", err)
	}

	return nil
}

func Main() {
	resourcedata := NewResourceFile()
	resourcedata.Version = "1.0.0"
	AppendResource(
		resourcedata,
		"org.jetbrains.kotlinx",           // Group
		"kotlinx-coroutines-core",         // Name
		"1.7.3",                           // Version
		"maven",                           // Source
		"jar",                             // Type
		"https://repo1.maven.org/maven2/", // Domain
		"org/jetbrains/kotlinx/kotlinx-coroutines-core/1.7.3/kotlinx-coroutines-core-1.7.3.jar",             // Path
		"./libs/kotlinx-coroutines-core-1.7.3.jar",                                                          // LPath
		"~/.kpm/repo/org/jetbrains/kotlinx/kotlinx-coroutines-core/1.7.3/kotlinx-coroutines-core-1.7.3.jar", // GPath
		"",                              // URL is nil for Maven
		"sha256:82af21f1c0e8ce74c5c...", // Hash
	)
	if err := SaveResourceFile(resourcedata); err != nil {
		fmt.Println("Error saving resource file:", err)
	} else {
		fmt.Println("Resource file saved successfully!")
	}
}
