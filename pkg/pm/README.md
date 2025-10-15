# PM Package

The `pm` package provides project management functionality for tracking work items in documentation-driven development workflows.

## Overview

This package implements the core business logic for managing work items (features, bugs, experiments) in a structured way. It handles:

- Work item creation with templates
- Status tracking and updates
- Archiving completed work
- Git branch integration
- File system operations

## Key Interfaces

### Manager

The `Manager` interface defines the main operations:

```go
type Manager interface {
    CreateWorkItem(ctx context.Context, req CreateRequest) (*WorkItem, error)
    ListWorkItems(ctx context.Context, filter ListFilter) ([]WorkItem, error)
    GetWorkItem(ctx context.Context, name string) (*WorkItem, error)
    UpdateProgress(ctx context.Context, name string, progress int) error
    AssignWorkItem(ctx context.Context, name, assignee string) error
    AdvancePhase(ctx context.Context, name string) error
    Checkpoint(ctx context.Context, name, message string) error
    RequestReview(ctx context.Context, name string) error
    ApproveReview(ctx context.Context, name string) error
    GetPhaseTasks(ctx context.Context, name string) ([]Task, error)
    CompleteTask(ctx context.Context, name string, taskId int) error
    GetProgressMetrics(ctx context.Context, name string) (*WorkItemMetrics, error)
    ArchiveWorkItem(ctx context.Context, name string) error
}
```

### FileSystem

Abstracts file system operations for testability:

```go
type FileSystem interface {
    CreateDirectory(path string) error
    CopyFile(src, dst string) error
    WriteFile(path string, data []byte) error
    ReadFile(path string) ([]byte, error)
    FileExists(path string) bool
    DirectoryExists(path string) bool
    ListDirectories(path string) ([]string, error)
    MoveDirectory(src, dst string) error
}
```

### GitClient

Handles git operations:

```go
type GitClient interface {
    CreateBranch(branchName string) error
    BranchExists(branchName string) bool
    GetCurrentBranch() (string, error)
    GetGitUserName() (string, error)
}
```

## Usage

### Basic Usage

```go
config := pm.DefaultConfig()
manager := pm.NewDefaultManager(config)

req := pm.CreateRequest{
    Type: pm.TypeFeature,
    Name: "user-auth",
}

workItem, err := manager.CreateWorkItem(ctx, req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Created work item: %s\n", workItem.Name)
```

### With Custom Dependencies

```go
config := pm.DefaultConfig()
fs := pm.NewOSFileSystem()
gitClient := pm.NewOSGitClient()

manager := pm.NewDefaultManagerWithDeps(config, fs, gitClient)
```

## Work Item Lifecycle

Work items follow a structured **3-phase development process** with **4 status values**:

### Status Values
- **PLANNING**: Initial planning and requirements gathering
- **IMPLEMENTATION**: Active development using TDD practices
- **REVIEW**: Final validation and approval
- **COMPLETED**: Work finished and ready for archive

### Phases
- **planning**: Analyze requirements, design solution, create specifications
- **implementation**: Follow TDD (RED → GREEN → REFACTOR), commit frequently
- **review**: Final testing, documentation, and approval

### Workflow Transitions

1. **Create**: Work item starts in PLANNING status, planning phase
2. **Advance**: `AdvancePhase()` moves planning → implementation
3. **Checkpoint**: `Checkpoint()` saves progress during implementation (optional)
4. **Request Review**: `RequestReview()` moves implementation → review
5. **Approve Review**: `ApproveReview()` moves review → completed (100% progress)
6. **Archive**: `ArchiveWorkItem()` moves to completed directory

### TDD Workflow

During the implementation phase, follow this cycle:
1. **RED**: Write failing tests that validate desired functionality
2. **GREEN**: Write minimal code to make tests pass
3. **REFACTOR**: Improve code while keeping tests green
4. Repeat with frequent commits (every 30 minutes)

Use `Checkpoint()` to save progress without advancing phases.

## Directory Structure

```
{backlog_dir}/
├── feature-name/      # Active work items
│   └── README.md
{completed_dir}/
├── feature-name/      # Archived work items
│   ├── README.md
│   └── POSTMORTEM.md
```

The `backlog_dir` and `completed_dir` are configurable via the `Config` struct (defaults: "work-items/backlog" and "work-items/completed").

## Error Handling

The package uses custom error types for better error handling:

- `WorkItemError`: Operation-specific errors with context
- `ValidationError`: Input validation errors

```go
workItem, err := manager.CreateWorkItem(ctx, req)
if err != nil {
    var workItemErr *pm.WorkItemError
    if errors.As(err, &workItemErr) {
        fmt.Printf("Work item operation failed: %s on %s\n", workItemErr.Op, workItemErr.Name)
    }
    return err
}
```

## Testing

The package is designed for testability with dependency injection:

```go
func TestCreateWorkItem(t *testing.T) {
    fs := &mockFileSystem{}
    gitClient := &mockGitClient{}
    config := pm.DefaultConfig()

    manager := pm.NewDefaultManagerWithDeps(config, fs, gitClient)

    req := pm.CreateRequest{Type: pm.TypeFeature, Name: "test"}
    workItem, err := manager.CreateWorkItem(context.Background(), req)

    assert.NoError(t, err)
    assert.Equal(t, "feature-test", workItem.Name)
}
```

## CLI Integration

The package includes CLI helpers for easy integration:

```go
helper := pm.NewCLIHelper(manager)

// Create and report
err := helper.CreateAndReport(ctx, pm.TypeFeature, "auth")

// List and report
err := helper.ListAndReport(ctx, pm.StatusPlanning)

// Checkpoint progress
err := helper.CheckpointAndReport(ctx, "feature-auth", "Completed user validation")

// Request review
err := helper.RequestReviewAndReport(ctx, "feature-auth")

// Approve review
err := helper.ApproveReviewAndReport(ctx, "feature-auth")
```

## Future Enhancements

- Plugin system for custom work item types
- Enhanced reporting and analytics
