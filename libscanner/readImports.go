package libscanner

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// readImports reads import lines from a Kotlin file
func ReadImports(filePath string) []string {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening", filePath+":", err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	imports := []string{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "package") {
			continue
		}

		if after, ok := strings.CutPrefix(line, "import "); ok {
			if !strings.HasPrefix(line, "import kotlin.") {
				println("dd",line)
				importPath := strings.TrimSpace(after)
				imports = append(imports, importPath)
				continue
			} else {
				println(line)
				continue
			}

		}

		break // Stop once imports section ends
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Scan error in", filePath+":", err)
	}

	return imports
}
