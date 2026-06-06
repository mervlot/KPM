package env

import (
	"os"
	"runtime"
)
type Environment struct {
	Windows string 
	linux   string
	darwin  string
}
var Env = Environment{
	Windows: "",
	linux:   "linux",
	darwin:  "macos",
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
home, _ := os.UserHomeDir()
fmt.Println("User Home Directory:", home)
}