package install

import (
	"bufio"
	"encoding/json"
	"fmt"
	"kpm/libscanner"
	"kpm/scanner"
	"os"
	"strings"
)

// ReadPackageFile reads and parses the package.kpm file
func ReadPackageFile(file string) (*scanner.PackageFile, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var pkg scanner.PackageFile
	err = json.Unmarshal(data, &pkg)
	if err != nil {
		return nil, err
	}

	// Ensure dependencies map is initialized
	if pkg.Dependencies == nil {
		pkg.Dependencies = make(map[string]string)
	}

	return &pkg, nil
}

// SavePackageFile marshals the PackageFile to JSON and writes it to package.kpm
func SavePackageFile(pkg *scanner.PackageFile) error {
	pkgBytes, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling package file to JSON: %w", err)
	}

	err = libscanner.WriteToJson(pkgBytes, "package.kpm")
	if err != nil {
		return fmt.Errorf("error writing package file to package.kpm: %w", err)
	}

	return nil
}

// PromptVersionSelection prompts the user to select a version if one is already installed with a different version
func PromptVersionSelection(packageName, currentVersion, newVersion string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Printf("\n[VERSION CONFLICT] Package '%s' is already installed with version %s\n", packageName, currentVersion)
	fmt.Printf("You are trying to install version %s\n", newVersion)
	fmt.Printf("Which version would you like to use?\n")
	fmt.Printf("1) Keep current version: %s\n", currentVersion)
	fmt.Printf("2) Switch to new version: %s\n", newVersion)
	fmt.Printf("3) Keep both versions (install new version separately)\n")
	fmt.Print("Enter choice (1-3): ")
	
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	
	switch choice {
	case "1":
		return currentVersion, nil
	case "2":
		return newVersion, nil
	case "3":
		return newVersion, nil
	default:
		fmt.Println("Invalid choice, defaulting to new version")
		return newVersion, nil
	}
}

// UpdatePackageDependency updates or adds a dependency in package.kpm
func UpdatePackageDependency(packageName, version string) error {
	// Read existing package.kpm
	pkg, err := ReadPackageFile("package.kpm")
	if err != nil {
		// If file doesn't exist, create a new one
		if os.IsNotExist(err) {
			pkg = &scanner.PackageFile{
				Name:         "KPM",
				Private:      false,
				Version:      "0.0.1",
				Path:         "./",
				Maindir:      "./src",
				Dependencies: make(map[string]string),
				Scripts:      make(map[string]string),
			}
		} else {
			return fmt.Errorf("error reading package.kpm: %w", err)
		}
	}

	// Check if package already exists with different version
	if existingVersion, exists := pkg.Dependencies[packageName]; exists && existingVersion != version {
		selectedVersion, err := PromptVersionSelection(packageName, existingVersion, version)
		if err != nil {
			return err
		}
		pkg.Dependencies[packageName] = selectedVersion
	} else {
		// Add or update the dependency
		pkg.Dependencies[packageName] = version
	}

	// Save the updated package.kpm
	err = SavePackageFile(pkg)
	if err != nil {
		return fmt.Errorf("error saving package.kpm: %w", err)
	}

	fmt.Printf("Updated package.kpm: %s@%s\n", packageName, pkg.Dependencies[packageName])
	return nil
}

// UpdatePackageDependencyFromMaven updates package.kpm with Maven artifact info
func UpdatePackageDependencyFromMaven(group, artifact, version string) error {
	packageName := fmt.Sprintf("%s:%s", group, artifact)
	return UpdatePackageDependency(packageName, version)
}

// UpdatePackageDependencyFromUrl updates package.kpm with URL artifact info
func UpdatePackageDependencyFromUrl(artifact string) error {
	return UpdatePackageDependency(artifact, "url")
}
