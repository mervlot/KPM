package initialization

import (
	"encoding/json"
	"fmt"
	"kpm/libscanner"
	"kpm/readline"
	"kpm/scanner"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Main() {
	OldPkg := scanner.PackageFile{
		Dependencies: make(map[string]string),
	}
	dirPath, err := os.Getwd()
	if err != nil {
		fmt.Println("Something wrong happened when trying to get the path:", err)
		return
	}
	libscanner.Pkg.Path = dirPath

	// Check for package.kpm once
	_, statErr := os.Stat("package.kpm")
	if statErr == nil {
		fmt.Println("A package.kpm file is present. Should we overwrite it? (y,N)")
		res, _ := readline.Main()
		if strings.ToLower(res) != "y" {
			scanner.ReadFile("package.kpm", OldPkg)
			return
		}
	} else if !os.IsNotExist(statErr) {
		fmt.Println("Something went wrong when checking package.kpm:", statErr)
		return
	}

	// Ask whether to scan dependencies
	fmt.Printf("Would you like us to scan all the dependencies you are currently using? (y,N): ")
	res, _ := readline.Main()
	if strings.ToLower(res) == "y" {
		libscanner.Start(true)
	} else {
		fmt.Println("Okay, we won't extract dependency info from your files, but make sure you run `kpm extract` later.")
		libscanner.Start(false)
	}

	// Project name
	fmt.Printf("Project name (default is folder name): ")
	proj_input, err := readline.Main()
	if err != nil {
		log.Fatal("error reading input")
	}
	if proj_input == "" {
		proj_input = filepath.Base(dirPath)
	}
	libscanner.Pkg.Name = proj_input

	// Main directory
	fmt.Printf("Main source folder (default './src'): ")
	mainDir, err := readline.Main()
	if err != nil {
		log.Fatal("error reading input")
	}
	if mainDir == "" {
		mainDir = "./src"
	}
	libscanner.Pkg.MainDir = mainDir

	// Private
	fmt.Printf("Private? (Y, n): ")
	private_bool := false
	for {
		private_val, err := readline.Main()
		if err != nil {
			fmt.Printf("Something went wrong, please try again: ")
			continue
		}
		if strings.ToLower(private_val) == "y" {
			private_bool = true
			break
		} else if strings.ToLower(private_val) == "n" || private_val == "" {
			break
		} else {
			fmt.Printf("Invalid response, please try again: ")
		}
	}
	libscanner.Pkg.Private = private_bool

	// Version
	version := "0.0.0"
	for {
		fmt.Printf("Enter a version (default %v, format x.y.z): ", version)
		versionInput, err := readline.Main()
		if err != nil {
			fmt.Println("Something went wrong, please try again")
			continue
		}
		if versionInput != "" {
			version = versionInput
		}
		break
	}
	libscanner.Pkg.Version = version

	fmt.Printf("\nProject initialized:\nName: %s\nMain Dir: %s\nPrivate: %v\nVersion: %s\n",
		libscanner.Pkg.Name, libscanner.Pkg.MainDir, libscanner.Pkg.Private, libscanner.Pkg.Version)
	data, err := json.MarshalIndent(libscanner.Pkg, "", "  ")
	if err != nil {
		log.Fatal("Failed to marshal package:", err)
	}
	err = libscanner.WriteToJson(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("package.kpm saved successfully!")

}
