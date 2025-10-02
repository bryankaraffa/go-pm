package pm

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.True(t, config.AutoDetectRepoRoot)
	assert.Equal(t, 7, config.PhaseTimeoutDays)
	assert.False(t, config.EnableGit)
	// BacklogDir and CompletedDir should be absolute paths
	assert.NotEmpty(t, config.BacklogDir)
	assert.NotEmpty(t, config.CompletedDir)
	assert.True(t, filepath.IsAbs(config.BacklogDir))
	assert.True(t, filepath.IsAbs(config.CompletedDir))
}

func TestConfigWithEnvVars(t *testing.T) {
	// Save original env vars
	origAutoDetect := os.Getenv("PM_AUTO_DETECT_REPO_ROOT")
	origBacklogDir := os.Getenv("PM_BACKLOG_DIR")
	origEnableGit := os.Getenv("PM_ENABLE_GIT")
	defer func() {
		_ = os.Setenv("PM_AUTO_DETECT_REPO_ROOT", origAutoDetect)
		_ = os.Setenv("PM_BACKLOG_DIR", origBacklogDir)
		_ = os.Setenv("PM_ENABLE_GIT", origEnableGit)
	}()

	// Set test env vars
	_ = os.Setenv("PM_AUTO_DETECT_REPO_ROOT", "false")
	_ = os.Setenv("PM_BACKLOG_DIR", "custom-backlog")
	_ = os.Setenv("PM_ENABLE_GIT", "true")

	config := DefaultConfig()
	assert.False(t, config.AutoDetectRepoRoot)
	assert.Equal(t, "custom-backlog", config.BacklogDir)
	assert.True(t, config.EnableGit)
}

func TestConfigWithAbsolutePathEnvVars(t *testing.T) {
	// Save original env vars
	origBacklogDir := os.Getenv("PM_BACKLOG_DIR")
	origCompletedDir := os.Getenv("PM_COMPLETED_DIR")
	defer func() {
		_ = os.Setenv("PM_BACKLOG_DIR", origBacklogDir)
		_ = os.Setenv("PM_COMPLETED_DIR", origCompletedDir)
	}()

	// Set test env vars with absolute paths (should be used as-is)
	_ = os.Setenv("PM_BACKLOG_DIR", "/tmp/test/absolute/backlog")
	_ = os.Setenv("PM_COMPLETED_DIR", "/tmp/test/absolute/completed")

	config := DefaultConfig()
	assert.Equal(t, "/tmp/test/absolute/backlog", config.BacklogDir)
	assert.Equal(t, "/tmp/test/absolute/completed", config.CompletedDir)
}

func TestConfigWithFile(t *testing.T) {
	// Create temp config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := `
auto_detect_repo_root: false
backlog_dir: "custom-backlog"
enable_git: true
phase_timeout_days: 10
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Change to temp dir so viper finds the config
	origWd, _ := os.Getwd()
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(origWd)
	}()

	// Reload config to pick up the new config file
	reloadConfigForTesting()

	config := DefaultConfig()
	assert.False(t, config.AutoDetectRepoRoot)
	assert.Contains(t, config.BacklogDir, "custom-backlog")
	assert.True(t, config.EnableGit)
	assert.Equal(t, 10, config.PhaseTimeoutDays)
}

func TestDetectRepoRoot(t *testing.T) {
	root := detectRepoRoot()
	// Should return "." if not in git repo or git fails
	assert.NotEmpty(t, root)
	// In this repo, it should detect
	if strings.HasSuffix(root, "go-pm-cli") {
		assert.Contains(t, root, "go-pm-cli")
	}
}

func TestAutoDetectFromSubdirectory(t *testing.T) {
	// Create a temporary directory structure to simulate a git repo
	tempDir, err := os.MkdirTemp("", "go-pm-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	// Initialize a git repo in tempDir
	err = exec.Command("git", "init", tempDir).Run()
	require.NoError(t, err)

	// Create a subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	// Change to the subdirectory
	origWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(origWd)
	}()

	err = os.Chdir(subDir)
	require.NoError(t, err)

	// Reload config to pick up the new working directory context
	reloadConfigForTesting()

	// Now test that DefaultConfig detects the repo root correctly
	config := DefaultConfig()
	// The backlog and completed dirs should be absolute paths under the detected repo root
	expectedBacklogDir := filepath.Join(tempDir, "work-items", "backlog")
	expectedCompletedDir := filepath.Join(tempDir, "work-items", "completed")
	assert.Equal(t, expectedBacklogDir, config.BacklogDir)
	assert.Equal(t, expectedCompletedDir, config.CompletedDir)

	// Create manager and test full lifecycle
	manager := NewDefaultManager(config)

	ctx := context.Background()

	// 1. Create a work item
	req := CreateRequest{Type: TypeFeature, Name: "auto-detect-test"}
	item, err := manager.CreateWorkItem(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "feature-auto-detect-test", item.Name)
	assert.Equal(t, StatusProposed, item.Status)
	assert.Equal(t, PhaseDiscovery, item.Phase)

	// Verify the work item was created in the correct location
	expectedPath := filepath.Join(tempDir, "work-items", "backlog", "feature-auto-detect-test", "README.md")
	assert.Equal(t, expectedPath, item.Path)

	// 2. List work items
	items, err := manager.ListWorkItems(ctx, ListFilter{})
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "feature-auto-detect-test", items[0].Name)

	// 3. Advance phase (proposed -> discovery)
	err = manager.AdvancePhase(ctx, "feature-auto-detect-test")
	require.NoError(t, err)

	item, err = manager.GetWorkItem(ctx, "feature-auto-detect-test")
	require.NoError(t, err)
	assert.Equal(t, PhaseDiscovery, item.Phase)
	assert.Equal(t, StatusInProgressDiscovery, item.Status)

	// 4. Update progress
	err = manager.UpdateProgress(ctx, "feature-auto-detect-test", 25)
	require.NoError(t, err)

	item, err = manager.GetWorkItem(ctx, "feature-auto-detect-test")
	require.NoError(t, err)
	assert.Equal(t, 25, item.Progress)

	// 5. Assign to agent
	err = manager.AssignWorkItem(ctx, "feature-auto-detect-test", "test-agent")
	require.NoError(t, err)

	item, err = manager.GetWorkItem(ctx, "feature-auto-detect-test")
	require.NoError(t, err)
	assert.Equal(t, "test-agent", item.AssignedTo)

	// 6. Get phase tasks and complete them all
	tasks, err := manager.GetPhaseTasks(ctx, "feature-auto-detect-test")
	require.NoError(t, err)
	assert.NotEmpty(t, tasks) // Should have tasks for discovery phase

	// Complete all tasks
	for i := 0; i < len(tasks); i++ {
		err = manager.CompleteTask(ctx, "feature-auto-detect-test", i)
		require.NoError(t, err)
	}

	// 8. Continue advancing through all phases
	phases := []WorkPhase{PhasePlanning, PhaseExecution, PhaseCleanup}
	statuses := []ItemStatus{StatusInProgressPlanning, StatusInProgressExecution, StatusInProgressCleanup}

	for i, phase := range phases {
		err = manager.AdvancePhase(ctx, "feature-auto-detect-test")
		require.NoError(t, err)

		item, err = manager.GetWorkItem(ctx, "feature-auto-detect-test")
		require.NoError(t, err)
		assert.Equal(t, phase, item.Phase)
		assert.Equal(t, statuses[i], item.Status)

		// Update progress for each phase
		progress := 25 + i*25
		err = manager.UpdateProgress(ctx, "feature-auto-detect-test", progress)
		require.NoError(t, err)

		// Get and complete all tasks in this phase
		tasks, err = manager.GetPhaseTasks(ctx, "feature-auto-detect-test")
		require.NoError(t, err)
		for j := 0; j < len(tasks); j++ {
			err = manager.CompleteTask(ctx, "feature-auto-detect-test", j)
			require.NoError(t, err)
		}
	}

	// 9. Final advance to completed (need to go through review phase)
	err = manager.AdvancePhase(ctx, "feature-auto-detect-test")
	require.NoError(t, err)

	// Complete review tasks if any
	tasks, err = manager.GetPhaseTasks(ctx, "feature-auto-detect-test")
	require.NoError(t, err)
	for j := 0; j < len(tasks); j++ {
		err = manager.CompleteTask(ctx, "feature-auto-detect-test", j)
		require.NoError(t, err)
	}

	// Final advance to completed
	err = manager.AdvancePhase(ctx, "feature-auto-detect-test")
	require.NoError(t, err)

	item, err = manager.GetWorkItem(ctx, "feature-auto-detect-test")
	require.NoError(t, err)
	assert.Equal(t, PhaseCleanup, item.Phase)
	assert.Equal(t, StatusCompleted, item.Status)

	// 10. Archive the work item
	err = manager.ArchiveWorkItem(ctx, "feature-auto-detect-test")
	require.NoError(t, err)

	// Verify it's moved to completed directory
	completedPath := filepath.Join(tempDir, "work-items", "completed", "feature-auto-detect-test")
	_, err = os.Stat(completedPath)
	assert.NoError(t, err) // Directory should exist

	// Verify it's no longer in backlog
	items, err = manager.ListWorkItems(ctx, ListFilter{})
	require.NoError(t, err)
	assert.Len(t, items, 0) // Should be empty since item was archived
}
