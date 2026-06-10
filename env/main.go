package env

import (
	"fmt"
	"os"
	"runtime"
)
var home, _ = os.UserHomeDir()
type Environment struct {
	Windows string 
	Linux   string
	Darwin  string
}
var Env = Environment{
	Windows: fmt.Sprintf("%s\\AppData\\Local\\kpm", home),
	Linux:   fmt.Sprintf("%s/.kpm", home),
	Darwin:  fmt.Sprintf("%s/Library/Application Support/kpm", home),
}

// DetectOS returns a clean string identifying the target environment
func DetectOS() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "linux":
		return "linux"
	case "darwin":
		return "macos"
	default:
		return "unknown"
	}
}
func main() {
	println(Env.Windows)
	println(Env.Linux)
	println(Env.Darwin)
}