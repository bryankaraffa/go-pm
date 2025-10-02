package pm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
