//go:build windows

package memory_bus

import "fmt"

// Dummy function to keep the compiler happy on Windows
func Dummy() {
	fmt.Println("Memory bus is not supported on Windows")
}
