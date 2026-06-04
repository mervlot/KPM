package clean

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"kpm/types"
)

func Main(args []string) {
	// Setup flag set
	fs := flag.NewFlagSet("clean", flag.ExitOnError)
	resourcesPath := fs.String("resources", "resource.kpm", "path to resource.kpm file")
	libsPath := fs.String("libs", "libs", "path to libs directory")
	dryRun := fs.Bool("dry-run", true, "don't actually delete, just show what would be removed")
	yes := fs.Bool("yes", false, "don't prompt, assume yes")
	verbose := fs.Bool("v", false, "verbose output")

	if err := fs.Parse(args); err != nil {
		log.Fatalf("failed parsing flags: %v", err)
	}

	// Read and parse resources file
	data, err := os.ReadFile(*resourcesPath)
	if err != nil {
		log.Fatalf("failed reading resources file %s: %v", *resourcesPath, err)
	}

	var rf types.ResourceFile
	if err := json.Unmarshal(data, &rf); err != nil {
		log.Fatalf("failed parsing resources file: %v", err)
	}

	used := make(map[string]struct{})
	for _, r := range rf.Resources {
		group := ""
		if r.Group != nil {
			group = *r.Group
		} else {
			parts := strings.Split(r.LPath, "/")
			if len(parts) >= 1 {
				group = parts[0]
			}
		}

		name := r.Name
		version := ""
		if r.Version != nil {
			version = *r.Version
		} else {
			parts := strings.Split(r.LPath, "/")
			if len(parts) >= 3 {
				version = parts[2]
			}
		}

		if group == "" || name == "" || version == "" {
			if *verbose {
				log.Printf("skipping resource with incomplete identity: %+v", r)
			}
			continue
		}
		key := filepath.Join(group, name, version)
		used[key] = struct{}{}
	}

	var candidates []string
	err = filepath.Walk(*libsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(*libsPath, path)
		if err != nil {
			return nil
		}
		parts := strings.Split(rel, string(os.PathSeparator))
		if len(parts) == 3 {
			key := filepath.Join(parts[0], parts[1], parts[2])
			if _, ok := used[key]; !ok {
				candidates = append(candidates, path)
			}
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		log.Fatalf("failed scanning libs directory %s: %v", *libsPath, err)
	}

	if len(candidates) == 0 {
		fmt.Println("no unused version directories found")
		return
	}

	fmt.Println("Unused version directories:")
	for _, c := range candidates {
		fmt.Println(" -", c)
	}

	if *dryRun {
		fmt.Println("dry-run: no files will be removed. Pass --dry-run=false and --yes to actually delete.")
		return
	}

	if !*yes {
		fmt.Print("Proceed to delete these directories? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		ans, _ := reader.ReadString('\n')
		ans = strings.TrimSpace(strings.ToLower(ans))
		if ans != "y" && ans != "yes" {
			fmt.Println("aborted")
			return
		}
	}

	for _, c := range candidates {
		if *verbose {
			fmt.Println("removing", c)
		}
		if err := os.RemoveAll(c); err != nil {
			log.Printf("failed to remove %s: %v", c, err)
		}
	}
	fmt.Println("done")
}
