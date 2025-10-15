# Go Project Management Tool

A CLI tool for managing work items (features, bugs, experiments) in documentation-driven development workflows. Supports phased development with collaboration between humans and AI agents.

## Documentation

- [`pm` Package](https://pkg.go.dev/github.com/bryankaraffa/go-pm/pkg/pm) - Complete API documentation
- [`go-pm` CLI](https://pkg.go.dev/github.com/bryankaraffa/go-pm/cmd/go-pm) - CLI command implementation details

## CLI Installation

```bash
go install github.com/bryankaraffa/go-pm/cmd/go-pm@latest
```

## CLI Usage

```bash
# Install go-pm cli
go install github.com/bryankaraffa/go-pm/cmd/go-pm@latest

# Create a new feature
go-pm new feature user-authentication
## Immediately edit the generated README.md with requirements

# List all work items
go-pm list all

# Advance through phases
go-pm phase advance feature-user-authentication

# Check version information
go-pm version

# Get help
go-pm --help
```

## Configuration

go-pm can be configured via environment variables or a `config.yaml` file in the current directory.

### Environment Variables

- `PM_AUTO_DETECT_REPO_ROOT` (default: true) - Automatically detect the git repository root
- `PM_BACKLOG_DIR` (default: "work-items/backlog") - Directory for active work items
- `PM_COMPLETED_DIR` (default: "work-items/completed") - Directory for completed work items
- `PM_TEMPLATES_DIR` (default: "work-items/templates") - Directory for external templates
- `PM_ENABLE_EXTERNAL_TEMPLATES` (default: true) - Use external templates if available
- `PM_PHASE_TIMEOUT_DAYS` (default: 7) - Days before phase timeout warning
- `PM_ENABLE_GIT` (default: false) - Enable git integration

When `PM_AUTO_DETECT_REPO_ROOT` is enabled (default), all directory paths are treated as relative to the detected git repository root.

### Config File

Create a `config.yaml` file in your project root:

```yaml
auto_detect_repo_root: true
backlog_dir: "work-items/backlog"
completed_dir: "work-items/completed"
templates_dir: "work-items/templates"
enable_external_templates: true
phase_timeout_days: 7
enable_git: false
```

## Templates

go-pm supports customizable templates for work items. Templates are Markdown files that define the structure and content of work item README files.

### Template Files

Templates are stored in the `work-items/templates/` directory (configurable via `PM_TEMPLATES_DIR`):

- `workitem-feature.md` - Template for feature work items
- `workitem-bug.md` - Template for bug work items  
- `workitem-experiment.md` - Template for experiment work items

### Template Placeholders

Templates support the following placeholders that are automatically replaced when creating work items:

- `{{name}}` - The work item name (e.g., "user-authentication")

Example template content:
```markdown
# Feature: {{name}}

## Status: PROPOSED
## Phase: discovery

## Overview
Brief description of the feature.
```

### External Templates

By default, go-pm uses external templates if available. If a template file doesn't exist in the templates directory, it automatically creates one from the built-in template. You can then customize these external templates for your project.

To disable external templates and always use built-in templates:
```bash
export PM_ENABLE_EXTERNAL_TEMPLATES=false
```

## Agent / Assistant Usage

To provide project-specific instructions to your agent or LLM, run:

```sh
go-pm instructions
```

- Use the output as your agent’s system prompt, guidelines, or context.
- For IDEs or Copilot, you can save the output to a file and reference it:

    ```sh
    mkdir -p .github/instructions
    go-pm instructions > .github/instructions/go-pm.instructions.md
    ```

- To keep documentation up to date, always re-run `go-pm instructions` after updating your workflow or templates.

## Library Usage

First add the package to your module:

```bash
go get github.com/bryankaraffa/go-pm/pkg/pm
```

Then use it in your Go code:
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

## CLI Commands

### Core Commands

- `go-pm new feature|bug|experiment <name>` - Create new work items
- `go-pm list proposed|active|completed|all` - List work items by status
- `go-pm status show <name>` - Show work item details
- `go-pm progress update <name> <percentage>` - Update progress percentage
- `go-pm progress show <name>` - Show detailed progress metrics
- `go-pm assign <name> <assignee>` - Assign work item to human/agent
- `go-pm phase advance <name>` - Advance work item to next phase
- `go-pm checkpoint <name> <message>` - Save progress during implementation
- `go-pm review request <name>` - Request review (IMPLEMENTATION → REVIEW)
- `go-pm review approve <name>` - Approve review (REVIEW → COMPLETED)
- `go-pm archive <name>` - Archive completed work item
- `go-pm instructions` - Print comprehensive guidelines for contributors
- `go-pm version` - Show version information

### Workflow

The tool follows a **3-phase development workflow** with **TDD practices**:

1. **Planning Phase**: Analyze requirements, design solution, create technical specifications
2. **Implementation Phase**: Follow TDD (RED → GREEN → REFACTOR), commit frequently (every 30 minutes)
3. **Review Phase**: Final validation, documentation, and approval

#### Basic Workflow

```bash
# Create work item (starts in PLANNING phase)
go-pm new feature user-authentication

# Advance to implementation
go-pm phase advance user-authentication

# During implementation: save progress checkpoints
go-pm checkpoint user-authentication "Completed user model"

# Request review when implementation is complete
go-pm review request user-authentication

# Approve review (moves to COMPLETED)
go-pm review approve user-authentication

# Archive when done
go-pm archive user-authentication
```

#### TDD Workflow During Implementation

```bash
# RED: Write failing test
# (edit test file)

# GREEN: Write minimal code to pass test
# (edit implementation)

# REFACTOR: Improve code while keeping tests green
go-pm checkpoint feature-name "Refactored authentication logic"

# Repeat cycle, commit every 30 minutes
```

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