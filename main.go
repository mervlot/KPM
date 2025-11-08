package main

import (
	"fmt"
	initialization "kpm/init"
	"kpm/install"
	"kpm/scanner"
	"os"
)

var version = "0.0.0"

func main() {
	args := os.Args

	if len(args) < 2 {
		printHelp()
		return
	}

	switch args[1] {
	case "init", "-i", "--init":
		initialization.Main()
	case "help", "-h", "--help":
		printHelp()
	case "version", "-v", "--version":
		fmt.Println("kpm version:", version)
	case "scan", "extract", "-s", "-e":
		fmt.Println("Scanning project for dependencies...")
		scanner.Scanner("./")
	case "i", "install", "get", "-g":
		install.Main()
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
