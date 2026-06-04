package types

import (
	"encoding/json"
	"fmt"
	"kpm/libscanner"
	"os"
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
