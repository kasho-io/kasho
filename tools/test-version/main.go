package main

import (
	"fmt"
	"kasho/pkg/version"
)

func main() {
	info := version.Info()
	fmt.Printf("Kasho Version Information:\n")
	fmt.Printf("  Version: %s\n", info.Version)
	fmt.Printf("  Major: %d\n", info.Major) 
	fmt.Printf("  Minor: %d\n", info.Minor)
	fmt.Printf("  Patch: %d\n", info.Patch)
	fmt.Printf("  Git Commit: %s\n", info.GitCommit)
	fmt.Printf("  Build Date: %s\n", info.BuildDate)
}