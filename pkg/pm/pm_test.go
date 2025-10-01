package pm

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.Equal(t, "wiki", config.DocsDir)
	assert.Equal(t, "wiki/work-items/templates", config.TemplatesDir)
	assert.Equal(t, "wiki/work-items/backlog", config.BacklogDir)
	assert.Equal(t, "wiki/work-items/completed", config.CompletedDir)
}

func TestItemTypeString(t *testing.T) {
	assert.Equal(t, "feature", string(TypeFeature))
	assert.Equal(t, "bug", string(TypeBug))
	assert.Equal(t, "experiment", string(TypeExperiment))
}

func TestItemStatusString(t *testing.T) {
	assert.Equal(t, "PROPOSED", string(StatusProposed))
	assert.Equal(t, "IN_PROGRESS_DISCOVERY", string(StatusInProgressDiscovery))
	assert.Equal(t, "IN_PROGRESS_PLANNING", string(StatusInProgressPlanning))
	assert.Equal(t, "IN_PROGRESS_EXECUTION", string(StatusInProgressExecution))
	assert.Equal(t, "IN_PROGRESS_CLEANUP", string(StatusInProgressCleanup))
	assert.Equal(t, "IN_PROGRESS_REVIEW", string(StatusInProgressReview))
	assert.Equal(t, "COMPLETED", string(StatusCompleted))
}

func TestWorkPhaseString(t *testing.T) {
	assert.Equal(t, "discovery", string(PhaseDiscovery))
	assert.Equal(t, "planning", string(PhasePlanning))
	assert.Equal(t, "execution", string(PhaseExecution))
	assert.Equal(t, "cleanup", string(PhaseCleanup))
}

func TestWorkItemError(t *testing.T) {
	err := &WorkItemError{
		Op:   "create",
		Name: "test-feature",
		Err:  assert.AnError,
	}

	expected := "go-pm create test-feature: assert.AnError general error for testing"
	assert.Equal(t, expected, err.Error())
	assert.Equal(t, assert.AnError, err.Unwrap())
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "name",
		Value:   "",
		Message: "name cannot be empty",
	}

	expected := "validation error for name '': name cannot be empty"
	assert.Equal(t, expected, err.Error())
}

func TestPhaseError(t *testing.T) {
	err := &PhaseError{
		WorkItem:     "test-feature",
		CurrentPhase: PhaseDiscovery,
		TargetPhase:  PhasePlanning,
		Reason:       "tasks not completed",
	}

	expected := "cannot advance test-feature from discovery to planning: tasks not completed"
	assert.Equal(t, expected, err.Error())
}

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

func TestTemplateProcessing(t *testing.T) {
	fs := NewMockFileSystem()
	config := DefaultConfig()
	tp := NewTemplateProcessor(fs, config)

	// Test feature template processing
	err := tp.ProcessTemplate("feature-template", "/tmp/test-feature.md", "user-auth", TypeFeature)
	require.NoError(t, err)

	content, err := fs.ReadFile("/tmp/test-feature.md")
	require.NoError(t, err)
	assert.Contains(t, string(content), "Feature: user-auth")
	assert.Contains(t, string(content), "## Status: PROPOSED")
}

func TestTemplateProcessingBug(t *testing.T) {
	fs := NewMockFileSystem()
	config := DefaultConfig()
	tp := NewTemplateProcessor(fs, config)

	err := tp.ProcessTemplate("bug-template", "/tmp/test-bug.md", "null-pointer", TypeBug)
	require.NoError(t, err)

	content, err := fs.ReadFile("/tmp/test-bug.md")
	require.NoError(t, err)
	assert.Contains(t, string(content), "Bug: null-pointer")
}

func TestTemplateProcessingExperiment(t *testing.T) {
	fs := NewMockFileSystem()
	config := DefaultConfig()
	tp := NewTemplateProcessor(fs, config)

	err := tp.ProcessTemplate("experiment-template", "/tmp/test-experiment.md", "ai-assistant", TypeExperiment)
	require.NoError(t, err)

	content, err := fs.ReadFile("/tmp/test-experiment.md")
	require.NoError(t, err)
	assert.Contains(t, string(content), "Experiment: ai-assistant")
}

func TestTemplateProcessingInvalidType(t *testing.T) {
	fs := NewMockFileSystem()
	config := DefaultConfig()
	tp := NewTemplateProcessor(fs, config)

	err := tp.ProcessTemplate("invalid-template", "/tmp/test-invalid.md", "test", ItemType("invalid"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported item type")
}

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

func TestProgressTracker(t *testing.T) {
	fs := NewMockFileSystem()
	pt := NewProgressTracker(fs)

	workItem := WorkItem{
		Name: "test-feature",
		Tasks: []Task{
			{Description: "Task 1", Completed: true, Phase: PhaseDiscovery},
			{Description: "Task 2", Completed: false, Phase: PhaseDiscovery},
			{Description: "Task 3", Completed: true, Phase: PhasePlanning},
		},
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now(),
	}

	metrics := pt.CalculateWorkItemMetrics(&workItem)
	assert.Equal(t, "test-feature", metrics.Name)
	assert.Equal(t, 3, metrics.TotalTasks)
	assert.Equal(t, 2, metrics.CompletedTasks)
	assert.Equal(t, 66, metrics.OverallProgress) // 2/3 * 100 rounded down
	assert.Len(t, metrics.PhaseProgress, 4)      // All phases
}

func TestPhaseProgressCalculation(t *testing.T) {
	fs := NewMockFileSystem()
	pt := NewProgressTracker(fs)

	workItem := WorkItem{
		Tasks: []Task{
			{Description: "Task 1", Completed: true, Phase: PhaseDiscovery},
			{Description: "Task 2", Completed: false, Phase: PhaseDiscovery},
		},
	}

	progress := pt.CalculatePhaseProgress(&workItem, PhaseDiscovery)
	assert.Equal(t, PhaseDiscovery, progress.Phase)
	assert.Equal(t, 2, progress.TotalTasks)
	assert.Equal(t, 1, progress.CompletedTasks)
	assert.Equal(t, 50, progress.ProgressPercent)
}

func TestProgressReport(t *testing.T) {
	fs := NewMockFileSystem()
	pt := NewProgressTracker(fs)

	metrics := WorkItemMetrics{
		Name:            "test-feature",
		TotalTasks:      4,
		CompletedTasks:  2,
		OverallProgress: 50,
		CreatedAt:       time.Now().Add(-time.Hour),
		UpdatedAt:       time.Now(),
	}

	report := pt.GetProgressReport(metrics)
	assert.Contains(t, report, "Progress Report for test-feature")
	assert.Contains(t, report, "Overall Progress: 50%")
	assert.Contains(t, report, "2/4 tasks completed")
}

func TestWorkItemParser(t *testing.T) {
	fs := NewMockFileSystem()
	parser := NewWorkItemParser(fs)

	// Create a mock README file
	content := `# Feature: user-auth

## Status: IN_PROGRESS_DISCOVERY
## Phase: discovery
## Progress: 25%
## Assigned To: agent

## Overview
User authentication feature

---

## Discovery Phase

### Tasks
- [x] Analyze requirements
- [ ] Interview stakeholders
`

	fs.WriteFile("/tmp/test.md", []byte(content)) //nolint:errcheck

	item, err := parser.ParseWorkItem("feature-user-auth", "/tmp/test.md")
	require.NoError(t, err)

	assert.Equal(t, "feature-user-auth", item.Name)
	assert.Equal(t, "user-auth", item.Title)
	assert.Equal(t, StatusInProgressDiscovery, item.Status)
	assert.Equal(t, PhaseDiscovery, item.Phase)
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

## Status: PROPOSED
## Phase: discovery
## Progress: 0%
`

	fs.WriteFile("/tmp/test.md", []byte(content)) //nolint:errcheck

	err := updater.UpdateStatus("/tmp/test.md", StatusInProgressPlanning)
	require.NoError(t, err)

	updated, err := fs.ReadFile("/tmp/test.md")
	require.NoError(t, err)
	assert.Contains(t, string(updated), "## Status: IN_PROGRESS_PLANNING")
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
	assert.Equal(t, StatusProposed, item.Status)
	assert.Equal(t, PhaseDiscovery, item.Phase)
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
	err = manager.UpdateStatus(context.Background(), "feature-test-feature", StatusInProgressDiscovery)
	require.NoError(t, err)

	// Verify status was updated
	item, err := manager.GetWorkItem(context.Background(), "feature-test-feature")
	require.NoError(t, err)
	assert.Equal(t, StatusInProgressDiscovery, item.Status)
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

	// Advance phase from PROPOSED to IN_PROGRESS_DISCOVERY
	err = manager.AdvancePhase(context.Background(), "feature-test-feature")
	require.NoError(t, err)

	// Verify phase was advanced
	item, err := manager.GetWorkItem(context.Background(), "feature-test-feature")
	require.NoError(t, err)
	assert.Equal(t, StatusInProgressDiscovery, item.Status)
	assert.Equal(t, PhaseDiscovery, item.Phase)
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

	// Advance to IN_PROGRESS_DISCOVERY status (first advance from PROPOSED)
	err = manager.AdvancePhase(context.Background(), "feature-test-feature")
	require.NoError(t, err)

	// Verify we're now in discovery phase with IN_PROGRESS_DISCOVERY status
	item, err := manager.GetWorkItem(context.Background(), "feature-test-feature")
	require.NoError(t, err)
	assert.Equal(t, StatusInProgressDiscovery, item.Status)
	assert.Equal(t, PhaseDiscovery, item.Phase)

	// Get tasks first
	tasks, err := manager.GetPhaseTasks(context.Background(), "feature-test-feature")
	require.NoError(t, err)

	// Complete all tasks in discovery phase
	for i := range tasks {
		err = manager.CompleteTask(context.Background(), "feature-test-feature", i)
		require.NoError(t, err)
	}

	// Now advance phase again (from IN_PROGRESS_DISCOVERY to IN_PROGRESS_PLANNING)
	err = manager.AdvancePhase(context.Background(), "feature-test-feature")
	require.NoError(t, err)

	// Verify phase was advanced to planning
	item, err = manager.GetWorkItem(context.Background(), "feature-test-feature")
	require.NoError(t, err)
	assert.Equal(t, StatusInProgressPlanning, item.Status)
	assert.Equal(t, PhasePlanning, item.Phase)
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

	// Set phase to discovery
	err = manager.SetPhase(context.Background(), "feature-test-feature", PhaseDiscovery)
	require.NoError(t, err)

	// Get phase tasks
	tasks, err := manager.GetPhaseTasks(context.Background(), "feature-test-feature")
	require.NoError(t, err)
	assert.True(t, len(tasks) > 0) // Should have tasks for discovery phase
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
		{StatusInProgressDiscovery, PhaseDiscovery, "PROPOSED -> IN_PROGRESS_DISCOVERY"},
		{StatusInProgressPlanning, PhasePlanning, "IN_PROGRESS_DISCOVERY -> IN_PROGRESS_PLANNING"},
		{StatusInProgressExecution, PhaseExecution, "IN_PROGRESS_PLANNING -> IN_PROGRESS_EXECUTION"},
		{StatusInProgressCleanup, PhaseCleanup, "IN_PROGRESS_EXECUTION -> IN_PROGRESS_CLEANUP"},
		{StatusInProgressReview, PhaseCleanup, "IN_PROGRESS_CLEANUP -> IN_PROGRESS_REVIEW"},
		{StatusCompleted, PhaseCleanup, "IN_PROGRESS_REVIEW -> COMPLETED"},
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

// MockFileSystem is a mock implementation of FileSystem for testing
type MockFileSystem struct {
	files map[string][]byte
	dirs  map[string]bool
}

func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
	}
}

func (fs *MockFileSystem) ReadFile(path string) ([]byte, error) {
	if content, exists := fs.files[path]; exists {
		return content, nil
	}
	return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
}

func (fs *MockFileSystem) WriteFile(path string, content []byte) error {
	fs.files[path] = content
	return nil
}

func (fs *MockFileSystem) FileExists(path string) bool {
	_, exists := fs.files[path]
	return exists
}

func (fs *MockFileSystem) DirectoryExists(path string) bool {
	return fs.dirs[path]
}

func (fs *MockFileSystem) CreateDirectory(path string) error {
	fs.dirs[path] = true
	return nil
}

func (fs *MockFileSystem) ListDirectories(path string) ([]string, error) {
	var dirs []string
	for dir := range fs.dirs {
		if strings.HasPrefix(dir, path+"/") || dir == path {
			// Extract just the directory name, not the full path
			if dir != path {
				relPath := strings.TrimPrefix(dir, path+"/")
				if !strings.Contains(relPath, "/") {
					dirs = append(dirs, relPath)
				}
			}
		}
	}
	return dirs, nil
}

func (fs *MockFileSystem) CopyFile(src, dst string) error {
	if content, exists := fs.files[src]; exists {
		fs.files[dst] = content
		return nil
	}
	return &os.PathError{Op: "open", Path: src, Err: os.ErrNotExist}
}

func (fs *MockFileSystem) ListFiles(path string) ([]string, error) {
	var files []string
	for file := range fs.files {
		if strings.HasPrefix(file, path) {
			files = append(files, file)
		}
	}
	return files, nil
}

func (fs *MockFileSystem) MoveDirectory(src, dst string) error {
	// Mark destination as existing and remove source
	fs.dirs[dst] = true
	delete(fs.dirs, src)
	return nil
}
