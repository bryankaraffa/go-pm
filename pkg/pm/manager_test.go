package pm

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultManager(t *testing.T) {
	config := DefaultConfig()
	manager := NewDefaultManager(config)

	require.NotNil(t, manager)
	assert.NotNil(t, manager.service)
}

func TestManagerFactory(t *testing.T) {
	config := DefaultConfig()
	manager := NewDefaultManager(config)
	require.NotNil(t, manager)
}

func TestManagerCreateWorkItem(t *testing.T) {
	config := DefaultConfig()
	fs := NewMockFileSystem()
	git := NewNoOpGitClient()
	manager := NewDefaultManagerWithDeps(config, fs, git)

	req := CreateRequest{
		Type: TypeFeature,
		Name: "test-feature",
	}

	item, err := manager.CreateWorkItem(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "feature-test-feature", item.Name)
	assert.Equal(t, TypeFeature, item.Type)
	assert.Equal(t, StatusPlanning, item.Status)
	assert.Equal(t, PhasePlanning, item.Phase)
}

func TestManagerListWorkItems(t *testing.T) {
	config := DefaultConfig()
	fs := NewMockFileSystem()
	git := NewNoOpGitClient()
	manager := NewDefaultManagerWithDeps(config, fs, git)

	// Create the backlog directory
	err := fs.CreateDirectory(config.BacklogDir)
	require.NoError(t, err)

	// Create a work item first
	req := CreateRequest{Type: TypeFeature, Name: "test-feature"}
	_, err = manager.CreateWorkItem(context.Background(), req)
	require.NoError(t, err)

	// List all items
	items, err := manager.ListWorkItems(context.Background(), ListFilter{})
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "feature-test-feature", items[0].Name)
}

func TestManagerGetWorkItem(t *testing.T) {
	config := DefaultConfig()
	fs := NewMockFileSystem()
	git := NewNoOpGitClient()
	manager := NewDefaultManagerWithDeps(config, fs, git)

	// Create the backlog directory
	err := fs.CreateDirectory(config.BacklogDir)
	require.NoError(t, err)

	// Create a work item first
	req := CreateRequest{Type: TypeFeature, Name: "test-feature"}
	_, err = manager.CreateWorkItem(context.Background(), req)
	require.NoError(t, err)

	// Get the item
	item, err := manager.GetWorkItem(context.Background(), "feature-test-feature")
	require.NoError(t, err)
	assert.Equal(t, "feature-test-feature", item.Name)
	assert.Equal(t, TypeFeature, item.Type)
}

func TestManagerUpdateStatus(t *testing.T) {
	config := DefaultConfig()
	fs := NewMockFileSystem()
	git := NewNoOpGitClient()
	manager := NewDefaultManagerWithDeps(config, fs, git)

	// Create the backlog directory
	err := fs.CreateDirectory(config.BacklogDir)
	require.NoError(t, err)

	// Create a work item first
	req := CreateRequest{Type: TypeFeature, Name: "test-feature"}
	_, err = manager.CreateWorkItem(context.Background(), req)
	require.NoError(t, err)

	// Update status
	err = manager.UpdateStatus(context.Background(), "feature-test-feature", StatusImplementation)
	require.NoError(t, err)

	// Verify status was updated
	item, err := manager.GetWorkItem(context.Background(), "feature-test-feature")
	require.NoError(t, err)
	assert.Equal(t, StatusImplementation, item.Status)
}

func TestManagerUpdateProgress(t *testing.T) {
	config := DefaultConfig()
	fs := NewMockFileSystem()
	git := NewNoOpGitClient()
	manager := NewDefaultManagerWithDeps(config, fs, git)

	// Create the backlog directory
	err := fs.CreateDirectory(config.BacklogDir)
	require.NoError(t, err)

	// Create a work item first
	req := CreateRequest{Type: TypeFeature, Name: "test-feature"}
	_, err = manager.CreateWorkItem(context.Background(), req)
	require.NoError(t, err)

	// Update progress
	err = manager.UpdateProgress(context.Background(), "feature-test-feature", 75)
	require.NoError(t, err)

	// Verify progress was updated
	item, err := manager.GetWorkItem(context.Background(), "feature-test-feature")
	require.NoError(t, err)
	assert.Equal(t, 75, item.Progress)
}

func TestManagerAssignWorkItem(t *testing.T) {
	config := DefaultConfig()
	fs := NewMockFileSystem()
	git := NewNoOpGitClient()
	manager := NewDefaultManagerWithDeps(config, fs, git)

	// Create the backlog directory
	err := fs.CreateDirectory(config.BacklogDir)
	require.NoError(t, err)

	// Create a work item first
	req := CreateRequest{Type: TypeFeature, Name: "test-feature"}
	_, err = manager.CreateWorkItem(context.Background(), req)
	require.NoError(t, err)

	// Assign work item
	err = manager.AssignWorkItem(context.Background(), "feature-test-feature", "john.doe")
	require.NoError(t, err)

	// Verify assignment
	item, err := manager.GetWorkItem(context.Background(), "feature-test-feature")
	require.NoError(t, err)
	assert.Equal(t, "john.doe", item.AssignedTo)
}

func TestManagerAdvancePhase(t *testing.T) {
	config := DefaultConfig()
	fs := NewMockFileSystem()
	git := NewNoOpGitClient()
	manager := NewDefaultManagerWithDeps(config, fs, git)

	// Create the backlog directory
	err := fs.CreateDirectory(config.BacklogDir)
	require.NoError(t, err)

	// Create a work item first
	req := CreateRequest{Type: TypeFeature, Name: "test-feature"}
	_, err = manager.CreateWorkItem(context.Background(), req)
	require.NoError(t, err)

	// Advance phase from PLANNING to IMPLEMENTATION
	err = manager.AdvancePhase(context.Background(), "feature-test-feature")
	require.NoError(t, err)

	// Verify phase was advanced
	item, err := manager.GetWorkItem(context.Background(), "feature-test-feature")
	require.NoError(t, err)
	assert.Equal(t, StatusImplementation, item.Status)
	assert.Equal(t, PhaseImplementation, item.Phase)
}

func TestManagerCompleteTask(t *testing.T) {
	config := DefaultConfig()
	fs := NewMockFileSystem()
	git := NewNoOpGitClient()
	manager := NewDefaultManagerWithDeps(config, fs, git)

	// Create the backlog directory
	err := fs.CreateDirectory(config.BacklogDir)
	require.NoError(t, err)

	// Create a work item first
	req := CreateRequest{Type: TypeFeature, Name: "test-feature"}
	_, err = manager.CreateWorkItem(context.Background(), req)
	require.NoError(t, err)

	// Advance to IMPLEMENTATION status (first advance from PLANNING)
	err = manager.AdvancePhase(context.Background(), "feature-test-feature")
	require.NoError(t, err)

	// Verify we're now in implementation phase with IMPLEMENTATION status
	item, err := manager.GetWorkItem(context.Background(), "feature-test-feature")
	require.NoError(t, err)
	assert.Equal(t, StatusImplementation, item.Status)
	assert.Equal(t, PhaseImplementation, item.Phase)

	// Get tasks first
	tasks, err := manager.GetPhaseTasks(context.Background(), "feature-test-feature")
	require.NoError(t, err)

	// Complete all tasks in implementation phase
	for i := range tasks {
		err = manager.CompleteTask(context.Background(), "feature-test-feature", i)
		require.NoError(t, err)
	}

	// Now advance phase again (from IMPLEMENTATION to REVIEW)
	err = manager.AdvancePhase(context.Background(), "feature-test-feature")
	require.NoError(t, err)

	// Verify phase was advanced to review
	item, err = manager.GetWorkItem(context.Background(), "feature-test-feature")
	require.NoError(t, err)
	assert.Equal(t, StatusReview, item.Status)
	assert.Equal(t, PhaseReview, item.Phase)
}

func TestManagerGetPhaseTasks(t *testing.T) {
	config := DefaultConfig()
	fs := NewMockFileSystem()
	git := NewNoOpGitClient()
	manager := NewDefaultManagerWithDeps(config, fs, git)

	// Create the backlog directory
	err := fs.CreateDirectory(config.BacklogDir)
	require.NoError(t, err)

	// Create a work item first
	req := CreateRequest{Type: TypeFeature, Name: "test-feature"}
	_, err = manager.CreateWorkItem(context.Background(), req)
	require.NoError(t, err)

	// Set phase to planning
	err = manager.SetPhase(context.Background(), "feature-test-feature", PhasePlanning)
	require.NoError(t, err)

	// Get phase tasks
	tasks, err := manager.GetPhaseTasks(context.Background(), "feature-test-feature")
	require.NoError(t, err)
	assert.True(t, len(tasks) > 0) // Should have tasks for planning phase
}

func TestManagerGetProgressMetrics(t *testing.T) {
	config := DefaultConfig()
	fs := NewMockFileSystem()
	git := NewNoOpGitClient()
	manager := NewDefaultManagerWithDeps(config, fs, git)

	// Create the backlog directory
	err := fs.CreateDirectory(config.BacklogDir)
	require.NoError(t, err)

	// Create a work item first
	req := CreateRequest{Type: TypeFeature, Name: "test-feature"}
	_, err = manager.CreateWorkItem(context.Background(), req)
	require.NoError(t, err)

	// Get progress metrics
	metrics, err := manager.GetProgressMetrics(context.Background(), "feature-test-feature")
	require.NoError(t, err)
	assert.Equal(t, "feature-test-feature", metrics.Name)
	assert.True(t, metrics.TotalTasks >= 0)
}

func TestManagerArchiveWorkItem(t *testing.T) {
	config := DefaultConfig()
	fs := NewMockFileSystem()
	git := NewNoOpGitClient()
	manager := NewDefaultManagerWithDeps(config, fs, git)

	// Create the backlog directory
	err := fs.CreateDirectory(config.BacklogDir)
	require.NoError(t, err)

	// Create a work item first
	req := CreateRequest{Type: TypeFeature, Name: "test-feature"}
	_, err = manager.CreateWorkItem(context.Background(), req)
	require.NoError(t, err)

	// Archive the work item
	err = manager.ArchiveWorkItem(context.Background(), "feature-test-feature")
	require.NoError(t, err)

	// Verify the work item is no longer in backlog
	workItemPath := filepath.Join(config.BacklogDir, "feature-test-feature")
	assert.False(t, fs.DirectoryExists(workItemPath))

	// Verify the work item was moved to completed directory
	completedPath := filepath.Join(config.CompletedDir, "feature-test-feature")
	assert.True(t, fs.DirectoryExists(completedPath))
}

func TestManagerAdvancePhaseThroughWorkflow(t *testing.T) {
	config := DefaultConfig()
	fs := NewMockFileSystem()
	git := NewNoOpGitClient()
	manager := NewDefaultManagerWithDeps(config, fs, git)

	// Create the backlog directory
	err := fs.CreateDirectory(config.BacklogDir)
	require.NoError(t, err)

	// Create a work item first
	req := CreateRequest{Type: TypeFeature, Name: "test-feature"}
	_, err = manager.CreateWorkItem(context.Background(), req)
	require.NoError(t, err)

	// Test phase advancement through the entire workflow
	testCases := []struct {
		expectedStatus ItemStatus
		expectedPhase  WorkPhase
		description    string
	}{
		{StatusImplementation, PhaseImplementation, "PLANNING -> IMPLEMENTATION"},
		{StatusReview, PhaseReview, "IMPLEMENTATION -> REVIEW"},
		{StatusCompleted, PhaseReview, "REVIEW -> COMPLETED"},
	}

	for i, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// Complete all tasks in current phase before advancing (except for first advance)
			if i > 0 {
				tasks, err := manager.GetPhaseTasks(context.Background(), "feature-test-feature")
				require.NoError(t, err)
				for j := range tasks {
					err = manager.CompleteTask(context.Background(), "feature-test-feature", j)
					require.NoError(t, err)
				}
			}

			// Advance phase
			err = manager.AdvancePhase(context.Background(), "feature-test-feature")
			require.NoError(t, err)

			// Verify status and phase
			item, err := manager.GetWorkItem(context.Background(), "feature-test-feature")
			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, item.Status, "Status mismatch for %s", tc.description)
			assert.Equal(t, tc.expectedPhase, item.Phase, "Phase mismatch for %s", tc.description)
		})
	}
}
