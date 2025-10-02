package pm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitIntegration(t *testing.T) {
	client := NewNoOpGitClient()
	gi := NewGitIntegration(client)

	// Test branch creation
	err := gi.CreateWorkItemBranch(TypeFeature, "user-auth")
	assert.NoError(t, err)

	err = gi.CreateWorkItemBranchForPhase(TypeFeature, "user-auth", PhaseExecution)
	assert.NoError(t, err)
}

func TestBranchNamer(t *testing.T) {
	bn := NewBranchNamer()

	branchName := bn.GenerateBranchName(TypeFeature, "user-auth")
	assert.Equal(t, "feature/user-auth", branchName)

	branchName = bn.GenerateBranchName(TypeBug, "fix-crash")
	assert.Equal(t, "bug/fix-crash", branchName)
}
