package pm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkItemParser(t *testing.T) {
	fs := NewMockFileSystem()
	parser := NewWorkItemParser(fs)

	// Create a mock README file with NEW status/phase values
	content := `# Feature: user-auth

## Status: PLANNING
## Phase: planning
## Progress: 25%
## Assigned To: agent

## Overview
User authentication feature

---

## Planning Phase

### Tasks
- [x] Analyze requirements
- [ ] Interview stakeholders
`

	fs.WriteFile("/tmp/test.md", []byte(content)) //nolint:errcheck

	item, err := parser.ParseWorkItem("feature-user-auth", "/tmp/test.md")
	require.NoError(t, err)

	assert.Equal(t, "feature-user-auth", item.Name)
	assert.Equal(t, "user-auth", item.Title)
	assert.Equal(t, StatusPlanning, item.Status)
	assert.Equal(t, PhasePlanning, item.Phase)
	assert.Equal(t, 25, item.Progress)
	assert.Equal(t, "agent", item.AssignedTo)
	assert.Equal(t, TypeFeature, item.Type)
	assert.Len(t, item.Tasks, 2)
	assert.True(t, item.Tasks[0].Completed)
	assert.Equal(t, "Analyze requirements", item.Tasks[0].Description)
	assert.False(t, item.Tasks[1].Completed)
	assert.Equal(t, "Interview stakeholders", item.Tasks[1].Description)
}

func TestStatusUpdater(t *testing.T) {
	fs := NewMockFileSystem()
	updater := NewStatusUpdater(fs)

	content := `# Feature: test

## Status: PLANNING
## Phase: planning
## Progress: 0%
`

	fs.WriteFile("/tmp/test.md", []byte(content)) //nolint:errcheck

	err := updater.UpdateStatus("/tmp/test.md", StatusImplementation)
	require.NoError(t, err)

	updated, err := fs.ReadFile("/tmp/test.md")
	require.NoError(t, err)
	assert.Contains(t, string(updated), "## Status: IMPLEMENTATION")
}

func TestPhaseUpdater(t *testing.T) {
	fs := NewMockFileSystem()
	updater := NewStatusUpdater(fs)

	content := `# Feature: test

## Status: PROPOSED
## Phase: discovery
## Progress: 0%
`

	fs.WriteFile("/tmp/test.md", []byte(content)) //nolint:errcheck

	err := updater.UpdatePhase("/tmp/test.md", PhasePlanning)
	require.NoError(t, err)

	updated, err := fs.ReadFile("/tmp/test.md")
	require.NoError(t, err)
	assert.Contains(t, string(updated), "## Phase: planning")
}

func TestProgressUpdater(t *testing.T) {
	fs := NewMockFileSystem()
	updater := NewStatusUpdater(fs)

	content := `# Feature: test

## Status: PROPOSED
## Phase: discovery
## Progress: 0%
`

	fs.WriteFile("/tmp/test.md", []byte(content)) //nolint:errcheck

	err := updater.UpdateProgress("/tmp/test.md", 75)
	require.NoError(t, err)

	updated, err := fs.ReadFile("/tmp/test.md")
	require.NoError(t, err)
	assert.Contains(t, string(updated), "## Progress: 75%")
}

func TestAssigneeUpdater(t *testing.T) {
	fs := NewMockFileSystem()
	updater := NewStatusUpdater(fs)

	content := `# Feature: test

## Status: PROPOSED
## Phase: discovery
## Progress: 0%
## Assigned To: agent
`

	fs.WriteFile("/tmp/test.md", []byte(content)) //nolint:errcheck

	err := updater.UpdateAssignee("/tmp/test.md", "john.doe")
	require.NoError(t, err)

	updated, err := fs.ReadFile("/tmp/test.md")
	require.NoError(t, err)
	assert.Contains(t, string(updated), "## Assigned To: john.doe")
}

func TestTaskCompletion(t *testing.T) {
	fs := NewMockFileSystem()
	updater := NewStatusUpdater(fs)

	content := `# Feature: test

## Discovery Phase

### Tasks
- [ ] Task 1
- [ ] Task 2
- [ ] Task 3
`

	fs.WriteFile("/tmp/test.md", []byte(content)) //nolint:errcheck

	err := updater.CompleteTask("/tmp/test.md", 1) // Complete second task (0-indexed)
	require.NoError(t, err)

	updated, err := fs.ReadFile("/tmp/test.md")
	require.NoError(t, err)
	assert.Contains(t, string(updated), "- [x] Task 2")
	assert.Contains(t, string(updated), "- [ ] Task 1")
	assert.Contains(t, string(updated), "- [ ] Task 3")
}

func TestTaskParser(t *testing.T) {
	fs := NewMockFileSystem()
	parser := NewTaskParser(fs)

	content := `# Feature: test

## Discovery Phase

### Tasks
- [x] Task 1
- [ ] Task 2
- [x] Task 3
`

	fs.WriteFile("/tmp/test.md", []byte(content)) //nolint:errcheck

	total, completed, err := parser.ParseTaskList("/tmp/test.md")
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Equal(t, 2, completed)
}

func TestPostmortemGenerator(t *testing.T) {
	fs := NewMockFileSystem()
	gen := NewPostmortemGenerator(fs)

	err := gen.GeneratePostmortem("/tmp/completed/feature-test", "feature-test")
	require.NoError(t, err)

	content, err := fs.ReadFile("/tmp/completed/feature-test/POSTMORTEM.md")
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Postmortem: feature-test")
	assert.Contains(t, string(content), "## What Went Well")
	assert.Contains(t, string(content), "## What Could Be Improved")
}

func TestGetNextPhase(t *testing.T) {
	fs := NewMockFileSystem()
	config := Config{BacklogDir: "/tmp/backlog"}
	service := NewWorkItemService(config, fs, nil)

	tests := []struct {
		name           string
		currentStatus  ItemStatus
		currentPhase   WorkPhase
		expectedPhase  WorkPhase
		expectedStatus ItemStatus
		expectError    bool
	}{
		{
			name:           "PLANNING to IMPLEMENTATION",
			currentStatus:  StatusPlanning,
			currentPhase:   PhasePlanning,
			expectedPhase:  PhaseImplementation,
			expectedStatus: StatusImplementation,
			expectError:    false,
		},
		{
			name:           "IMPLEMENTATION to REVIEW",
			currentStatus:  StatusImplementation,
			currentPhase:   PhaseImplementation,
			expectedPhase:  PhaseReview,
			expectedStatus: StatusReview,
			expectError:    false,
		},
		{
			name:           "REVIEW to COMPLETED",
			currentStatus:  StatusReview,
			currentPhase:   PhaseReview,
			expectedPhase:  PhaseReview,
			expectedStatus: StatusCompleted,
			expectError:    false,
		},
		{
			name:           "COMPLETED cannot advance",
			currentStatus:  StatusCompleted,
			currentPhase:   PhaseReview,
			expectedPhase:  "",
			expectedStatus: "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phase, status, err := service.getNextPhase(tt.currentPhase, tt.currentStatus)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedPhase, phase)
				assert.Equal(t, tt.expectedStatus, status)
			}
		})
	}
}

func TestValidateStatus(t *testing.T) {
	fs := NewMockFileSystem()
	config := Config{BacklogDir: "/tmp/backlog"}
	service := NewWorkItemService(config, fs, nil)

	tests := []struct {
		name        string
		status      ItemStatus
		expectError bool
	}{
		{"PLANNING is valid", StatusPlanning, false},
		{"IMPLEMENTATION is valid", StatusImplementation, false},
		{"REVIEW is valid", StatusReview, false},
		{"COMPLETED is valid", StatusCompleted, false},
		{"Invalid status", ItemStatus("INVALID"), true},
		{"Legacy status should fail", ItemStatus("IN_PROGRESS_DISCOVERY"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateStatus(tt.status)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePhase(t *testing.T) {
	fs := NewMockFileSystem()
	config := Config{BacklogDir: "/tmp/backlog"}
	service := NewWorkItemService(config, fs, nil)

	tests := []struct {
		name        string
		phase       WorkPhase
		expectError bool
	}{
		{"PLANNING is valid", PhasePlanning, false},
		{"IMPLEMENTATION is valid", PhaseImplementation, false},
		{"REVIEW is valid", PhaseReview, false},
		{"Invalid phase", WorkPhase("invalid"), true},
		{"Legacy phase should fail", WorkPhase("discovery"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validatePhase(tt.phase)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckpoint(t *testing.T) {
	fs := NewMockFileSystem()
	config := Config{BacklogDir: "/tmp/backlog", EnableGit: true}
	gitClient := NewNoOpGitClient()
	service := NewWorkItemService(config, fs, gitClient)

	// Create a work item first
	require.NoError(t, fs.CreateDirectory(config.BacklogDir))

	// Create a simple work item README
	content := `# Feature: test-feature

## Status: PLANNING
## Phase: planning
## Progress: 15%
## Assigned To: agent

## Planning Phase
### Tasks
- [x] Task 1
- [ ] Task 2
`
	workItemPath := "/tmp/backlog/feature-test-feature"
	require.NoError(t, fs.CreateDirectory(workItemPath))
	require.NoError(t, fs.WriteFile(workItemPath+"/README.md", []byte(content)))

	// Test checkpoint without advancing phase
	err := service.Checkpoint(context.Background(), "feature-test-feature", "Completed design document")
	require.NoError(t, err)

	// Verify work item still in same phase/status
	item, err := service.parser.ParseWorkItem("feature-test-feature", workItemPath+"/README.md")
	require.NoError(t, err)
	assert.Equal(t, StatusPlanning, item.Status)
	assert.Equal(t, PhasePlanning, item.Phase)
}

func TestCheckpointGitEnabled(t *testing.T) {
	fs := NewMockFileSystem()
	config := Config{BacklogDir: "/tmp/backlog", EnableGit: true}
	gitClient := NewNoOpGitClient()
	service := NewWorkItemService(config, fs, gitClient)

	// Create a work item
	require.NoError(t, fs.CreateDirectory(config.BacklogDir))
	workItemPath := "/tmp/backlog/feature-test-feature"
	require.NoError(t, fs.CreateDirectory(workItemPath))

	content := `# Feature: test-feature

## Status: IMPLEMENTATION
## Phase: implementation
## Progress: 50%
## Assigned To: agent
`
	require.NoError(t, fs.WriteFile(workItemPath+"/README.md", []byte(content)))

	// Checkpoint should create git commit when enabled
	err := service.Checkpoint(context.Background(), "feature-test-feature", "Completed unit tests")
	require.NoError(t, err)

	// Verify git commit was attempted (mock should record this)
	// Note: Actual verification depends on mock implementation
}

func TestCheckpointGitDisabled(t *testing.T) {
	fs := NewMockFileSystem()
	config := Config{BacklogDir: "/tmp/backlog", EnableGit: false}
	gitClient := NewNoOpGitClient()
	service := NewWorkItemService(config, fs, gitClient)

	// Create a work item
	require.NoError(t, fs.CreateDirectory(config.BacklogDir))
	workItemPath := "/tmp/backlog/feature-test-feature"
	require.NoError(t, fs.CreateDirectory(workItemPath))

	content := `# Feature: test-feature

## Status: IMPLEMENTATION
## Phase: implementation
## Progress: 50%
## Assigned To: agent
`
	require.NoError(t, fs.WriteFile(workItemPath+"/README.md", []byte(content)))

	// Checkpoint should succeed without git operations
	err := service.Checkpoint(context.Background(), "feature-test-feature", "Mid-phase checkpoint")
	require.NoError(t, err)

	// Should not have called git (mock could verify no git calls made)
}

func TestRequestReview(t *testing.T) {
	fs := NewMockFileSystem()
	config := Config{BacklogDir: "/tmp/backlog"}
	service := NewWorkItemService(config, fs, NewNoOpGitClient())

	// Create a work item in IMPLEMENTATION phase
	require.NoError(t, fs.CreateDirectory(config.BacklogDir))
	workItemPath := "/tmp/backlog/feature-test-feature"
	require.NoError(t, fs.CreateDirectory(workItemPath))

	content := `# Feature: test-feature

## Status: IMPLEMENTATION
## Phase: implementation
## Progress: 75%
## Assigned To: agent

## Implementation Phase
### Tasks
- [x] Implement core functionality
- [x] Write unit tests
- [ ] Code review
`
	require.NoError(t, fs.WriteFile(workItemPath+"/README.md", []byte(content)))

	// Request review should succeed from IMPLEMENTATION phase
	err := service.RequestReview(context.Background(), "feature-test-feature")
	require.NoError(t, err)

	// Verify work item advanced to REVIEW phase
	item, err := service.parser.ParseWorkItem("feature-test-feature", workItemPath+"/README.md")
	require.NoError(t, err)
	assert.Equal(t, StatusReview, item.Status)
	assert.Equal(t, PhaseReview, item.Phase)
}

func TestRequestReviewFromWrongPhase(t *testing.T) {
	fs := NewMockFileSystem()
	config := Config{BacklogDir: "/tmp/backlog"}
	service := NewWorkItemService(config, fs, NewNoOpGitClient())

	// Create a work item in PLANNING phase
	require.NoError(t, fs.CreateDirectory(config.BacklogDir))
	workItemPath := "/tmp/backlog/feature-test-feature"
	require.NoError(t, fs.CreateDirectory(workItemPath))

	content := `# Feature: test-feature

## Status: PLANNING
## Phase: planning
## Progress: 25%
## Assigned To: agent
`
	require.NoError(t, fs.WriteFile(workItemPath+"/README.md", []byte(content)))

	// Request review should fail from PLANNING phase
	err := service.RequestReview(context.Background(), "feature-test-feature")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in IMPLEMENTATION phase")
}

func TestRequestReviewNonExistentWorkItem(t *testing.T) {
	fs := NewMockFileSystem()
	config := Config{BacklogDir: "/tmp/backlog"}
	service := NewWorkItemService(config, fs, NewNoOpGitClient())

	require.NoError(t, fs.CreateDirectory(config.BacklogDir))

	// Request review should fail for non-existent work item
	err := service.RequestReview(context.Background(), "feature-nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "work item not found")
}

func TestApproveReview(t *testing.T) {
	fs := NewMockFileSystem()
	config := Config{BacklogDir: "/tmp/backlog"}
	service := NewWorkItemService(config, fs, NewNoOpGitClient())

	// Create a work item in REVIEW phase
	require.NoError(t, fs.CreateDirectory(config.BacklogDir))
	workItemPath := "/tmp/backlog/feature-test-feature"
	require.NoError(t, fs.CreateDirectory(workItemPath))

	content := `# Feature: test-feature

## Status: REVIEW
## Phase: review
## Progress: 95%
## Assigned To: agent

## Review Phase
### Tasks
- [x] Final testing
- [x] Documentation review
- [ ] Human approval
`
	require.NoError(t, fs.WriteFile(workItemPath+"/README.md", []byte(content)))

	// Approve review should advance to COMPLETED
	err := service.ApproveReview(context.Background(), "feature-test-feature")
	require.NoError(t, err)

	// Verify work item advanced to COMPLETED status
	item, err := service.parser.ParseWorkItem("feature-test-feature", workItemPath+"/README.md")
	require.NoError(t, err)
	assert.Equal(t, StatusCompleted, item.Status)
	// Phase should still be review (COMPLETED is a status, not a phase)
	assert.Equal(t, PhaseReview, item.Phase)
}

func TestApproveReviewFromWrongPhase(t *testing.T) {
	fs := NewMockFileSystem()
	config := Config{BacklogDir: "/tmp/backlog"}
	service := NewWorkItemService(config, fs, NewNoOpGitClient())

	// Create a work item in IMPLEMENTATION phase
	require.NoError(t, fs.CreateDirectory(config.BacklogDir))
	workItemPath := "/tmp/backlog/feature-test-feature"
	require.NoError(t, fs.CreateDirectory(workItemPath))

	content := `# Feature: test-feature

## Status: IMPLEMENTATION
## Phase: implementation
## Progress: 75%
## Assigned To: agent
`
	require.NoError(t, fs.WriteFile(workItemPath+"/README.md", []byte(content)))

	// Approve review should fail from IMPLEMENTATION phase
	err := service.ApproveReview(context.Background(), "feature-test-feature")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in REVIEW phase")
}

func TestApproveReviewNonExistentWorkItem(t *testing.T) {
	fs := NewMockFileSystem()
	config := Config{BacklogDir: "/tmp/backlog"}
	service := NewWorkItemService(config, fs, NewNoOpGitClient())

	require.NoError(t, fs.CreateDirectory(config.BacklogDir))

	// Approve review should fail for non-existent work item
	err := service.ApproveReview(context.Background(), "feature-nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "work item not found")
}
