package main

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// version is set during build time via -ldflags
var version = "dev"

// gitSHA is set during build time via -ldflags
var gitSHA = "unknown"

// buildDate is set during build time via -ldflags
var buildDate = "unknown"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print detailed version information including build details",
	Run: func(cmd *cobra.Command, args []string) {
		// Use ldflags values if they were set during build (goreleaser)
		// Otherwise fall back to build info (go install)
		displayVersion := version
		displayGitSHA := gitSHA
		displayBuildDate := buildDate

		// If ldflags weren't set, try to get info from build info
		if version == "dev" && gitSHA == "unknown" && buildDate == "unknown" {
			if buildInfo, ok := debug.ReadBuildInfo(); ok {
				// Use module version if available
				if buildInfo.Main.Version != "" && buildInfo.Main.Version != "(devel)" {
					displayVersion = buildInfo.Main.Version
				}

				// Extract git info from build settings
				for _, setting := range buildInfo.Settings {
					switch setting.Key {
					case "vcs.revision":
						if len(setting.Value) >= 12 {
							displayGitSHA = setting.Value[:12] // Short SHA
						} else {
							displayGitSHA = setting.Value
						}
					case "vcs.time":
						displayBuildDate = setting.Value
					}
				}
			}
		}

		fmt.Printf("go-pm version %s\n", displayVersion)
		fmt.Printf("Git SHA: %s\n", displayGitSHA)
		fmt.Printf("Build date: %s\n", displayBuildDate)
		fmt.Printf("Go version: %s\n", runtime.Version())
		fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}
