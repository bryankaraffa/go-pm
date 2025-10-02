package pm

import (
	"context"
	"fmt"
)

// DefaultManager is the default implementation of the Manager interface.
// It provides a complete project management system using the local filesystem
// and git integration.
type DefaultManager struct {
	service *WorkItemService
}

// NewCLIHelper creates a new CLI helper that provides formatted output
// and user-friendly interfaces for the Manager. While designed for CLI use,
// it can be used by any application that wants similar formatted output.
//
// Example:
//
//	manager := NewDefaultManager(config)
//	helper := NewCLIHelper(manager, config)
//	err := helper.CreateAndReport(ctx, TypeFeature, "user-auth")
func NewCLIHelper(manager Manager, config Config) *CLIHelper {
	return &CLIHelper{
		manager: manager,
		config:  config,
		fs:      NewOSFileSystem(),
	}
}

// NewDefaultManager creates a new default manager with standard dependencies.
// It uses the OS filesystem and git client for all operations.
//
// Example:
//
//	config := DefaultConfig()
//	manager := NewDefaultManager(config)
func NewDefaultManager(config Config) *DefaultManager {
	fs := NewOSFileSystem()
	gitClient := NewOSGitClient()

	return &DefaultManager{
		service: NewWorkItemService(config, fs, gitClient),
	}
}

// NewDefaultManagerWithDeps creates a new default manager with custom dependencies.
// This is primarily useful for testing or when custom filesystem/git implementations
// are needed.
//
// Example:
//
//	fs := NewMockFileSystem()
//	git := NewMockGitClient()
//	manager := NewDefaultManagerWithDeps(config, fs, git)
func NewDefaultManagerWithDeps(config Config, fs FileSystem, gitClient GitClient) *DefaultManager {
	return &DefaultManager{
		service: NewWorkItemService(config, fs, gitClient),
	}
}

// CreateWorkItem creates a new work item with the given parameters.
// It generates the directory structure, applies templates, creates a git branch,
// and returns the created work item. The work item starts in PROPOSED status
// in the discovery phase.
//
// Example:
//
//	config := DefaultConfig()
//	manager := NewDefaultManager(config)
//	req := CreateRequest{Type: TypeFeature, Name: "user-auth"}
//	item, err := manager.CreateWorkItem(ctx, req)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Created work item: %s\n", item.Name)
func (m *DefaultManager) CreateWorkItem(ctx context.Context, req CreateRequest) (*WorkItem, error) {
	return m.service.CreateWorkItem(ctx, req)
}

// ListWorkItems returns work items matching the filter criteria.
// Use an empty filter to return all work items.
//
// Example:
//
//	config := DefaultConfig()
//	manager := NewDefaultManager(config)
//	filter := ListFilter{Status: StatusInProgressExecution}
//	items, err := manager.ListWorkItems(ctx, filter)
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, item := range items {
//		fmt.Printf("%s: %s\n", item.Name, item.Status)
//	}
func (m *DefaultManager) ListWorkItems(ctx context.Context, filter ListFilter) ([]WorkItem, error) {
	return m.service.ListWorkItems(ctx, filter)
}

// GetWorkItem retrieves a specific work item by name.
// Returns an error if the work item doesn't exist.
//
// Example:
//
//	config := DefaultConfig()
//	manager := NewDefaultManager(config)
//	item, err := manager.GetWorkItem(ctx, "feature-user-auth")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Work item: %s (%s)\n", item.Title, item.Status)
func (m *DefaultManager) GetWorkItem(ctx context.Context, name string) (*WorkItem, error) {
	return m.service.GetWorkItem(ctx, name)
}

// UpdateStatus updates the status of a work item.
// This may trigger phase transitions or other workflow changes.
//
// Example:
//
//	config := DefaultConfig()
//	manager := NewDefaultManager(config)
//	err := manager.UpdateStatus(ctx, "feature-user-auth", StatusInProgressExecution)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println("Status updated successfully")
func (m *DefaultManager) UpdateStatus(ctx context.Context, name string, status ItemStatus) error {
	return m.service.UpdateStatus(ctx, name, status)
}

// UpdateProgress updates the progress of a work item.
// Progress is represented as a percentage (0-100).
//
// Example:
//
//	config := DefaultConfig()
//	manager := NewDefaultManager(config)
//	err := manager.UpdateProgress(ctx, "feature-user-auth", 75)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println("Progress updated to 75%")
func (m *DefaultManager) UpdateProgress(ctx context.Context, name string, progress int) error {
	return m.service.UpdateProgress(ctx, name, progress)
}

// AssignWorkItem assigns a work item to a user.
// The assignee field will be updated in the work item.
//
// Example:
//
//	config := DefaultConfig()
//	manager := NewDefaultManager(config)
//	err := manager.AssignWorkItem(ctx, "feature-user-auth", "john.doe@example.com")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println("Work item assigned to john.doe@example.com")
func (m *DefaultManager) AssignWorkItem(ctx context.Context, name string, assignee string) error {
	return m.service.AssignWorkItem(ctx, name, assignee)
}

// AdvancePhase advances a work item to the next phase in its workflow.
// This automatically updates the status and may create new tasks.
//
// Example:
//
//	config := DefaultConfig()
//	manager := NewDefaultManager(config)
//	err := manager.AdvancePhase(ctx, "feature-user-auth")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println("Work item advanced to next phase")
func (m *DefaultManager) AdvancePhase(ctx context.Context, name string) error {
	return m.service.AdvancePhase(ctx, name)
}

// SetPhase sets a work item to a specific phase.
// This may reset progress and create appropriate tasks for the phase.
//
// Example:
//
//	config := DefaultConfig()
//	manager := NewDefaultManager(config)
//	err := manager.SetPhase(ctx, "feature-user-auth", PhaseExecution)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println("Work item set to execution phase")
func (m *DefaultManager) SetPhase(ctx context.Context, name string, phase WorkPhase) error {
	return m.service.SetPhase(ctx, name, phase)
}

// GetPhaseTasks returns tasks for the current phase of a work item.
//
// Example:
//
//	config := DefaultConfig()
//	manager := NewDefaultManager(config)
//	tasks, err := manager.GetPhaseTasks(ctx, "feature-user-auth")
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, task := range tasks {
//		fmt.Printf("Task: %s (Completed: %v)\n", task.Description, task.Completed)
//	}
func (m *DefaultManager) GetPhaseTasks(ctx context.Context, name string) ([]Task, error) {
	return m.service.GetPhaseTasks(ctx, name)
}

// CompleteTask marks a task as completed.
// Task IDs can be obtained using GetPhaseTasks().
//
// Example:
//
//	config := DefaultConfig()
//	manager := NewDefaultManager(config)
//	// Get tasks first to find task ID
//	tasks, err := manager.GetPhaseTasks(ctx, "feature-user-auth")
//	if err != nil {
//		log.Fatal(err)
//	}
//	if len(tasks) > 0 {
//		err = manager.CompleteTask(ctx, "feature-user-auth", 0) // Complete first task
//		if err != nil {
//			log.Fatal(err)
//		}
//		fmt.Println("Task completed")
//	}
func (m *DefaultManager) CompleteTask(ctx context.Context, name string, taskId int) error {
	return m.service.CompleteTask(ctx, name, taskId)
}

// GetProgressMetrics returns progress metrics for a work item.
//
// Example:
//
//	config := DefaultConfig()
//	manager := NewDefaultManager(config)
//	metrics, err := manager.GetProgressMetrics(ctx, "feature-user-auth")
//	if err != nil {
//		log.Fatal(err)
//	}
//	// The returned WorkItemMetrics contains fields such as OverallProgress,
//	// TotalTasks, CompletedTasks and PhaseProgress. Use those fields to
//	// construct a user-facing report. For example:
//	fmt.Printf("Progress: %d%% (%d/%d tasks completed)\n", metrics.OverallProgress, metrics.CompletedTasks, metrics.TotalTasks)
//	for _, pp := range metrics.PhaseProgress {
//		fmt.Printf("  %s: %d%% (%d/%d tasks)\n", pp.Phase, pp.ProgressPercent, pp.CompletedTasks, pp.TotalTasks)
func (m *DefaultManager) GetProgressMetrics(ctx context.Context, name string) (*WorkItemMetrics, error) {
	return m.service.GetProgressMetrics(ctx, name)
}

// ArchiveWorkItem moves a completed work item to the completed directory.
//
// Example:
//
//	config := DefaultConfig()
//	manager := NewDefaultManager(config)
//	err := manager.ArchiveWorkItem(ctx, "feature-user-auth")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println("Work item archived")
func (m *DefaultManager) ArchiveWorkItem(ctx context.Context, name string) error {
	return m.service.ArchiveWorkItem(ctx, name)
}

type CLIHelper struct {
	manager Manager
	config  Config
	fs      FileSystem
}

// NewCLIHelper creates a new CLI helper that provides formatted output
// and user-friendly interfaces for the Manager. While designed for CLI use,
// it can be used by any application that wants similar formatted output.
//
// CreateAndReport creates a work item and reports the result with user-friendly output.
// This method handles the creation and provides formatted success/error messages.
// It prints to stdout with emojis and helpful next steps.
func (h *CLIHelper) CreateAndReport(ctx context.Context, itemType ItemType, name string) error {
	req := CreateRequest{
		Type: itemType,
		Name: name,
	}

	item, err := h.manager.CreateWorkItem(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create work item: %w", err)
	}

	fmt.Printf("✅ Work item created successfully!\n")
	fmt.Printf("📁 Directory: %s\n", item.Path)
	if item.Title != "" {
		fmt.Printf("📝 Title: %s\n", item.Title)
	}
	fmt.Printf("🌿 Branch: %s/%s\n", item.Type, item.Name)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("1. Edit %s with details\n", item.Path)
	fmt.Printf("2. Update status as work progresses\n")
	fmt.Printf("3. Reference this path in commit messages\n")

	return nil
}

// ListAndReport lists work items with the specified status and prints formatted output.
// It displays work items in a user-friendly format with emojis and titles.
// If no items are found, it prints an appropriate message.
func (h *CLIHelper) ListAndReport(ctx context.Context, status ItemStatus) error {
	filter := ListFilter{Status: status}

	items, err := h.manager.ListWorkItems(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to list work items: %w", err)
	}

	statusStr := string(status)
	if statusStr == "" {
		statusStr = "all"
	}

	fmt.Printf("Work items with status: %s\n", statusStr)

	if len(items) == 0 {
		fmt.Printf("  No %s items found\n", statusStr)
		return nil
	}

	for _, item := range items {
		fmt.Printf("  📋 %s\n", item.Name)
		if item.Title != "" {
			fmt.Printf("     %s\n", item.Title)
		}
	}

	return nil
}

// ListActiveAndReport lists all work items that are currently in progress.
// Active work items are those with status IN_PROGRESS_* and are grouped by phase.
// It provides a comprehensive view of ongoing work across all phases.
func (h *CLIHelper) ListActiveAndReport(ctx context.Context) error {
	filter := ListFilter{} // Empty filter gets all items

	items, err := h.manager.ListWorkItems(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to list work items: %w", err)
	}

	activeStatuses := []ItemStatus{
		StatusInProgressDiscovery,
		StatusInProgressPlanning,
		StatusInProgressExecution,
		StatusInProgressCleanup,
		StatusInProgressReview,
	}

	statusGroups := make(map[ItemStatus][]WorkItem)
	for _, item := range items {
		for _, activeStatus := range activeStatuses {
			if item.Status == activeStatus {
				statusGroups[item.Status] = append(statusGroups[item.Status], item)
				break
			}
		}
	}

	fmt.Println("Active work items:")

	hasActive := false
	for _, status := range activeStatuses {
		if items, exists := statusGroups[status]; exists && len(items) > 0 {
			hasActive = true
			fmt.Printf("\n%s:\n", status)
			for _, item := range items {
				fmt.Printf("  📋 %s", item.Name)
				if item.Title != "" {
					fmt.Printf(" - %s", item.Title)
				}
				fmt.Printf(" [%s]", item.Phase)
				if item.Progress > 0 {
					fmt.Printf(" (%d%%)", item.Progress)
				}
				fmt.Println()
			}
		}
	}

	if !hasActive {
		fmt.Println("  No active work items found")
	}

	return nil
}

// ListAllAndReport lists all work items grouped by status with formatted output.
// It displays work items organized by their current status (PROPOSED, IN_PROGRESS_*, COMPLETED)
// with progress indicators and titles.
func (h *CLIHelper) ListAllAndReport(ctx context.Context) error {
	filter := ListFilter{} // Empty filter gets all items

	items, err := h.manager.ListWorkItems(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to list work items: %w", err)
	}

	fmt.Println("All work items:")

	if len(items) == 0 {
		fmt.Println("  No work items found")
		return nil
	}

	statusGroups := make(map[ItemStatus][]WorkItem)
	for _, item := range items {
		statusGroups[item.Status] = append(statusGroups[item.Status], item)
	}

	statuses := []ItemStatus{StatusProposed, StatusInProgressDiscovery, StatusInProgressPlanning, StatusInProgressExecution, StatusInProgressCleanup, StatusInProgressReview, StatusCompleted}
	for _, status := range statuses {
		if items, exists := statusGroups[status]; exists && len(items) > 0 {
			fmt.Printf("\n%s:\n", status)
			for _, item := range items {
				fmt.Printf("  📋 %s", item.Name)
				if item.Title != "" {
					fmt.Printf(" - %s", item.Title)
				}
				fmt.Printf(" [%s]", item.Phase)
				if item.Progress > 0 {
					fmt.Printf(" (%d%%)", item.Progress)
				}
				fmt.Println()
			}
		}
	}

	return nil
}

// UpdateStatusAndReport updates work item status and reports the result.
// Status must be a valid ItemStatus constant.
// It prints success/error messages to stdout.
func (h *CLIHelper) UpdateStatusAndReport(ctx context.Context, name string, status ItemStatus) error {
	if err := h.manager.UpdateStatus(ctx, name, status); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	fmt.Printf("✅ Updated '%s' status to: %s\n", name, status)
	return nil
}

// ArchiveAndReport archives a work item and reports the result.
// This moves the work item to the completed directory and generates a postmortem template.
// It prints success/error messages and next steps to stdout.
func (h *CLIHelper) ArchiveAndReport(ctx context.Context, name string) error {
	if err := h.manager.ArchiveWorkItem(ctx, name); err != nil {
		return fmt.Errorf("failed to archive work item: %w", err)
	}

	fmt.Printf("✅ Archived '%s' to docs/completed/\n", name)
	fmt.Printf("📝 Consider filling out the postmortem\n")

	return nil
}

// ShowDetails shows detailed information about a work item.
// It displays the work item's metadata, status, progress, and current tasks.
func (h *CLIHelper) ShowDetails(ctx context.Context, name string) error {
	item, err := h.manager.GetWorkItem(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get work item: %w", err)
	}

	fmt.Printf("📋 Work Item: %s\n", item.Name)
	if item.Title != "" {
		fmt.Printf("📝 Title: %s\n", item.Title)
	}
	fmt.Printf("⏱️  Status: %s\n", item.Status)
	fmt.Printf("📂 Path: %s\n", item.Path)

	// Show task completion summary (this would need to be added to the service)
	// For now, we'll skip this as it requires additional implementation

	return nil
}

// AdvancePhaseAndReport advances work item phase and reports the result.
// This validates that all current phase tasks are completed before advancing.
// It prints success/error messages to stdout.
func (h *CLIHelper) AdvancePhaseAndReport(ctx context.Context, name string) error {
	if err := h.manager.AdvancePhase(ctx, name); err != nil {
		return fmt.Errorf("failed to advance phase: %w", err)
	}

	fmt.Printf("✅ Advanced '%s' to next phase\n", name)
	return nil
}

// UpdateProgressAndReport updates work item progress and reports the result.
// Progress should be an integer between 0 and 100.
// It prints success/error messages to stdout.
func (h *CLIHelper) UpdateProgressAndReport(ctx context.Context, name, progressStr string) error {
	// Parse progress percentage
	var progress int
	if _, err := fmt.Sscanf(progressStr, "%d", &progress); err != nil {
		return fmt.Errorf("invalid progress percentage: %s", progressStr)
	}

	if err := h.manager.UpdateProgress(ctx, name, progress); err != nil {
		return fmt.Errorf("failed to update progress: %w", err)
	}

	fmt.Printf("✅ Updated '%s' progress to %d%%\n", name, progress)
	return nil
}

// ShowProgressMetrics shows detailed progress metrics for a work item.
// It displays task completion statistics, phase progress, and timing information.
func (h *CLIHelper) ShowProgressMetrics(ctx context.Context, name string) error {
	metrics, err := h.manager.GetProgressMetrics(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get progress metrics: %w", err)
	}

	// Create a progress tracker to generate the report
	tracker := NewProgressTracker(NewOSFileSystem())
	report := tracker.GetProgressReport(*metrics)
	fmt.Print(report)

	return nil
}

// AssignAndReport assigns work item and reports the result.
// Assignee can be "human", "agent", or a specific user identifier.
// It prints success/error messages to stdout.
func (h *CLIHelper) AssignAndReport(ctx context.Context, name, assignee string) error {
	if err := h.manager.AssignWorkItem(ctx, name, assignee); err != nil {
		return fmt.Errorf("failed to assign work item: %w", err)
	}

	fmt.Printf("✅ Assigned '%s' to %s\n", name, assignee)
	return nil
}

// PrintInstructions prints comprehensive guidelines for project contributors and AI agents
func (h *CLIHelper) PrintInstructions(ctx context.Context) error {
	fmt.Print(GetInstructions(h.config))
	return nil
}

// SetPhaseAndReport sets work item phase and reports the result.
// This is an admin override that bypasses normal phase advancement rules.
// It prints success/error messages to stdout.
func (h *CLIHelper) SetPhaseAndReport(ctx context.Context, name string, phase WorkPhase) error {
	if err := h.manager.SetPhase(ctx, name, phase); err != nil {
		return fmt.Errorf("failed to set phase: %w", err)
	}

	fmt.Printf("✅ Set '%s' phase to: %s\n", name, phase)
	return nil
}

// ShowPhaseTasks shows tasks for the current phase.
// It displays all tasks in the work item's current phase with completion status.
func (h *CLIHelper) ShowPhaseTasks(ctx context.Context, name string) error {
	tasks, err := h.manager.GetPhaseTasks(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get phase tasks: %w", err)
	}

	if len(tasks) == 0 {
		fmt.Printf("No tasks found for current phase of '%s'\n", name)
		return nil
	}

	fmt.Printf("Tasks for '%s' current phase:\n", name)
	for i, task := range tasks {
		status := "[ ]"
		if task.Completed {
			status = "[x]"
		}
		fmt.Printf("  %d. %s %s", i, status, task.Description)
		if task.AssignedTo != "" {
			fmt.Printf(" (%s)", task.AssignedTo)
		}
		fmt.Println()
	}

	return nil
}

// CompleteTaskAndReport completes a task and reports the result.
// Task IDs can be obtained from ShowPhaseTasks output.
// It prints success/error messages to stdout.
func (h *CLIHelper) CompleteTaskAndReport(ctx context.Context, name string, taskId int) error {
	if err := h.manager.CompleteTask(ctx, name, taskId); err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}

	fmt.Printf("✅ Marked task %d as completed for '%s'\n", taskId, name)
	return nil
}
