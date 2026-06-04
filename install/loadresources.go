package install

import (
	"encoding/json"
	"kpm/types"
	"os"
)
func ReadFile(file string, Res *types.ResourceFile) {

	data, err := os.ReadFile(file)
	if err != nil {
		return
	}
	json.Unmarshal(data, &Res)

}
