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

## Configuration

The tool supports configuration through config files and environment variables. Configuration files take precedence over defaults, and environment variables override both.

### CLI flags

The CLI also accepts a couple of persistent flags which the `main` program maps to environment variables for convenience:

- `--enable-git` — enable git integration for branch creation and related operations (sets `PM_ENABLE_GIT=true` when passed).
- `--auto-detect-repo-root` / `--auto-detect-repo-root=false` — control whether the repository root is auto-detected (this maps to `PM_AUTO_DETECT_REPO_ROOT`).

When a CLI flag is provided, the program sets the corresponding `PM_` environment variable at startup; environment variables continue to take precedence over config file values.

### Config Files

Create a `config.yaml`, `config.json`, or `config.toml` file in the current directory or your home directory. See `config.yaml.example` for all available options:

```yaml
# config.yaml
auto_detect_repo_root: true
backlog_dir: "work-items/backlog"
completed_dir: "work-items/completed"
phase_timeout_days: 7
enable_git: false
```

### Environment Variables

Environment variables with the `PM_` prefix override config file values:

| Variable | Description | Default |
|----------|-------------|---------|
| `PM_AUTO_DETECT_REPO_ROOT` | Auto-detect repository root | `true` |
| `PM_BACKLOG_DIR` | Active work items directory (relative to repository root by default) | `"work-items/backlog"` |
| `PM_COMPLETED_DIR` | Completed work items directory (relative to repository root by default) | `"work-items/completed"` |
| `PM_PHASE_TIMEOUT_DAYS` | Days before phase timeout warning | `7` |
| `PM_ENABLE_GIT` | Enable git integration | `false` |

Example:
```bash
export PM_BACKLOG_DIR="work-items/backlog"
export PM_COMPLETED_DIR="work-items/completed"
go-pm new feature my-feature
```

## CLI Commands

### Core Commands

- `go-pm new feature|bug|experiment <name>` - Create new work items
- `go-pm list proposed|active|completed|all` - List work items by status
- `go-pm status show <name>` - Show work item details
- `go-pm status update <name> <status>` - Update work item status (proposed, discovery, planning, execution, cleanup, review, completed)
- `go-pm phase advance <name>` - Advance work item to next phase
- `go-pm phase set <name> <phase>` - Manually set phase (admin override) (discovery, planning, execution, cleanup)
- `go-pm phase tasks <name>` - Show current phase tasks
- `go-pm phase complete <name> <task-id>` - Mark task as completed
- `go-pm progress update <name> <percentage>` - Update progress percentage
- `go-pm progress show <name>` - Show detailed progress metrics
- `go-pm assign <name> <assignee>` - Assign work item to human/agent
- `go-pm archive <name>` - Archive completed work item
- `go-pm instructions` - Print comprehensive guidelines for contributors
- `go-pm version` - Show version information

### Workflow

1. **Create**: `go-pm new feature my-feature`
2. **Develop**: Edit the generated README.md with requirements
3. **Advance**: `go-pm phase advance my-feature` through phases (requires completing phase tasks)
4. **Complete**: Mark tasks done with `go-pm phase complete my-feature <task-id>`
5. **Finish**: In cleanup phase, advance twice: once to review status, then to completed
6. **Archive**: `go-pm archive my-feature` when in completed status

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