// Package pm provides project management functionality for tracking work items
// in documentation-driven development workflows.
package pm

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// Global viper instance for configuration
var configViper *viper.Viper

// initializeViper sets up viper configuration
func initializeViper() {
	// Set config file name and paths
	configViper.SetConfigName("config") // name of config file (without extension)
	configViper.AddConfigPath(".")      // look for config in the working directory
	configViper.AddConfigPath("$HOME")  // look for config in home directory

	// Set default values
	configViper.SetDefault("auto_detect_repo_root", true)
	configViper.SetDefault("base_dir", "") // Will be computed
	configViper.SetDefault("backlog_dir", "work-items/backlog")
	configViper.SetDefault("completed_dir", "work-items/completed")
	configViper.SetDefault("phase_timeout_days", 7)
	configViper.SetDefault("auto_assign_agent", true)
	configViper.SetDefault("enable_git", false)

	// Bind environment variables (these override config file values)
	_ = configViper.BindEnv("auto_detect_repo_root", "PM_AUTO_DETECT_REPO_ROOT")
	_ = configViper.BindEnv("base_dir", "PM_BASE_DIR")
	_ = configViper.BindEnv("backlog_dir", "PM_BACKLOG_DIR")
	_ = configViper.BindEnv("completed_dir", "PM_COMPLETED_DIR")
	_ = configViper.BindEnv("phase_timeout_days", "PM_PHASE_TIMEOUT_DAYS")
	_ = configViper.BindEnv("auto_assign_agent", "PM_AUTO_ASSIGN_AGENT")
	_ = configViper.BindEnv("enable_git", "PM_ENABLE_GIT")

	// Read config file (ignore error if file doesn't exist)
	_ = configViper.ReadInConfig()
}

// init initializes the global viper configuration
func init() {
	configViper = viper.New()
	initializeViper()
}

// reloadConfigForTesting reloads the configuration (used for testing)
func reloadConfigForTesting() {
	// Reset viper instance
	configViper = viper.New()
	initializeViper()
}

// ItemType represents the type of work item
type ItemType string

const (
	TypeFeature    ItemType = "feature"
	TypeBug        ItemType = "bug"
	TypeExperiment ItemType = "experiment"
)

// ItemStatus represents the current status of a work item
type ItemStatus string

const (
	StatusProposed            ItemStatus = "PROPOSED"
	StatusInProgressDiscovery ItemStatus = "IN_PROGRESS_DISCOVERY"
	StatusInProgressPlanning  ItemStatus = "IN_PROGRESS_PLANNING"
	StatusInProgressExecution ItemStatus = "IN_PROGRESS_EXECUTION"
	StatusInProgressCleanup   ItemStatus = "IN_PROGRESS_CLEANUP"
	StatusInProgressReview    ItemStatus = "IN_PROGRESS_REVIEW"
	StatusCompleted           ItemStatus = "COMPLETED"
)

// WorkPhase represents the current phase of work
type WorkPhase string

const (
	PhaseDiscovery WorkPhase = "discovery"
	PhasePlanning  WorkPhase = "planning"
	PhaseExecution WorkPhase = "execution"
	PhaseCleanup   WorkPhase = "cleanup"
)

// Task represents a phase-specific task
type Task struct {
	Description string
	Completed   bool
	Phase       WorkPhase
	AssignedTo  string // "human" or "agent"
}

// WorkItem represents a project management work item with its metadata
type WorkItem struct {
	// Name is the directory name (e.g., "feature-auth")
	Name string
	// Title is the human-readable title extracted from the README
	Title string
	// Type is the work item type (feature, bug, experiment)
	Type ItemType
	// Status is the current workflow status
	Status ItemStatus
	// Phase is the current work phase
	Phase WorkPhase
	// Progress is the completion percentage (0-100)
	Progress int
	// AssignedTo is the current assignee ("human", "agent", or specific agent ID)
	AssignedTo string
	// Path is the full path to the work item directory
	Path string
	// CreatedAt is when the work item was created
	CreatedAt time.Time
	// UpdatedAt is when the work item was last updated
	UpdatedAt time.Time
	// Tasks are the phase-specific task checklists
	Tasks []Task
}

// CreateRequest contains the parameters for creating a new work item
type CreateRequest struct {
	// Type is the work item type to create
	Type ItemType
	// Name is the work item name (without type prefix)
	Name string
}

// ListFilter contains filtering options for listing work items
type ListFilter struct {
	// Status filters by work item status (empty means all statuses)
	Status ItemStatus
	// Type filters by work item type (empty means all types)
	Type ItemType
}

// Manager defines the interface for project management operations
type Manager interface {
	// CreateWorkItem creates a new work item with the given parameters
	CreateWorkItem(ctx context.Context, req CreateRequest) (*WorkItem, error)

	// ListWorkItems returns work items matching the filter criteria
	ListWorkItems(ctx context.Context, filter ListFilter) ([]WorkItem, error)

	// GetWorkItem retrieves a specific work item by name
	GetWorkItem(ctx context.Context, name string) (*WorkItem, error)

	// UpdateStatus updates the status of a work item
	UpdateStatus(ctx context.Context, name string, status ItemStatus) error

	// UpdateProgress updates the progress of a work item
	UpdateProgress(ctx context.Context, name string, progress int) error

	// AssignWorkItem assigns a work item to an assignee
	AssignWorkItem(ctx context.Context, name, assignee string) error

	// AdvancePhase advances a work item to the next phase
	AdvancePhase(ctx context.Context, name string) error

	// SetPhase sets the phase of a work item (admin override)
	SetPhase(ctx context.Context, name string, phase WorkPhase) error

	// GetPhaseTasks returns tasks for the current phase of a work item
	GetPhaseTasks(ctx context.Context, name string) ([]Task, error)

	// CompleteTask marks a task as completed
	CompleteTask(ctx context.Context, name string, taskId int) error

	// GetProgressMetrics returns progress metrics for a work item
	GetProgressMetrics(ctx context.Context, name string) (*WorkItemMetrics, error)

	// ArchiveWorkItem moves a completed work item to the completed directory
	ArchiveWorkItem(ctx context.Context, name string) error
}

// WorkItemError represents an error that occurred during a work item operation
type WorkItemError struct {
	// Op is the operation that failed (create, update, etc.)
	Op string
	// Name is the work item name
	Name string
	// Err is the underlying error
	Err error
}

func (e *WorkItemError) Error() string {
	return fmt.Sprintf("go-pm %s %s: %v", e.Op, e.Name, e.Err)
}

func (e *WorkItemError) Unwrap() error {
	return e.Err
}

// ValidationError represents a validation error for work item data
type ValidationError struct {
	// Field is the field that failed validation
	Field string
	// Value is the invalid value
	Value string
	// Message describes the validation error
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for %s '%s': %s", e.Field, e.Value, e.Message)
}

// PhaseError represents a phase transition error
type PhaseError struct {
	WorkItem     string
	CurrentPhase WorkPhase
	TargetPhase  WorkPhase
	Reason       string
}

func (e *PhaseError) Error() string {
	return fmt.Sprintf("cannot advance %s from %s to %s: %s", e.WorkItem, e.CurrentPhase, e.TargetPhase, e.Reason)
}

// WorkItemMetrics represents comprehensive metrics for a work item.
// It includes task completion statistics, phase progress, and timing information
// used for progress tracking and reporting.
type WorkItemMetrics struct {
	Name            string          // Work item name
	TotalTasks      int             // Total number of tasks across all phases
	CompletedTasks  int             // Number of completed tasks
	OverallProgress int             // Overall progress percentage (0-100)
	PhaseProgress   []PhaseProgress // Progress metrics for each phase
	TotalTimeSpent  time.Duration   // Total time spent on the work item
	CreatedAt       time.Time       // When the work item was created
	UpdatedAt       time.Time       // When the work item was last updated
}

// PhaseProgress represents progress metrics for a specific phase.
// It tracks task completion and time spent within a particular work phase.
type PhaseProgress struct {
	Phase           WorkPhase     // The work phase these metrics apply to
	TotalTasks      int           // Total tasks in this phase
	CompletedTasks  int           // Completed tasks in this phase
	ProgressPercent int           // Progress percentage for this phase (0-100)
	TimeSpent       time.Duration // Time spent working on this phase
}

// Config holds configuration for the PM system
type Config struct {
	// BaseDir is the base directory for all operations (default: auto-detected repo root or ".")
	BaseDir string
	// AutoDetectRepoRoot indicates whether to auto-detect the repository root (default: true)
	AutoDetectRepoRoot bool
	// BacklogDir is the active work items directory (default: "work-items/backlog")
	BacklogDir string
	// CompletedDir is the completed work items directory (default: "work-items/completed")
	CompletedDir string
	// PhaseTimeoutDays is the number of days before phase timeout warning (default: 7)
	PhaseTimeoutDays int
	// AutoAssignAgent indicates whether to auto-assign agents to execution phase (default: true)
	AutoAssignAgent bool
	// EnableGit indicates whether to enable git integration (default: false)
	EnableGit bool
}

// detectRepoRoot attempts to detect the git repository root directory
func detectRepoRoot() string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		// Not a git repo or git not available, use current directory
		return "."
	}
	return string(output[:len(output)-1]) // Remove trailing newline
}

// DefaultConfig returns the default configuration with file and environment variable support
func DefaultConfig() Config {
	// Determine base directory with proper precedence
	baseDir := configViper.GetString("base_dir")
	if baseDir == "" {
		// Not set in config or env, use auto-detection logic
		autoDetect := configViper.GetBool("auto_detect_repo_root")
		if autoDetect {
			baseDir = detectRepoRoot()
		} else {
			baseDir = "./wiki"
		}
	}

	return Config{
		BaseDir:            baseDir,
		AutoDetectRepoRoot: configViper.GetBool("auto_detect_repo_root"),
		BacklogDir:         filepath.Join(baseDir, configViper.GetString("backlog_dir")),
		CompletedDir:       filepath.Join(baseDir, configViper.GetString("completed_dir")),
		PhaseTimeoutDays:   configViper.GetInt("phase_timeout_days"),
		AutoAssignAgent:    configViper.GetBool("auto_assign_agent"),
		EnableGit:          configViper.GetBool("enable_git"),
	}
}
