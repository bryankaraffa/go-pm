package pm

import (
	"fmt"
	"time"
)

// ProgressTracker provides progress tracking and metrics functionality.
// It calculates completion percentages and phase-specific metrics.
type ProgressTracker struct {
	fs FileSystem
}

// NewProgressTracker creates a new progress tracker.
// Requires a FileSystem implementation for file operations.
func NewProgressTracker(fs FileSystem) *ProgressTracker {
	return &ProgressTracker{fs: fs}
}

// CalculatePhaseProgress calculates progress for a specific phase.
// Returns metrics including task counts and completion percentage for the given phase.
func (pt *ProgressTracker) CalculatePhaseProgress(workItem *WorkItem, phase WorkPhase) PhaseProgress {
	var phaseTasks []Task
	for _, task := range workItem.Tasks {
		if task.Phase == phase {
			phaseTasks = append(phaseTasks, task)
		}
	}

	completed := 0
	for _, task := range phaseTasks {
		if task.Completed {
			completed++
		}
	}

	progressPercent := 0
	if len(phaseTasks) > 0 {
		progressPercent = (completed * 100) / len(phaseTasks)
	}

	return PhaseProgress{
		Phase:           phase,
		TotalTasks:      len(phaseTasks),
		CompletedTasks:  completed,
		ProgressPercent: progressPercent,
		TimeSpent:       pt.calculateTimeSpentInPhase(workItem, phase),
	}
}

// CalculateWorkItemMetrics calculates comprehensive metrics for a work item.
// Returns detailed statistics including task completion, phase progress, and timing.
func (pt *ProgressTracker) CalculateWorkItemMetrics(workItem *WorkItem) WorkItemMetrics {
	totalTasks := len(workItem.Tasks)
	completedTasks := 0
	for _, task := range workItem.Tasks {
		if task.Completed {
			completedTasks++
		}
	}

	overallProgress := 0
	if totalTasks > 0 {
		overallProgress = (completedTasks * 100) / totalTasks
	}

	// Calculate progress for each phase
	var phaseProgress []PhaseProgress
	for _, phase := range []WorkPhase{PhaseDiscovery, PhasePlanning, PhaseExecution, PhaseCleanup} {
		phaseProgress = append(phaseProgress, pt.CalculatePhaseProgress(workItem, phase))
	}

	totalTimeSpent := time.Duration(0)
	for _, pp := range phaseProgress {
		totalTimeSpent += pp.TimeSpent
	}

	return WorkItemMetrics{
		Name:            workItem.Name,
		TotalTasks:      totalTasks,
		CompletedTasks:  completedTasks,
		OverallProgress: overallProgress,
		PhaseProgress:   phaseProgress,
		TotalTimeSpent:  totalTimeSpent,
		CreatedAt:       workItem.CreatedAt,
		UpdatedAt:       workItem.UpdatedAt,
	}
}

// calculateTimeSpentInPhase estimates time spent in a phase based on work item timestamps
// This is a simplified implementation - in a real system this would track actual time
func (pt *ProgressTracker) calculateTimeSpentInPhase(workItem *WorkItem, phase WorkPhase) time.Duration {
	// For now, return a simple estimate based on phase order
	// In a real implementation, this would track actual time spent in each phase
	if workItem.CreatedAt.IsZero() {
		return 0
	}

	age := time.Since(workItem.CreatedAt)
	phaseIndex := 0

	switch phase {
	case PhaseDiscovery:
		phaseIndex = 0
	case PhasePlanning:
		phaseIndex = 1
	case PhaseExecution:
		phaseIndex = 2
	case PhaseCleanup:
		phaseIndex = 3
	}

	// Rough estimate: assume equal time distribution across completed phases
	// If we're in planning phase, assume we've spent time in discovery
	estimatedPhases := phaseIndex + 1
	if estimatedPhases > 0 {
		return age / time.Duration(estimatedPhases)
	}

	return 0
}

// GetProgressReport generates a human-readable progress report.
// Returns a formatted string with progress statistics and phase information.
func (pt *ProgressTracker) GetProgressReport(metrics WorkItemMetrics) string {
	report := fmt.Sprintf("Progress Report for %s\n", metrics.Name)
	report += "================================\n"
	report += fmt.Sprintf("Overall Progress: %d%% (%d/%d tasks completed)\n",
		metrics.OverallProgress, metrics.CompletedTasks, metrics.TotalTasks)
	report += fmt.Sprintf("Total Time Spent: %v\n", metrics.TotalTimeSpent.Round(time.Hour))
	report += fmt.Sprintf("Created: %s\n", metrics.CreatedAt.Format("2006-01-02 15:04"))
	report += fmt.Sprintf("Updated: %s\n\n", metrics.UpdatedAt.Format("2006-01-02 15:04"))

	report += "Phase Progress:\n"
	for _, pp := range metrics.PhaseProgress {
		report += fmt.Sprintf("  %s: %d%% (%d/%d tasks)",
			pp.Phase, pp.ProgressPercent, pp.CompletedTasks, pp.TotalTasks)
		if pp.TimeSpent > 0 {
			report += fmt.Sprintf(" - Spent: %v", pp.TimeSpent.Round(time.Hour))
		}
		report += "\n"
	}

	return report
}

// PredictCompletionTime estimates when the work item will be completed.
// Returns the predicted completion time and a status message.
func (pt *ProgressTracker) PredictCompletionTime(metrics WorkItemMetrics) (time.Time, string) {
	if metrics.OverallProgress >= 100 {
		return metrics.UpdatedAt, "Already completed"
	}

	// Calculate remaining work
	remainingTasks := metrics.TotalTasks - metrics.CompletedTasks
	if remainingTasks <= 0 {
		return metrics.UpdatedAt, "All tasks completed"
	}

	// Use actual time data for prediction if available
	totalActualTime := time.Duration(0)
	for _, pp := range metrics.PhaseProgress {
		totalActualTime += pp.TimeSpent
	}

	// If we have actual time data, use it for prediction
	if totalActualTime > 0 && metrics.OverallProgress > 0 {
		avgTimePerPercent := totalActualTime / time.Duration(metrics.OverallProgress)
		remainingPercent := 100 - metrics.OverallProgress
		estimatedRemaining := avgTimePerPercent * time.Duration(remainingPercent)
		completionTime := time.Now().Add(estimatedRemaining)
		return completionTime, fmt.Sprintf("Based on current progress rate: %v remaining", estimatedRemaining.Round(time.Hour))
	}

	// No prediction available without historical data
	return time.Time{}, "Insufficient data for prediction"
}

// GetPhaseEfficiency calculates how efficiently time is being used in each phase.
// Returns a map of phase to efficiency ratio (currently basic implementation).
func (pt *ProgressTracker) GetPhaseEfficiency(metrics WorkItemMetrics) map[WorkPhase]float64 {
	efficiency := make(map[WorkPhase]float64)

	for _, pp := range metrics.PhaseProgress {
		if pp.TimeSpent > 0 {
			// Without estimated times, we can only report that time has been spent
			// Efficiency calculation would require estimated vs actual comparison
			efficiency[pp.Phase] = 1.0 // Placeholder - time has been spent
		} else {
			// No progress yet
			efficiency[pp.Phase] = 0.0
		}
	}

	return efficiency
}
