package pm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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

func TestPredictCompletionTime(t *testing.T) {
	fs := NewMockFileSystem()
	pt := NewProgressTracker(fs)

	// Test with completed work item
	metrics := WorkItemMetrics{
		OverallProgress: 100,
		UpdatedAt:       time.Now(),
	}
	completionTime, status := pt.PredictCompletionTime(metrics)
	assert.Equal(t, metrics.UpdatedAt, completionTime)
	assert.Equal(t, "Already completed", status)

	// Test with insufficient data
	metrics = WorkItemMetrics{
		OverallProgress: 0,
		TotalTasks:      5,
		CompletedTasks:  0,
	}
	completionTime, status = pt.PredictCompletionTime(metrics)
	assert.True(t, completionTime.IsZero())
	assert.Equal(t, "Insufficient data for prediction", status)
}

func TestGetPhaseEfficiency(t *testing.T) {
	fs := NewMockFileSystem()
	pt := NewProgressTracker(fs)

	metrics := WorkItemMetrics{
		PhaseProgress: []PhaseProgress{
			{Phase: PhaseDiscovery, TimeSpent: time.Hour},
			{Phase: PhasePlanning, TimeSpent: 0},
		},
	}

	efficiency := pt.GetPhaseEfficiency(metrics)
	assert.Equal(t, 1.0, efficiency[PhaseDiscovery])
	assert.Equal(t, 0.0, efficiency[PhasePlanning])
}
