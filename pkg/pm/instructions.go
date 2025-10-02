package pm

import (
	_ "embed"
	"strings"
)

//go:embed templates/instructions.md
var goInstructions string

// GetInstructions returns the project management instructions.
// The instructions provide guidance for documentation-driven development
// and collaboration between humans and agents. The embedded template
// is processed with the provided config to replace directory placeholders
// with actual configured paths.
//
// The instructions cover:
//   - When to use the PM tool
//   - Phase-specific responsibilities
//   - Best practices for work item management
//   - Collaboration guidelines
func GetInstructions(config Config) string {
	// Process template with config values
	instructions := strings.ReplaceAll(goInstructions, "{{backlog_dir}}", config.BacklogDir)
	instructions = strings.ReplaceAll(instructions, "{{completed_dir}}", config.CompletedDir)
	return instructions
}
