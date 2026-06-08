package types

import (
	"encoding/json"
	"fmt"
	"kpm/libscanner"
	"os"
	"path/filepath"
)

func GetRoot() (string, error) {
	var kpm libscanner.PackageFile
	file, err := os.ReadFile("package.kpm")
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	json.Unmarshal(file, &kpm)
	return kpm.MainDir, nil
}
func fmtRelativePath(group, artifact, version, name, ver, ext string) string {
	return filepath.ToSlash(filepath.Join(".", "libs", group, artifact, version, fmt.Sprintf("%s-%s.%s", name, ver, ext)))
}