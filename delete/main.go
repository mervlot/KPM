package delete

import (
	"log"
	"os"
)

func DeleteFileOrFolder(name string) error {
	err := os.Remove(name)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}
