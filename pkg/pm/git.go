package pm

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitClient provides git operations for the PM system.
// Implementations can use different git backends or mock implementations for testing.
type GitClient interface {
	// CreateBranch creates a new git branch for a work item.
	CreateBranch(branchName string) error

	// BranchExists checks if a branch already exists.
	BranchExists(branchName string) bool

	// GetCurrentBranch returns the current branch name.
	GetCurrentBranch() (string, error)

	// GetGitUserName returns the git user name from config.
	GetGitUserName() (string, error)
}

// OSGitClient implements GitClient using OS exec commands.
// It executes git commands directly on the system.
type OSGitClient struct{}

// NewOSGitClient creates a new OS git client.
// Requires git to be installed and available in PATH.
func NewOSGitClient() *OSGitClient {
	return &OSGitClient{}
}

// CreateBranch creates a new git branch.
// It switches to the new branch after creation.
func (gc *OSGitClient) CreateBranch(branchName string) error {
	cmd := exec.Command("git", "checkout", "-b", branchName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create branch %s: %s", branchName, string(output))
	}
	return nil
}

// BranchExists checks if a branch exists.
// Returns true if the branch exists locally.
func (gc *OSGitClient) BranchExists(branchName string) bool {
	cmd := exec.Command("git", "branch", "--list", branchName)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) != ""
}

// GetCurrentBranch returns the current branch name.
// Returns an error if not in a git repository or command fails.
func (gc *OSGitClient) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetGitUserName returns the git user name from config.
// Returns an error if git config is not set or command fails.
func (gc *OSGitClient) GetGitUserName() (string, error) {
	cmd := exec.Command("git", "config", "user.name")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git user name: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// BranchNamer generates branch names for work items.
// It creates standardized branch names based on item type and name.
type BranchNamer struct{}

// NewBranchNamer creates a new branch namer.
// No configuration needed for standard branch naming.
func NewBranchNamer() *BranchNamer {
	return &BranchNamer{}
}

// GenerateBranchName creates a branch name for a work item.
// Format: "{itemType}/{name}" (e.g., "feature/user-auth").
func (bn *BranchNamer) GenerateBranchName(itemType ItemType, name string) string {
	return fmt.Sprintf("%s/%s", itemType, name)
}

// GitIntegration handles git operations for work items.
// It manages branch creation and git-related workflow operations.
type GitIntegration struct {
	client GitClient
	namer  *BranchNamer
}

// NewGitIntegration creates a new git integration instance.
// Requires a GitClient implementation for git operations.
func NewGitIntegration(client GitClient) *GitIntegration {
	return &GitIntegration{
		client: client,
		namer:  NewBranchNamer(),
	}
}

// CreateWorkItemBranch creates a git branch for a new work item.
// Branch name format: "{itemType}/{name}". Does not fail if branch already exists.
func (gi *GitIntegration) CreateWorkItemBranch(itemType ItemType, name string) error {
	branchName := gi.namer.GenerateBranchName(itemType, name)

	if gi.client.BranchExists(branchName) {
		// Branch already exists, don't error
		return nil
	}

	if err := gi.client.CreateBranch(branchName); err != nil {
		// Log warning but don't fail the work item creation
		fmt.Printf("Note: Could not create git branch %s (%v)\n", branchName, err)
		return nil // Don't return error to avoid breaking work item creation
	}

	return nil
}

// CreateWorkItemBranchForPhase creates a git branch for a work item phase.
// Branch name format: "{itemType}/{name}/{phase}". Does not fail if branch already exists.
func (gi *GitIntegration) CreateWorkItemBranchForPhase(itemType ItemType, name string, phase WorkPhase) error {
	branchName := fmt.Sprintf("%s/%s/%s", itemType, name, phase)

	if gi.client.BranchExists(branchName) {
		// Branch already exists, don't error
		return nil
	}

	if err := gi.client.CreateBranch(branchName); err != nil {
		// Log warning but don't fail the phase advancement
		fmt.Printf("Note: Could not create git branch %s (%v)\n", branchName, err)
		return nil // Don't return error to avoid breaking phase advancement
	}

	return nil
}

// NoOpGitClient is a git client that does nothing (for testing or when git is not available).
// All operations succeed without doing anything.
type NoOpGitClient struct{}

// NewNoOpGitClient creates a no-op git client.
// Useful for testing or environments where git is not available.
func NewNoOpGitClient() *NoOpGitClient {
	return &NoOpGitClient{}
}

func (gc *NoOpGitClient) CreateBranch(branchName string) error {
	return nil
}

func (gc *NoOpGitClient) BranchExists(branchName string) bool {
	return false
}

func (gc *NoOpGitClient) GetCurrentBranch() (string, error) {
	return "main", nil
}

func (gc *NoOpGitClient) GetGitUserName() (string, error) {
	return "test-user", nil
}
