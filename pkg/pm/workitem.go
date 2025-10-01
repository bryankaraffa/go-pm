package pm

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WorkItemService provides operations for managing work items.
// It coordinates between filesystem operations, git integration, and business logic
// to provide a complete work item management system.
type WorkItemService struct {
	config     Config
	fs         FileSystem
	parser     *WorkItemParser
	updater    *StatusUpdater
	templater  *TemplateProcessor
	git        *GitIntegration
	postmortem *PostmortemGenerator
	progress   *ProgressTracker
}

// NewWorkItemService creates a new work item service with the given dependencies.
// This is the core service that implements the business logic for work item management.
//
// Example:
//
//	config := DefaultConfig()
//	fs := NewOSFileSystem()
//	git := NewOSGitClient()
//	service := NewWorkItemService(config, fs, git)
func NewWorkItemService(config Config, fs FileSystem, gitClient GitClient) *WorkItemService {
	return &WorkItemService{
		config:     config,
		fs:         fs,
		parser:     NewWorkItemParser(fs),
		updater:    NewStatusUpdater(fs),
		templater:  NewTemplateProcessor(fs, config),
		git:        NewGitIntegration(gitClient),
		postmortem: NewPostmortemGenerator(fs),
		progress:   NewProgressTracker(fs),
	}
}

// CreateWorkItem creates a new work item with the given parameters.
// It generates the directory structure, applies templates, creates a git branch,
// and returns the created work item. The work item starts in PROPOSED status
// in the discovery phase.
func (s *WorkItemService) CreateWorkItem(ctx context.Context, req CreateRequest) (*WorkItem, error) {
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	workDir := s.getWorkItemPath(req.Type, req.Name)
	readmePath := filepath.Join(workDir, "README.md")
	templatePath := s.getTemplatePath(req.Type)

	// Create directory
	if err := s.fs.CreateDirectory(workDir); err != nil {
		return nil, &WorkItemError{Op: "create", Name: req.Name, Err: fmt.Errorf("failed to create directory: %w", err)}
	}

	// Process template
	if err := s.templater.ProcessTemplate(templatePath, readmePath, req.Name, req.Type); err != nil {
		return nil, &WorkItemError{Op: "create", Name: req.Name, Err: fmt.Errorf("failed to process template: %w", err)}
	}

	// Create git branch
	if s.config.EnableGit {
		if err := s.git.CreateWorkItemBranch(req.Type, req.Name); err != nil {
			// Log but don't fail
			fmt.Printf("Warning: Git branch creation failed: %v\n", err)
		}
	}

	// Parse the created work item
	item, err := s.parser.ParseWorkItem(s.getWorkItemDirName(req.Type, req.Name), readmePath)
	if err != nil {
		return nil, &WorkItemError{Op: "create", Name: req.Name, Err: fmt.Errorf("failed to parse created work item: %w", err)}
	}

	return &item, nil
}

// ListWorkItems returns work items matching the filter criteria.
// It searches the backlog directory and applies the provided filter.
// If no filter is provided (empty ListFilter), all work items are returned.
//
// Example:
//
//	filter := ListFilter{Status: StatusProposed}
//	items, err := service.ListWorkItems(ctx, filter)
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, item := range items {
//		fmt.Printf("Found: %s (%s)\n", item.Name, item.Status)
//	}
func (s *WorkItemService) ListWorkItems(ctx context.Context, filter ListFilter) ([]WorkItem, error) {
	var items []WorkItem

	// List from backlog directory
	if s.fs.DirectoryExists(s.config.BacklogDir) {
		backlogItems, err := s.listWorkItemsInDir(s.config.BacklogDir)
		if err != nil {
			return nil, fmt.Errorf("failed to list backlog items: %w", err)
		}
		items = append(items, backlogItems...)
	}

	// Apply filters
	var filtered []WorkItem
	for _, item := range items {
		if s.matchesFilter(item, filter) {
			filtered = append(filtered, item)
		}
	}

	return filtered, nil
}

// GetWorkItem retrieves a specific work item by name from the backlog directory.
// It parses the work item's README.md file and returns the complete WorkItem struct.
// Returns an error if the work item doesn't exist or cannot be parsed.
//
// Example:
//
//	item, err := service.GetWorkItem(ctx, "feature-user-auth")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Work item: %s, Status: %s\n", item.Name, item.Status)
func (s *WorkItemService) GetWorkItem(ctx context.Context, name string) (*WorkItem, error) {
	readmePath := filepath.Join(s.config.BacklogDir, name, "README.md")

	if !s.fs.FileExists(readmePath) {
		return nil, &WorkItemError{Op: "get", Name: name, Err: fmt.Errorf("work item not found")}
	}

	item, err := s.parser.ParseWorkItem(name, readmePath)
	if err != nil {
		return nil, &WorkItemError{Op: "get", Name: name, Err: fmt.Errorf("failed to parse work item: %w", err)}
	}

	return &item, nil
}

// UpdateStatus updates the status of a work item in its README.md file.
// The status must be a valid ItemStatus constant. This operation updates
// the work item's metadata but does not perform phase transitions.
//
// Example:
//
//	err := service.UpdateStatus(ctx, "feature-user-auth", StatusInProgressPlanning)
//	if err != nil {
//		log.Fatal(err)
//	}
func (s *WorkItemService) UpdateStatus(ctx context.Context, name string, status ItemStatus) error {
	if err := s.validateStatus(status); err != nil {
		return err
	}

	readmePath := filepath.Join(s.config.BacklogDir, name, "README.md")
	if !s.fs.FileExists(readmePath) {
		return &WorkItemError{Op: "update", Name: name, Err: fmt.Errorf("work item not found")}
	}

	// Update status in file
	if err := s.updater.UpdateStatus(readmePath, status); err != nil {
		return &WorkItemError{Op: "update", Name: name, Err: fmt.Errorf("failed to update status: %w", err)}
	}

	// Move to appropriate directory based on status (future enhancement)
	// For now, items stay in backlog until archived

	return nil
}

// ArchiveWorkItem moves a completed work item to the completed directory.
// It creates a postmortem template and moves the entire work item directory
// from the backlog to the completed location. The work item should be in
// COMPLETED status before archiving.
//
// Example:
//
//	err := service.ArchiveWorkItem(ctx, "feature-user-auth")
//	if err != nil {
//		log.Fatal(err)
//	}
//	// Work item is now in completed/ directory with postmortem template
func (s *WorkItemService) ArchiveWorkItem(ctx context.Context, name string) error {
	source := filepath.Join(s.config.BacklogDir, name)
	dest := filepath.Join(s.config.CompletedDir, name)

	if !s.fs.DirectoryExists(source) {
		return &WorkItemError{Op: "archive", Name: name, Err: fmt.Errorf("work item not found in backlog")}
	}

	// Create completed directory if it doesn't exist
	if err := s.fs.CreateDirectory(s.config.CompletedDir); err != nil {
		return &WorkItemError{Op: "archive", Name: name, Err: fmt.Errorf("failed to create completed directory: %w", err)}
	}

	// Move directory
	if err := s.fs.MoveDirectory(source, dest); err != nil {
		return &WorkItemError{Op: "archive", Name: name, Err: fmt.Errorf("failed to move work item: %w", err)}
	}

	// Generate postmortem
	if err := s.postmortem.GeneratePostmortem(dest, name); err != nil {
		fmt.Printf("Warning: Could not create postmortem template: %v\n", err)
	}

	return nil
}

// SetPhase sets the phase of a work item to a specific value (admin override).
// This bypasses normal phase advancement rules and should be used with caution.
// The phase must be a valid WorkPhase constant.
//
// Example:
//
//	err := service.SetPhase(ctx, "feature-user-auth", PhaseExecution)
//	if err != nil {
//		log.Fatal(err)
//	}
//	// Work item phase is now set to execution regardless of current state
func (s *WorkItemService) SetPhase(ctx context.Context, name string, phase WorkPhase) error {
	if err := s.validatePhase(phase); err != nil {
		return err
	}

	readmePath := filepath.Join(s.config.BacklogDir, name, "README.md")
	if !s.fs.FileExists(readmePath) {
		return &WorkItemError{Op: "set_phase", Name: name, Err: fmt.Errorf("work item not found")}
	}

	// Update phase in file
	if err := s.updater.UpdatePhase(readmePath, phase); err != nil {
		return &WorkItemError{Op: "set_phase", Name: name, Err: fmt.Errorf("failed to update phase: %w", err)}
	}

	return nil
}

// GetPhaseTasks returns all tasks for the current phase of a work item.
// Tasks are parsed from the work item's README.md file and filtered by the
// work item's current phase. Returns an empty slice if no tasks are found.
//
// Example:
//
//	tasks, err := service.GetPhaseTasks(ctx, "feature-user-auth")
//	if err != nil {
//		log.Fatal(err)
//	}
//	for i, task := range tasks {
//		status := "[ ]"
//		if task.Completed {
//			status = "[x]"
//		}
//		fmt.Printf("%d. %s %s\n", i, status, task.Description)
//	}
func (s *WorkItemService) GetPhaseTasks(ctx context.Context, name string) ([]Task, error) {
	readmePath := filepath.Join(s.config.BacklogDir, name, "README.md")
	if !s.fs.FileExists(readmePath) {
		return nil, &WorkItemError{Op: "get_phase_tasks", Name: name, Err: fmt.Errorf("work item not found")}
	}

	// Get current work item to determine phase
	item, err := s.parser.ParseWorkItem(name, readmePath)
	if err != nil {
		return nil, &WorkItemError{Op: "get_phase_tasks", Name: name, Err: fmt.Errorf("failed to parse work item: %w", err)}
	}

	// Filter tasks by current phase
	var phaseTasks []Task
	for _, task := range item.Tasks {
		if task.Phase == item.Phase {
			phaseTasks = append(phaseTasks, task)
		}
	}

	return phaseTasks, nil
}

// GetProgressMetrics returns detailed progress metrics for a work item.
// This includes task completion statistics, phase progress, and overall
// completion percentage. The metrics are calculated from the work item's
// current state and task completion status.
//
// Example:
//
//	metrics, err := service.GetProgressMetrics(ctx, "feature-user-auth")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Progress: %d%% (%d/%d tasks completed)\n",
//		metrics.ProgressPercent, metrics.TasksCompleted, metrics.TasksTotal)
func (s *WorkItemService) GetProgressMetrics(ctx context.Context, name string) (*WorkItemMetrics, error) {
	readmePath := filepath.Join(s.config.BacklogDir, name, "README.md")
	if !s.fs.FileExists(readmePath) {
		return nil, &WorkItemError{Op: "get_progress_metrics", Name: name, Err: fmt.Errorf("work item not found")}
	}

	// Get current work item
	item, err := s.parser.ParseWorkItem(name, readmePath)
	if err != nil {
		return nil, &WorkItemError{Op: "get_progress_metrics", Name: name, Err: fmt.Errorf("failed to parse work item: %w", err)}
	}

	// Calculate metrics
	metrics := s.progress.CalculateWorkItemMetrics(&item)
	return &metrics, nil
}

// CompleteTask marks a specific task as completed in a work item.
// The taskId corresponds to the index of the task in the current phase's task list.
// Task IDs can be obtained using GetPhaseTasks(). This updates the work item's
// README.md file to mark the task as completed ([x] instead of [ ]).
//
// Example:
//
//	// Get tasks first to find the right task ID
//	tasks, err := service.GetPhaseTasks(ctx, "feature-user-auth")
//	if err != nil {
//		log.Fatal(err)
//	}
//	// Complete the first task (index 0)
//	err = service.CompleteTask(ctx, "feature-user-auth", 0)
//	if err != nil {
//		log.Fatal(err)
//	}
func (s *WorkItemService) CompleteTask(ctx context.Context, name string, taskId int) error {
	readmePath := filepath.Join(s.config.BacklogDir, name, "README.md")
	if !s.fs.FileExists(readmePath) {
		return &WorkItemError{Op: "complete_task", Name: name, Err: fmt.Errorf("work item not found")}
	}

	// Get current work item to find the task
	item, err := s.parser.ParseWorkItem(name, readmePath)
	if err != nil {
		return &WorkItemError{Op: "complete_task", Name: name, Err: fmt.Errorf("failed to parse work item: %w", err)}
	}

	// Filter tasks by current phase to get phase-specific tasks
	var phaseTasks []Task
	for _, task := range item.Tasks {
		if task.Phase == item.Phase {
			phaseTasks = append(phaseTasks, task)
		}
	}

	// Validate task ID against phase tasks
	if taskId < 0 || taskId >= len(phaseTasks) {
		return &ValidationError{Field: "taskId", Value: fmt.Sprintf("%d", taskId), Message: "invalid task ID for current phase"}
	}

	// Find the global index of the phase task
	globalTaskId := -1
	phaseTaskIndex := 0
	for i, task := range item.Tasks {
		if task.Phase == item.Phase {
			if phaseTaskIndex == taskId {
				globalTaskId = i
				break
			}
			phaseTaskIndex++
		}
	}

	if globalTaskId == -1 {
		return &ValidationError{Field: "taskId", Value: fmt.Sprintf("%d", taskId), Message: "could not find task"}
	}

	// Mark task as completed in file using global index
	if err := s.updater.CompleteTask(readmePath, globalTaskId); err != nil {
		return &WorkItemError{Op: "complete_task", Name: name, Err: fmt.Errorf("failed to complete task: %w", err)}
	}

	// Automatically recalculate and update progress
	if err := s.updateProgressFromTasks(readmePath); err != nil {
		// Log warning but don't fail the task completion
		fmt.Printf("Warning: Could not update progress: %v\n", err)
	}

	return nil
}

// UpdateProgress updates the overall progress percentage of a work item.
// Progress should be an integer between 0 and 100 representing completion percentage.
// This updates the work item's README.md file with the new progress value.
//
// Example:
//
//	err := service.UpdateProgress(ctx, "feature-user-auth", 75)
//	if err != nil {
//		log.Fatal(err)
//	}
//	// Work item now shows 75% progress
func (s *WorkItemService) UpdateProgress(ctx context.Context, name string, progress int) error {
	if progress < 0 || progress > 100 {
		return &ValidationError{Field: "progress", Value: fmt.Sprintf("%d", progress), Message: "progress must be between 0 and 100"}
	}

	readmePath := filepath.Join(s.config.BacklogDir, name, "README.md")
	if !s.fs.FileExists(readmePath) {
		return &WorkItemError{Op: "update_progress", Name: name, Err: fmt.Errorf("work item not found")}
	}

	// Update progress in file
	if err := s.updater.UpdateProgress(readmePath, progress); err != nil {
		return &WorkItemError{Op: "update_progress", Name: name, Err: fmt.Errorf("failed to update progress: %w", err)}
	}

	return nil
}

// AssignWorkItem assigns a work item to a specific assignee.
// The assignee can be "human", "agent", or a specific user identifier.
// This updates the work item's README.md file with the new assignee.
//
// Example:
//
//	err := service.AssignWorkItem(ctx, "feature-user-auth", "john.doe")
//	if err != nil {
//		log.Fatal(err)
//	}
//	// Work item is now assigned to john.doe
//
//	// Or assign to agent
//	err = service.AssignWorkItem(ctx, "feature-user-auth", "agent")
//	if err != nil {
//		log.Fatal(err)
//	}
func (s *WorkItemService) AssignWorkItem(ctx context.Context, name, assignee string) error {
	if assignee == "" {
		return &ValidationError{Field: "assignee", Value: assignee, Message: "assignee cannot be empty"}
	}

	readmePath := filepath.Join(s.config.BacklogDir, name, "README.md")
	if !s.fs.FileExists(readmePath) {
		return &WorkItemError{Op: "assign", Name: name, Err: fmt.Errorf("work item not found")}
	}

	// Update assignee in file
	if err := s.updater.UpdateAssignee(readmePath, assignee); err != nil {
		return &WorkItemError{Op: "assign", Name: name, Err: fmt.Errorf("failed to update assignee: %w", err)}
	}

	return nil
}

// AdvancePhase advances a work item to the next phase in the workflow.
// This operation validates that all tasks in the current phase are completed
// before allowing the transition. It updates both the phase and status in the
// work item's README.md file and may auto-assign agents for certain phases.
//
// The phase progression is:
//
//	PROPOSED → IN_PROGRESS_DISCOVERY (discovery phase)
//	IN_PROGRESS_DISCOVERY → IN_PROGRESS_PLANNING (planning phase)
//	IN_PROGRESS_PLANNING → IN_PROGRESS_EXECUTION (execution phase)
//	IN_PROGRESS_EXECUTION → IN_PROGRESS_CLEANUP (cleanup phase)
//	IN_PROGRESS_CLEANUP → IN_PROGRESS_REVIEW → COMPLETED
//
// Example:
//
//	err := service.AdvancePhase(ctx, "feature-user-auth")
//	if err != nil {
//		log.Fatal(err)
//	}
//	// Work item advances to next phase if all current tasks are completed
func (s *WorkItemService) AdvancePhase(ctx context.Context, name string) error {
	readmePath := filepath.Join(s.config.BacklogDir, name, "README.md")
	if !s.fs.FileExists(readmePath) {
		return &WorkItemError{Op: "advance_phase", Name: name, Err: fmt.Errorf("work item not found")}
	}

	// Get current work item to determine next phase
	item, err := s.parser.ParseWorkItem(name, readmePath)
	if err != nil {
		return &WorkItemError{Op: "advance_phase", Name: name, Err: fmt.Errorf("failed to parse work item: %w", err)}
	}

	// Validate that all tasks in current phase are completed
	if err := s.validatePhaseTasksCompleted(item); err != nil {
		return err
	}

	// Determine next phase and status
	nextPhase, nextStatus, err := s.getNextPhase(item.Phase, item.Status)
	if err != nil {
		return err
	}

	// Update phase and status in file
	if err := s.updater.UpdatePhaseAndStatus(readmePath, nextPhase, nextStatus); err != nil {
		return &WorkItemError{Op: "advance_phase", Name: name, Err: fmt.Errorf("failed to update phase: %w", err)}
	}

	// Auto-assign agent if advancing to execution phase and auto-assign is enabled
	if nextPhase == PhaseExecution && s.config.AutoAssignAgent {
		// Check current assignee
		currentItem, err := s.parser.ParseWorkItem(name, readmePath)
		if err != nil {
			// Log warning but don't fail phase advancement
			fmt.Printf("Warning: Could not parse work item for auto-assignment check: %v\n", err)
		} else if currentItem.AssignedTo == "" || currentItem.AssignedTo == "human" {
			// Only auto-assign if not assigned or assigned to human
			gitUser, err := s.git.client.GetGitUserName()
			if err != nil {
				// Log warning but don't fail phase advancement
				fmt.Printf("Warning: Could not get git user name for auto-assignment: %v\n", err)
			} else if gitUser != "" {
				if err := s.updater.UpdateAssignee(readmePath, gitUser); err != nil {
					// Log warning but don't fail phase advancement
					fmt.Printf("Warning: Could not auto-assign work item to %s: %v\n", gitUser, err)
				} else {
					fmt.Printf("Auto-assigned work item to %s\n", gitUser)
				}
			}
		}
	}

	// Create git branch for new phase if needed
	if err := s.git.CreateWorkItemBranchForPhase(item.Type, item.Name, nextPhase); err != nil {
		// Log but don't fail
		fmt.Printf("Warning: Git branch creation failed: %v\n", err)
	}

	return nil
}

// updateProgressFromTasks recalculates and updates progress based on task completion
func (s *WorkItemService) updateProgressFromTasks(readmePath string) error {
	// Get task completion counts
	parser := NewTaskParser(s.fs)
	total, completed, err := parser.ParseTaskList(readmePath)
	if err != nil {
		return fmt.Errorf("failed to parse task list: %w", err)
	}

	// Calculate progress percentage
	var progress int
	if total > 0 {
		progress = (completed * 100) / total
	}

	// Update progress in the file
	return s.updater.UpdateProgress(readmePath, progress)
}

// validatePhaseTasksCompleted checks that all tasks in the current phase are completed
func (s *WorkItemService) validatePhaseTasksCompleted(item WorkItem) error {
	// Only validate task completion when actively working in a phase (IN_PROGRESS statuses)
	// PROPOSED status allows advancing to start working without requiring task completion
	if item.Status == StatusProposed {
		return nil
	}

	// Filter tasks by current phase
	var phaseTasks []Task
	for _, task := range item.Tasks {
		if task.Phase == item.Phase {
			phaseTasks = append(phaseTasks, task)
		}
	}

	// Check if all phase tasks are completed
	for _, task := range phaseTasks {
		if !task.Completed {
			return &PhaseError{
				WorkItem:     item.Name,
				CurrentPhase: item.Phase,
				TargetPhase:  "",
				Reason:       fmt.Sprintf("task '%s' is not completed", task.Description),
			}
		}
	}

	return nil
}

// validateCreateRequest validates a create request
func (s *WorkItemService) validateCreateRequest(req CreateRequest) error {
	if req.Name == "" {
		return &ValidationError{Field: "name", Value: req.Name, Message: "name cannot be empty"}
	}

	if req.Type == "" {
		return &ValidationError{Field: "type", Value: string(req.Type), Message: "type cannot be empty"}
	}

	validTypes := map[ItemType]bool{
		TypeFeature:    true,
		TypeBug:        true,
		TypeExperiment: true,
	}

	if !validTypes[req.Type] {
		return &ValidationError{Field: "type", Value: string(req.Type), Message: "invalid work item type"}
	}

	// Check if work item already exists
	workDir := s.getWorkItemPath(req.Type, req.Name)
	if s.fs.DirectoryExists(workDir) {
		return &ValidationError{Field: "name", Value: req.Name, Message: "work item already exists"}
	}

	return nil
}

// validateStatus validates an item status
func (s *WorkItemService) validateStatus(status ItemStatus) error {
	validStatuses := map[ItemStatus]bool{
		StatusProposed:            true,
		StatusInProgressDiscovery: true,
		StatusInProgressPlanning:  true,
		StatusInProgressExecution: true,
		StatusInProgressCleanup:   true,
		StatusInProgressReview:    true,
		StatusCompleted:           true,
	}

	if !validStatuses[status] {
		return &ValidationError{Field: "status", Value: string(status), Message: "invalid status"}
	}

	return nil
}

// validatePhase validates a work phase
func (s *WorkItemService) validatePhase(phase WorkPhase) error {
	validPhases := map[WorkPhase]bool{
		PhaseDiscovery: true,
		PhasePlanning:  true,
		PhaseExecution: true,
		PhaseCleanup:   true,
	}

	if !validPhases[phase] {
		return &ValidationError{Field: "phase", Value: string(phase), Message: "invalid phase"}
	}

	return nil
}

// getWorkItemPath returns the full path for a work item
func (s *WorkItemService) getWorkItemPath(itemType ItemType, name string) string {
	dirName := s.getWorkItemDirName(itemType, name)
	return filepath.Join(s.config.BacklogDir, dirName)
}

// getWorkItemDirName returns the directory name for a work item
func (s *WorkItemService) getWorkItemDirName(itemType ItemType, name string) string {
	return fmt.Sprintf("%s-%s", itemType, name)
}

//go:embed templates/workitem-bug.md
var embeddedTemplateWorkItemBug string

//go:embed templates/workitem-experiment.md
var embeddedTemplateWorkItemExperiment string

//go:embed templates/workitem-feature.md
var embeddedTemplateWorkItemFeature string

// getTemplatePath returns the template path for a work item type
func (s *WorkItemService) getTemplatePath(itemType ItemType) string {
	templateName := "workitem-" + strings.ToLower(string(itemType)) + ".md"
	return filepath.Join(s.config.TemplatesDir, templateName)
}

// listWorkItemsInDir lists all work items in a directory
func (s *WorkItemService) listWorkItemsInDir(dir string) ([]WorkItem, error) {
	dirs, err := s.fs.ListDirectories(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []WorkItem{}, nil
		}
		return nil, err
	}

	var items []WorkItem
	for _, name := range dirs {
		readmePath := filepath.Join(dir, name, "README.md")
		if s.fs.FileExists(readmePath) {
			item, err := s.parser.ParseWorkItem(name, readmePath)
			if err != nil {
				fmt.Printf("Warning: Could not parse %s: %v\n", readmePath, err)
				continue
			}
			items = append(items, item)
		}
	}

	return items, nil
}

// matchesFilter checks if a work item matches the filter criteria
func (s *WorkItemService) matchesFilter(item WorkItem, filter ListFilter) bool {
	if filter.Status != "" && item.Status != filter.Status {
		return false
	}

	if filter.Type != "" && item.Type != filter.Type {
		return false
	}

	return true
}

// getNextPhase determines the next phase and status for a work item
func (s *WorkItemService) getNextPhase(currentPhase WorkPhase, currentStatus ItemStatus) (WorkPhase, ItemStatus, error) {
	switch currentStatus {
	case StatusProposed:
		return PhaseDiscovery, StatusInProgressDiscovery, nil
	case StatusInProgressDiscovery:
		return PhasePlanning, StatusInProgressPlanning, nil
	case StatusInProgressPlanning:
		return PhaseExecution, StatusInProgressExecution, nil
	case StatusInProgressExecution:
		return PhaseCleanup, StatusInProgressCleanup, nil
	case StatusInProgressCleanup:
		return PhaseCleanup, StatusInProgressReview, nil
	case StatusInProgressReview:
		return PhaseCleanup, StatusCompleted, nil
	default:
		return "", "", &PhaseError{
			WorkItem:     "",
			CurrentPhase: currentPhase,
			TargetPhase:  "",
			Reason:       "cannot advance from current status",
		}
	}
}
