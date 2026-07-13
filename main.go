// This file must live at cmd/kpm/main.go, relative to the same directory
// that contains go.mod (the module root). That's a Go convention, not a
// KPM-specific rule: go.mod declares "module kpm", and every other file in
// this project imports its packages by that path (e.g. "kpm/internal/cli"),
// so go.mod's location IS what "kpm/..." resolves against. `go build
// ./cmd/kpm` (or `go run ./cmd/kpm`) must be run from that module root, or
// anywhere inside it — Go finds go.mod by walking up from your cwd.
//
// Why cmd/kpm/ specifically, instead of main.go sitting at the repo root
// like the old version: it's the standard Go layout for "this repo produces
// one or more executable binaries." cmd/<binary-name>/main.go is the
// convention so main.go's own directory name IS the binary name, and so the
// repo root stays free for exactly one thing per top-level folder
// (cmd/ for entrypoints, internal/ for everything else). If KPM ever grows
// a second binary — say a `kpmd` daemon for a remote build cache — it gets
// its own cmd/kpmd/main.go sitting right next to this one, with zero
// ambiguity about which main.go is which.
package main

import (
	"os"

	"kpm/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}