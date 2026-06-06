package sync

import (
	"encoding/json"
	"fmt"
	"kpm/types"
	"os"
	"path/filepath"
	"strings"
)

func Main() {
	fmt.Println("syncing...")

	libsDir := "./libs"

	var rf types.ResourceFile
	rf.Version = "1.0.0"
	rf.Resources = []types.Resource{}

	err := filepath.Walk(libsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if !(strings.HasSuffix(info.Name(), ".jar") || strings.HasSuffix(info.Name(), ".war")) {
			return nil
		}

		// normalize to clean forward-slash path
		rel := filepath.ToSlash(path)

		// ensure we always start from ./libs
		rel = strings.TrimPrefix(rel, "./")

		parts := strings.Split(rel, "/")

		// expected: libs/group/artifact/version/file
		if len(parts) < 5 || parts[0] != "libs" {
			return nil
		}

		group := parts[1]
		artifact := parts[2]
		version := parts[3]

		fileType := "jar"
		if strings.HasSuffix(info.Name(), ".war") {
			fileType = "war"
		}

		rf.Resources = append(rf.Resources, types.Resource{
			Group:   &group,
			Name:    artifact,
			Version: &version,
			Source:  "local",
			Type:    fileType,
			LPath:   "./" + rel, // force ./ prefix cleanly
		})

		return nil
	})

	if err != nil {
		fmt.Println("scan error:", err)
		return
	}

	data, _ := json.MarshalIndent(rf, "", "  ")
	_ = os.WriteFile("./resource.kpm", data, 0644)

	fmt.Println("resource.kpm rebuilt ✔")
}