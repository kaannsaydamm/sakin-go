// Package main is the legacy entry point for sakin-go
// This file is maintained for backward compatibility.
// New deployments should use cmd/sge-network-sensor/main.go
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/atailh4n/sakin/cmd/sge-network-sensor"
)

func main() {
	fmt.Println("==============================================")
	fmt.Println("  SGE Network Sensor - Legacy Entry Point")
	fmt.Println("==============================================")
	fmt.Println("")
	fmt.Println("WARNING: This is a legacy entry point.")
	fmt.Println("For new deployments, run the binary from:")
	fmt.Println("  cmd/sge-network-sensor/main.go")
	fmt.Println("")
	fmt.Println("Migrating to new entry point...")

	// Check if new binary exists
	execPath, err := os.Executable()
	if err != nil {
		log.Printf("Warning: Could not determine executable path: %v", err)
	} else {
		newBinary := filepath.Join(filepath.Dir(execPath), "sge-network-sensor")
		if _, err := os.Stat(newBinary); err == nil {
			fmt.Printf("New binary found at: %s\n", newBinary)
		}
	}

	fmt.Println("")
	fmt.Println("For usage information, run with --help flag")
	fmt.Println("")
	fmt.Println("Available commands:")
	fmt.Println("  --config <path>  : Configuration file path")
	fmt.Println("  --preset <name>  : Configuration preset (light|standard|aggressive)")
	fmt.Println("  --workers <n>    : Number of worker threads")
	fmt.Println("  --version        : Print version information")
	fmt.Println("")
	fmt.Println("Example:")
	fmt.Println("  ./sge-network-sensor --preset standard")
	fmt.Println("==============================================")
}
