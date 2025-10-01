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
    UpdateStatus(ctx context.Context, name string, status ItemStatus) error
    ArchiveWorkItem(ctx context.Context, name string) error
    ...
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

1. **PROPOSED**: Initial state, documentation created
2. **IN_PROGRESS**: Implementation started
3. **COMPLETED**: Implementation finished, ready for archive

## Directory Structure

```
docs/
├── backlog/           # Active work items
│   └── feature-name/
│       └── README.md
├── completed/         # Archived work items
│   └── feature-name/
│       ├── README.md
│       └── POSTMORTEM.md
```

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
err := helper.ListAndReport(ctx, pm.StatusProposed)

// Update and report
err := helper.UpdateStatusAndReport(ctx, "feature-auth", pm.StatusInProgress)
```

## Future Enhancements

- Configuration file support
- Progress tracking with time estimates
- Team collaboration features
- REST API endpoints
