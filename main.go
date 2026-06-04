package main

import (
	"fmt"
	"kpm/clean"
	initialization "kpm/init"
	"kpm/install"
	kotlinlanguageserver "kpm/kotlin_language_server"
	"kpm/list"
	"kpm/run"
	"kpm/scanner"
	"kpm/search"
	"kpm/sync"
	"kpm/types"
	"os"
)

var version = "0.0.0"

func main() {
	args := os.Args
	root, root_err := types.GetRoot()

	if len(args) < 2 {
		printHelp()
		return
	}
	// ram.Main()
	switch args[1] {
	case "init", "-i", "--init":
		initialization.Main()
		kotlinlanguageserver.Main()
	case "run":
		run.Main(args[2])
	case "list", "-l", "--list":
		list.Main()

	case "clean", "--clean":
		clean.Main(os.Args[2:]) // pass args after "clean"

	case "help", "-h", "--help":
		printHelp()
	case "version", "-v", "--version":
		fmt.Println("kpm version:", version)
	case "scan", "extract", "-s", "-e":
		fmt.Println("Scanning project for dependencies...")
		if root_err != nil {
			fmt.Println("error encounderd on getting root dir ensure your package.kpm is clean if unsure re run \n `kpm init`")
		}

		scanner.Scanner(root)
		kotlinlanguageserver.Main()

	case "i", "install", "get", "-g":
		if len(args) > 2 {
			// Pass all arguments after the command (args[2:] is a slice)
			install.Main(false, args[2:]...)
		} else {
			install.Main(false)
		}
		kotlinlanguageserver.Main()

	case "update":
		if len(args) > 2 {
			// Pass all arguments after the command (args[2:] is a slice)
			install.Main(true, args[2:]...)
		} else {
			install.Main(true)
		}
		kotlinlanguageserver.Main()
	case "sync":
		sync.Main()
	case "search", "-f", "--find":
		search.Main()
	default:
		fmt.Printf("Unknown command: %s\n\n", args[1])
		printHelp()
	}
}

func printHelp() {
	fmt.Print(`
kpm CLI - available commands:

init, -i, --init       Initialize a new KPM project
scan, extract, -s, -e  Scan project files for dependencies
version, -v, --version Show the KPM version
help, -h, --help       Show this help message

Usage:
kpm <command> [options]
`)
}
