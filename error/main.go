package error

import (
	"fmt"
)
type MyError struct {
    Code    int
    Message string
}
func (e *MyError) Error() string {
    return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}
func doSomething() error {
    return &MyError{Code: 404, Message: "Resource not found"}
}
func main() {
    err := doSomething()
    if err != nil {
        fmt.Println(err) // Output: Error 404: Resource not found
    }
}