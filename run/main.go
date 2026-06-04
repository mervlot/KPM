package run

import (
	"context"
	"encoding/json"
	"fmt"
	"kpm/scanner"
	"os"
	"os/exec"
	"strings"
)

func Main(cmd string) {
	var pkgfile scanner.PackageFile
	pkgbyte, err := os.ReadFile("package.kpm")
	if err != nil {
		return
	}
	json.Unmarshal(pkgbyte, &pkgfile)
	if len(pkgfile.Scripts) == 0 {
		fmt.Println("no script is defined")
	}

	if ncmd, exist := pkgfile.Scripts[cmd]; exist {
		res := strings.Split(ncmd, " ")

		out, err := exec.CommandContext(context.Background(),res[0], res[1:]...).Output()
		if err != nil {
			fmt.Println(err)
			return
		} else {
			fmt.Println(string(out))
		}
	} else {
		fmt.Println("command  not found")
	}

	if err != nil {
		fmt.Println(err)
		return
	}
	// fmt.Println(string(out))
}
