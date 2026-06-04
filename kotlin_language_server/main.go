package kotlinlanguageserver

import (
	"encoding/json"
	"kpm/types"
	"os"
)

func Main() error {
	var cp string
	var rkpm types.ResourceFile
	file, err := os.ReadFile("resource.kpm")
	if err != nil {
		return err
	}
	json.Unmarshal(file, &rkpm)
	for _, v := range rkpm.Resources {
		cp += v.LPath + ":"
	}
	os.WriteFile(".kls_classpath", []byte(cp), 0644)
	return nil
}
