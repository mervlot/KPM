package scanner

import (
	"encoding/json"
	"os"
)

type PackageFile struct {
	Name         string            `json:"name"`
	Private      bool              `json:"private"`
	Version      string            `json:"version"`
	Path         string            `json:"path"`
	Maindir      string            `json:"mainDir"`
	Dependencies map[string]string `json:"dependencies"`
	Scripts      map[string]string `json:"scripts"`
}

var Pkg = PackageFile{
	Dependencies: make(map[string]string),
}

func ReadFile(file string, Pkg PackageFile) {

	data, err := os.ReadFile(file)
	if err != nil {
		return
	}
	json.Unmarshal(data, &Pkg)

}
