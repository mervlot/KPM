package kotlinlanguageserver

import (
	"encoding/json"
	"fmt"
	"os"
)

func VsCode() {
	if err := os.Mkdir("./.vscode", 0755); err != nil {
		fmt.Println("failed to create vsc folder: %w", err)
	}

	// Step 3: VS Code settings for fwcd.kotlin
	settings := map[string]interface{}{
		"kotlin.languageServer.enabled": true,
		"kotlin.languageServer.path":    "", // empty means use default LSP
		"kotlin.classpath":              ".kls_classpath",
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling settings:", err)
		return
	}

	// Write settings.json
	err = os.WriteFile(".vscode/settings.json", jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing settings:", err)
		return
	}
}
