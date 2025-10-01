package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate documentation for all commands",
	Long:  `Generate Markdown documentation for all commands in the CLI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		outputDir, _ := cmd.Flags().GetString("output")
		// Ensure the output directory exists
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return err
		}
		// Generate the documentation
		if err := doc.GenMarkdownTree(rootCmd, outputDir); err != nil {
			return err
		}
		// Rename the top-level index to README.md if it exists
		rootFile := filepath.Join(outputDir, rootCmd.Use+".md")
		readmeFile := filepath.Join(outputDir, "README.md")
		if _, err := os.Stat(rootFile); err == nil {
			return os.Rename(rootFile, readmeFile)
		}
		return nil
	},
}

func init() {
	docsCmd.Flags().StringP("output", "o", "./docs", "Output directory for generated documentation")
	rootCmd.AddCommand(docsCmd)
}
