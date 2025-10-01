# Go Project Management Tool

A CLI tool for managing work items (features, bugs, experiments) in documentation-driven development workflows. Supports phased development with collaboration between humans and AI agents.

## Documentation

- [Package Documentation](https://pkg.go.dev/github.com/bryankaraffa/go-pm/pkg/pm) - Complete API documentation
- [`go-pm` CLI Implementation](https://pkg.go.dev/github.com/bryankaraffa/go-pm/cmd/go-pm) - CLI command implementation details

## Installation

### As a CLI Tool

```bash
go install github.com/bryankaraffa/go-pm/cmd/go-pm@latest
```

### As a Library

```bash
go get github.com/bryankaraffa/go-pm/pkg/pm
```

## Quick Start

```bash
# Create a new feature
go-pm new feature user-authentication

# List all work items
go-pm list all

# Advance through phases
go-pm phase advance feature-user-authentication

# Check version information
go-pm version

# Get help
go-pm --help
```

## Library Usage

```go
package main

import (
    "context"
    "log"

    "github.com/bryankaraffa/go-pm/pkg/pm"
)

func main() {
    // Create manager with default config
    config := pm.DefaultConfig()
    manager := pm.NewDefaultManager(config)

    ctx := context.Background()

    // Create a work item
    req := pm.CreateRequest{
        Type: pm.TypeFeature,
        Name: "user-auth",
    }
    item, err := manager.CreateWorkItem(ctx, req)
    if err != nil {
        log.Fatal(err)
    }

    // List work items
    items, err := manager.ListWorkItems(ctx, pm.ListFilter{})
    if err != nil {
        log.Fatal(err)
    }

    for _, item := range items {
        log.Printf("Work item: %s (%s)", item.Name, item.Status)
    }
}
```

## Configuration

The tool supports configuration through config files and environment variables. Configuration files take precedence over defaults, and environment variables override both.

### Config Files

Create a `config.yaml`, `config.json`, or `config.toml` file in the current directory or your home directory. See `config.yaml.example` for all available options:

```yaml
# config.yaml
docs_dir: "docs"
templates_dir: "docs/templates"
backlog_dir: "docs/backlog"
completed_dir: "docs/completed"
phase_timeout_days: 7
auto_assign_agent: true
```

### Environment Variables

Environment variables with the `PM_` prefix override config file values:

| Variable | Description | Default |
|----------|-------------|---------|
| `PM_DOCS_DIR` | Base docs directory | `"wiki"` |
| `PM_TEMPLATES_DIR` | Templates directory | `"wiki/work-items/templates"` |
| `PM_BACKLOG_DIR` | Active work items directory | `"wiki/work-items/backlog"` |
| `PM_COMPLETED_DIR` | Completed work items directory | `"wiki/work-items/completed"` |
| `PM_PHASE_TIMEOUT_DAYS` | Days before phase timeout warning | `7` |
| `PM_AUTO_ASSIGN_AGENT` | Auto-assign agents to execution phase | `true` |
| `PM_ENABLE_GIT` | Enable git integration | `false` |

Example:
```bash
export PM_DOCS_DIR="/custom/docs"
export PM_BACKLOG_DIR="/custom/docs/work-items/backlog"
go-pm new feature my-feature
```

## CLI Commands

### Core Commands

- `go-pm new feature|bug|experiment <name>` - Create new work items
- `go-pm list proposed|active|completed|all` - List work items by status
- `go-pm status update <name> <status>` - Update work item status
- `go-pm status show <name>` - Show work item details
- `go-pm phase advance <name>` - Advance work item to next phase
- `go-pm phase set <name> <phase>` - Manually set phase (admin override)
- `go-pm phase tasks <name>` - Show current phase tasks
- `go-pm phase complete <name> <task-id>` - Mark task as completed
- `go-pm pm progress update <name> <percentage>` - Update progress percentage
- `go-pm progress show <name>` - Show detailed progress metrics
- `go-pm assign <name> <assignee>` - Assign work item to human/agent
- `go-pm archive <name>` - Archive completed work item
- `go-pm instructions` - Print comprehensive guidelines for contributors
- `go-pm version` - Show version information

### Workflow

1. **Create**: `go-pm new feature my-feature`
2. **Develop**: Edit the generated README.md with requirements
3. **Advance**: `go-pm phase advance my-feature` (requires completing phase tasks)
4. **Complete**: Mark tasks done with `go-pm phase complete my-feature <task-id>`
5. **Archive**: `go-pm archive my-feature` when finished

## Development

### Building

```bash
# Build the CLI
make build

# Run tests
make test

# Generate documentation
go doc ./pkg
```

### Versioning

This project follows [semantic versioning](https://semver.org/). Version information is available via:

```bash
go-pm version
```

For development builds, the version shows as "dev" with the current git SHA.

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0). See [LICENSE](./LICENSE) for details.

The AGPL-3.0 ensures that:
- Modified versions remain open source
- Network deployments make source code available to users
- Commercial use is allowed but modifications must be shared