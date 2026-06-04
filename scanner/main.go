package scanner

import (
	"encoding/json"
	"fmt"
	"kpm/libscanner"
	"os"
)

func Scanner(root string) {
	var oldPkg PackageFile
	_, statErr := os.Stat("package.kpm")
	if statErr == nil {
		bytes, err := os.ReadFile("package.kpm")
		if err != nil {
			fmt.Println("we where unable to read the file", err)
			return
		}
		json.Unmarshal(bytes, &oldPkg)
	} else if !os.IsNotExist(statErr) {
		fmt.Println("Something went wrong when checking package.kpm please run `kpm init` to init a proj. More info: ", statErr)
		return
	}
	err, ktFiles := libscanner.KtFinder(root)
	if err != nil {
		return
	}
	if len(ktFiles) == 0 {
		fmt.Println("No Kotlin (.kt) files found.")
		return
	}
	for _, filePath := range ktFiles {
		imports := libscanner.ReadImports(filePath)
		for _, imp := range imports {
			oldPkg.Dependencies[imp] = "0" // set version "0" for now
		}
	}
	data, err := json.MarshalIndent(oldPkg, "", "  ")
	libscanner.WriteToJson(data, "package.kpm")
}
