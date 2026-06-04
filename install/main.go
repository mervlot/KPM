package install

import (
	"encoding/json"
	"fmt"
	"kpm/libscanner"
	"kpm/types"
	"os"
	"strings"
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
		return fmt.Errorf("error writing resource file to resource.kpm: %w", err)
	}

	return nil
}

var resourcedata = NewResourceFile()

func Main(update bool, params ...string) {
	var oldfile types.ResourceFile

	// Load existing resource file if it exists
	if _, err := os.Stat("resource.kpm"); err == nil {
		ReadFile("resource.kpm", &oldfile)
		resourcedata.Version = oldfile.Version
		resourcedata.Resources = oldfile.Resources
	}

	// Parse optional leading flags for install mode: -m (maven), -u (url).
	mode := "auto" // auto-detect by content if not specified
	i := 0
	for i < len(params) {
		p := params[i]
		if strings.HasPrefix(p, "-") {
			// flag(s) encountered; support -m and -u. Multiple flags allowed but last wins.
			switch p {
			case "-m", "--maven":
				mode = "maven"
				i++
				continue
			case "-u", "--url":
				mode = "url"
				i++
				continue
			default:
				// Unknown flag — treat as error and show guidance
				fmt.Println("Unknown install flag:", p, "Supported: -m (maven), -u (url)")
				i++
				continue
			}
		}

		// Process according to selected mode
		if mode == "url" {
			// Expect pairs: name URL
			if i+1 >= len(params) {
				fmt.Println("Missing URL for artifact:", params[i])
				break
			}
			name := params[i]
			url := params[i+1]
			DownloadUrl(name, resourcedata, url, update, -1)
			i += 2
			continue
		}

		// mode == "maven" or auto: try to parse as maven coordinates
		nameOrMaven := params[i]

		// If next param looks like a URL and we're in auto mode, treat current as name and next as URL
		if mode == "auto" && i+1 < len(params) && (strings.HasPrefix(params[i+1], "http://") || strings.HasPrefix(params[i+1], "https://")) {
			url := params[i+1]
			DownloadUrl(nameOrMaven, resourcedata, url, update, -1)
			i += 2
			continue
		}

		// Otherwise, treat as Maven coordinates.
		val := strings.Split(nameOrMaven, ":")
		if len(val) == 2 && update {
			group := val[0]
			artifact := val[1]
			version := "" // let DownloadMaven fetch latest via metadata
			fmt.Println("Detected group:artifact with update flag, fetching version from metadata for", group+":"+artifact)
			DownloadMaven(group, artifact, version, update, resourcedata, -1)
		} else if len(val) == 3 {
			group := val[0]
			artifact := val[1]
			version := val[2]
			if update {
				version = ""
			} else if version == "latest" {
				version = "latest"
			}
			DownloadMaven(group, artifact, version, update, resourcedata, -1)
		} else {
			fmt.Println("Invalid Maven parameter, must be group:artifact:version (or group:artifact when using update):", nameOrMaven)
		}
		i++
	}

	// Update existing resources if -u was used without extra params
	if update && len(params) == 0 {
		for index, v := range resourcedata.Resources {
			if _, err := os.Stat(v.LPath); err == nil {
				fmt.Println("updating:", v.Name)
				switch v.Source {
				case "maven":
					DownloadMaven(*v.Group, v.Name, *v.Version, true, resourcedata, index)
				case "url":
					fmt.Println("Cannot update URL source:", v.Name)
				}
			} else {
				fmt.Println("Package missing, installing latest version:", v.Name)
				switch v.Source {
				case "maven":
					DownloadMaven(*v.Group, v.Name, *v.Version, true, resourcedata, index)
				case "url":
					DownloadUrl(v.Name, resourcedata, *v.URL, true, index)
				}
			}
		}
	}

	resourcedata.Version = "1.0.0"
	if err := SaveResourceFile(resourcedata); err != nil {
		fmt.Println("Error saving resource file:", err)
	} else {
		fmt.Println("Resource file saved successfully!")
	}
}
