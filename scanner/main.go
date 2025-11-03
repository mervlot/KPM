package scanner

import (
	"encoding/json"
	"fmt"
	"kpm/libscanner"
)

func Scanner(root string) {

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
			Pkg.Dependencies[imp] = "0" // set version "0" for now
		}
	}
	ReadFile("package.kpm", Pkg)
	data, err := json.MarshalIndent(Pkg, "", "  ")
	libscanner.WriteToJson(data)
}
