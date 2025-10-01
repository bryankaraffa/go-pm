package pm

import (
	_ "embed"
)

//go:embed templates/instructions.md
var goInstructions string

// GetInstructions returns the project management instructions.
// The instructions provide guidance for documentation-driven development
// and collaboration between humans and agents.
//
// The instructions cover:
//   - When to use the PM tool
//   - Phase-specific responsibilities
//   - Best practices for work item management
//   - Collaboration guidelines
func GetInstructions(config Config) string {
	// Return embedded instructions
	return goInstructions
}
