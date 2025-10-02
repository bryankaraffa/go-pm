// Package main provides the CLI entry point for the Project Management tool
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bryankaraffa/go-pm/pkg/pm"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-pm",
	Short: "Project management CLI tool written in Go",
	Long:  "A CLI tool to manage features, bugs, experiments, support questions and project workflow.  Help maintain markdown files for project tracking and documentation-driven development.",
}

var enableGit bool
var autoDetectRepoRoot bool
var baseDir string

func init() {
	rootCmd.PersistentFlags().BoolVar(&enableGit, "enable-git", false, "Enable git integration")
	rootCmd.PersistentFlags().BoolVar(&autoDetectRepoRoot, "auto-detect-repo-root", true, "Auto-detect repository root directory")
	rootCmd.PersistentFlags().StringVar(&baseDir, "base-dir", "", "Base directory for operations")
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create new work items",
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List work items by status",
}

var phaseCmd = &cobra.Command{
	Use:   "phase",
	Short: "Manage work item phases",
}

var progressCmd = &cobra.Command{
	Use:   "progress",
	Short: "Track work item progress",
}

// createWorkItemCommand creates a cobra command for creating work items of a specific type
func createWorkItemCommand(itemType pm.ItemType, description string) *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("%s [name]", strings.ToLower(string(itemType))),
		Short: fmt.Sprintf("Create new %s", description),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			config := pm.DefaultConfig()
			manager := pm.NewDefaultManager(config)

			req := pm.CreateRequest{
				Type: itemType,
				Name: args[0],
			}

			item, err := manager.CreateWorkItem(ctx, req)
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
		},
	}
}

func main() {
	// Check for flags and set env vars
	for i, arg := range os.Args {
		if arg == "--enable-git" {
			_ = os.Setenv("PM_ENABLE_GIT", "true")
		}
		if arg == "--auto-detect-repo-root=false" {
			_ = os.Setenv("PM_AUTO_DETECT_REPO_ROOT", "false")
		}
		if arg == "--base-dir" && i+1 < len(os.Args) {
			_ = os.Setenv("PM_BASE_DIR", os.Args[i+1])
		}
	}

	ctx := context.Background()

	config := pm.DefaultConfig()
	manager := pm.NewDefaultManager(config)
	newCmd.AddCommand(createWorkItemCommand(pm.TypeFeature, "feature"))
	newCmd.AddCommand(createWorkItemCommand(pm.TypeBug, "bug report"))
	newCmd.AddCommand(createWorkItemCommand(pm.TypeExperiment, "experiment"))
	listCmd.AddCommand(&cobra.Command{
		Use:   "proposed",
		Short: "List proposed work items",
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := pm.ListFilter{Status: pm.StatusProposed}

			items, err := manager.ListWorkItems(ctx, filter)
			if err != nil {
				return fmt.Errorf("failed to list work items: %w", err)
			}

			fmt.Println("Proposed work items:")
			if len(items) == 0 {
				fmt.Println("  No proposed work items found")
				return nil
			}

			for _, item := range items {
				fmt.Printf("  📋 %s", item.Name)
				if item.Title != "" {
					fmt.Printf(" - %s", item.Title)
				}
				fmt.Println()
			}

			return nil
		},
	})

	listCmd.AddCommand(&cobra.Command{
		Use:   "active",
		Short: "List active work items (in progress)",
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := pm.ListFilter{} // Empty filter gets all items

			items, err := manager.ListWorkItems(ctx, filter)
			if err != nil {
				return fmt.Errorf("failed to list work items: %w", err)
			}

			activeStatuses := []pm.ItemStatus{
				pm.StatusInProgressDiscovery,
				pm.StatusInProgressPlanning,
				pm.StatusInProgressExecution,
				pm.StatusInProgressCleanup,
				pm.StatusInProgressReview,
			}

			statusGroups := make(map[pm.ItemStatus][]pm.WorkItem)
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
		},
	})

	listCmd.AddCommand(&cobra.Command{
		Use:   "completed",
		Short: "List completed work items",
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := pm.ListFilter{Status: pm.StatusCompleted}

			items, err := manager.ListWorkItems(ctx, filter)
			if err != nil {
				return fmt.Errorf("failed to list work items: %w", err)
			}

			fmt.Println("Completed work items:")
			if len(items) == 0 {
				fmt.Println("  No completed work items found")
				return nil
			}

			for _, item := range items {
				fmt.Printf("  📋 %s", item.Name)
				if item.Title != "" {
					fmt.Printf(" - %s", item.Title)
				}
				fmt.Println()
			}

			return nil
		},
	})

	listCmd.AddCommand(&cobra.Command{
		Use:   "all",
		Short: "List all work items with status",
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := pm.ListFilter{} // Empty filter gets all items

			items, err := manager.ListWorkItems(ctx, filter)
			if err != nil {
				return fmt.Errorf("failed to list work items: %w", err)
			}

			fmt.Println("All work items:")

			if len(items) == 0 {
				fmt.Println("  No work items found")
				return nil
			}

			statusGroups := make(map[pm.ItemStatus][]pm.WorkItem)
			for _, item := range items {
				statusGroups[item.Status] = append(statusGroups[item.Status], item)
			}

			statuses := []pm.ItemStatus{pm.StatusProposed, pm.StatusInProgressDiscovery, pm.StatusInProgressPlanning, pm.StatusInProgressExecution, pm.StatusInProgressCleanup, pm.StatusInProgressReview, pm.StatusCompleted}
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
		},
	})

	// Archive command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "archive [name]",
		Short: "Archive completed work item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := manager.ArchiveWorkItem(ctx, args[0]); err != nil {
				return fmt.Errorf("failed to archive work item: %w", err)
			}

			fmt.Printf("✅ Archived '%s' to docs/completed/\n", args[0])
			fmt.Printf("📝 Consider filling out the postmortem\n")

			return nil
		},
	})

	// Status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Manage work item status",
	}

	statusCmd.AddCommand(&cobra.Command{
		Use:   "update [name] [status]",
		Short: "Update work item status",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var status pm.ItemStatus
			switch strings.ToLower(args[1]) {
			case "proposed":
				status = pm.StatusProposed
			case "in_progress_discovery", "discovery":
				status = pm.StatusInProgressDiscovery
			case "in_progress_planning", "planning":
				status = pm.StatusInProgressPlanning
			case "in_progress_execution", "execution":
				status = pm.StatusInProgressExecution
			case "in_progress_cleanup", "cleanup":
				status = pm.StatusInProgressCleanup
			case "in_progress_review", "review":
				status = pm.StatusInProgressReview
			case "completed":
				status = pm.StatusCompleted
			default:
				return fmt.Errorf("invalid status: %s. Valid statuses: proposed, discovery, planning, execution, cleanup, review, completed", args[1])
			}
			if err := manager.UpdateStatus(ctx, args[0], status); err != nil {
				return fmt.Errorf("failed to update status: %w", err)
			}

			fmt.Printf("✅ Updated '%s' status to: %s\n", args[0], status)
			return nil
		},
	})

	statusCmd.AddCommand(&cobra.Command{
		Use:   "show [name]",
		Short: "Show work item details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			item, err := manager.GetWorkItem(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get work item: %w", err)
			}

			fmt.Printf("📋 Work Item: %s\n", item.Name)
			if item.Title != "" {
				fmt.Printf("📝 Title: %s\n", item.Title)
			}
			fmt.Printf("⏱️  Status: %s\n", item.Status)
			fmt.Printf("� Phase: %s\n", item.Phase)
			if item.Progress > 0 {
				fmt.Printf("📈 Progress: %d%%\n", item.Progress)
			}
			if item.AssignedTo != "" {
				fmt.Printf("👤 Assigned To: %s\n", item.AssignedTo)
			}
			fmt.Printf("�📂 Path: %s\n", item.Path)
			fmt.Printf("📅 Created: %s\n", item.CreatedAt.Format("2006-01-02 15:04"))
			fmt.Printf("🔄 Updated: %s\n", item.UpdatedAt.Format("2006-01-02 15:04"))

			return nil
		},
	})

	rootCmd.AddCommand(statusCmd)

	// Phase commands
	phaseCmd.AddCommand(&cobra.Command{
		Use:   "advance [name]",
		Short: "Advance work item to next phase",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := manager.AdvancePhase(ctx, args[0]); err != nil {
				return fmt.Errorf("failed to advance phase: %w", err)
			}

			fmt.Printf("✅ Advanced '%s' to next phase\n", args[0])
			return nil
		},
	})

	phaseCmd.AddCommand(&cobra.Command{
		Use:   "set [name] [phase]",
		Short: "Set work item phase (admin override)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var phase pm.WorkPhase
			switch strings.ToLower(args[1]) {
			case "discovery":
				phase = pm.PhaseDiscovery
			case "planning":
				phase = pm.PhasePlanning
			case "execution":
				phase = pm.PhaseExecution
			case "cleanup":
				phase = pm.PhaseCleanup
			default:
				return fmt.Errorf("invalid phase: %s. Valid phases: discovery, planning, execution, cleanup", args[1])
			}
			if err := manager.SetPhase(ctx, args[0], phase); err != nil {
				return fmt.Errorf("failed to set phase: %w", err)
			}

			fmt.Printf("✅ Set '%s' phase to: %s\n", args[0], phase)
			return nil
		},
	})

	phaseCmd.AddCommand(&cobra.Command{
		Use:   "tasks [name]",
		Short: "Show current phase tasks",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tasks, err := manager.GetPhaseTasks(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get phase tasks: %w", err)
			}

			if len(tasks) == 0 {
				fmt.Printf("No tasks found for current phase of '%s'\n", args[0])
				return nil
			}

			fmt.Printf("Tasks for '%s' current phase:\n", args[0])
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
		},
	})

	phaseCmd.AddCommand(&cobra.Command{
		Use:   "complete [name] [task-id]",
		Short: "Mark task as completed",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskId, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid task ID: %s", args[1])
			}
			if err := manager.CompleteTask(ctx, args[0], taskId); err != nil {
				return fmt.Errorf("failed to complete task: %w", err)
			}

			fmt.Printf("✅ Marked task %d as completed for '%s'\n", taskId, args[0])
			return nil
		},
	})

	// Progress commands
	progressCmd.AddCommand(&cobra.Command{
		Use:   "update [name] [percentage]",
		Short: "Update work item progress percentage",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse progress percentage
			var progress int
			if _, err := fmt.Sscanf(args[1], "%d", &progress); err != nil {
				return fmt.Errorf("invalid progress percentage: %s", args[1])
			}

			if err := manager.UpdateProgress(ctx, args[0], progress); err != nil {
				return fmt.Errorf("failed to update progress: %w", err)
			}

			fmt.Printf("✅ Updated '%s' progress to %d%%\n", args[0], progress)
			return nil
		},
	})

	progressCmd.AddCommand(&cobra.Command{
		Use:   "show [name]",
		Short: "Show detailed progress metrics for a work item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			metrics, err := manager.GetProgressMetrics(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get progress metrics: %w", err)
			}

			// Create a progress tracker to generate the report
			tracker := pm.NewProgressTracker(pm.NewOSFileSystem())
			report := tracker.GetProgressReport(*metrics)
			fmt.Print(report)

			return nil
		},
	})

	// Assign commands
	rootCmd.AddCommand(&cobra.Command{
		Use:   "assign [name] [assignee]",
		Short: "Assign work item to human/agent",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := manager.AssignWorkItem(ctx, args[0], args[1]); err != nil {
				return fmt.Errorf("failed to assign work item: %w", err)
			}

			fmt.Printf("✅ Assigned '%s' to %s\n", args[0], args[1])
			return nil
		},
	}) // Instructions command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "instructions",
		Short: "Print comprehensive guidelines for project contributors and AI agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := pm.DefaultConfig()
			instructions := pm.GetInstructions(config)
			fmt.Print(instructions)
			return nil
		},
	})

	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(phaseCmd)
	rootCmd.AddCommand(progressCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
