package main

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// version is set during build time via -ldflags
var version = "dev"

// gitSHA is set during build time via -ldflags
var gitSHA = "unknown"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print detailed version information including build details",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("go-pm version %s\n", version)
		fmt.Printf("Git SHA: %s\n", gitSHA)
		fmt.Printf("Go version: %s\n", runtime.Version())
		fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}
