package libscanner

import (
	"fmt"
)

var Pkg = PackageFile{
	Dependencies: make(map[string]string),
}

// PackageFile represents the JSON structure of your package
type PackageFile struct {
	Name         string            `json:"name"`
	Private      bool              `json:"private"`
	Version      string            `json:"version"`
	Path         string            `json:"path"`
	MainDir      string            `json:"mainDir"` // <-- added field for main source folder
	Dependencies map[string]string `json:"dependencies"`
}

var root = "."
var ktFiles []string

func Start(scan bool) {
	if scan {
		err, ktFiles := KtFinder(root)
		if err != nil {
			fmt.Println("Error walking directory:", err)
			return
		}
		if len(ktFiles) == 0 {
			fmt.Println("No Kotlin (.kt) files found.")
			return
		}
		for _, filePath := range ktFiles {
			imports := ReadImports(filePath)
			for _, imp := range imports {
				Pkg.Dependencies[imp] = "0"
			}
		}
	}
	// Don't write JSON here
}
