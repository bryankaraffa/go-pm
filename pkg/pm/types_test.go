package pm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestItemTypeString(t *testing.T) {
	assert.Equal(t, "feature", string(TypeFeature))
	assert.Equal(t, "bug", string(TypeBug))
	assert.Equal(t, "experiment", string(TypeExperiment))
}

func TestItemStatusString(t *testing.T) {
	// New 4-status model tests
	assert.Equal(t, "PLANNING", string(StatusPlanning))
	assert.Equal(t, "IMPLEMENTATION", string(StatusImplementation))
	assert.Equal(t, "REVIEW", string(StatusReview))
	assert.Equal(t, "COMPLETED", string(StatusCompleted))
}

func TestWorkPhaseString(t *testing.T) {
	// New 3-phase model tests
	assert.Equal(t, "planning", string(PhasePlanning))
	assert.Equal(t, "implementation", string(PhaseImplementation))
	assert.Equal(t, "review", string(PhaseReview))
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

func TestMilestone(t *testing.T) {
	// Test Milestone struct creation and fields
	milestone := &Milestone{
		Name:        "Design approved",
		Description: "Stakeholder approved the technical design",
		Achieved:    true,
		AchievedAt:  time.Now(),
	}

	assert.Equal(t, "Design approved", milestone.Name)
	assert.Equal(t, "Stakeholder approved the technical design", milestone.Description)
	assert.True(t, milestone.Achieved)
	assert.False(t, milestone.AchievedAt.IsZero())
}

func TestPhaseError(t *testing.T) {
	err := &PhaseError{
		WorkItem:     "test-feature",
		CurrentPhase: PhasePlanning,
		TargetPhase:  PhaseImplementation,
		Reason:       "tasks not completed",
	}

	expected := "cannot advance test-feature from planning to implementation: tasks not completed"
	assert.Equal(t, expected, err.Error())
}
