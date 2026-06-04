package sync

import (
	"encoding/json"
	"fmt"
	"kpm/scanner"
	"kpm/types"
	"os"
)

func Main() {
	var resourcefile types.ResourceFile
	var packagefile scanner.PackageFile
	fmt.Println("begining syncing ...")
	resource, err := os.ReadFile("resource.kpm")

	if err != nil {
		fmt.Println(err)
	}

	pkg, err := os.ReadFile("package.kpm")
	if err != nil {
		fmt.Println(err)
	}

	json.Unmarshal(resource, &resourcefile)
	json.Unmarshal(pkg, &packagefile)
	for i,v := range packagefile.Dependencies{
		fmt.Println(i,v)
	}
}
