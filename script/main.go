package script

import (
	"fmt"
	"os/exec"

	install "kpm/package"
)

func Main(key string) {
	pkg, err := install.ReadPackageFile("package.kpm")
	if err != nil {
		fmt.Println("error reading package.kpm", err)
		return
	}
	scripts := pkg.Dependencies
	if scripts[key] == ""{
		fmt.Println("i think this command does not exist in your script")
		return
	}
	cmd := exec.Command(scripts[key])
	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}
