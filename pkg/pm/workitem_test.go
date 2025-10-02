package pm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
