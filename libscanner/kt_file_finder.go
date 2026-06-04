package libscanner

import (
	"io/fs"
	"path/filepath"
)

func KtFinder(root string) (error, []string) {
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".kt" {
			ktFiles = append(ktFiles, path)
		}
		return nil
	})

	if err != nil {
		return err, nil
	}
	return nil, ktFiles
}
