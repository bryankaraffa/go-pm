package pm

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FileSystem provides file system operations for the PM system.
// Implementations can use the OS file system or other storage backends.
type FileSystem interface {
	// CreateDirectory creates a directory and all necessary parents.
	// The directory permissions are set to 0755.
	CreateDirectory(path string) error

	// CopyFile copies a file from src to dst.
	// If dst already exists, it will be overwritten.
	CopyFile(src, dst string) error

	// WriteFile writes data to a file.
	// The file is created if it doesn't exist, and truncated if it does.
	WriteFile(path string, data []byte) error

	// ReadFile reads the contents of a file.
	ReadFile(path string) ([]byte, error)

	// FileExists checks if a file exists and is accessible.
	FileExists(path string) bool

	// DirectoryExists checks if a directory exists and is accessible.
	DirectoryExists(path string) bool

	// ListDirectories lists all directories in a path.
	ListDirectories(path string) ([]string, error)

	// ListFiles lists all files in a path.
	ListFiles(path string) ([]string, error)

	// MoveDirectory moves a directory from src to dst.
	// This is equivalent to renaming the directory.
	MoveDirectory(src, dst string) error
}

// OSFileSystem implements FileSystem using the OS file system
type OSFileSystem struct{}

// NewOSFileSystem creates a new OS file system instance.
// This provides access to the local file system for PM operations.
//
// Example:
//
//	fs := NewOSFileSystem()
//	err := fs.CreateDirectory("/tmp/pm-work")
//	if err != nil {
//		log.Fatal(err)
//	}
func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

// CreateDirectory creates a directory and all necessary parents.
// The directory permissions are set to 0755 (rwxr-xr-x).
func (fs *OSFileSystem) CreateDirectory(path string) error {
	return os.MkdirAll(path, 0o755)
}

// CopyFile copies a file from src to dst.
// If dst already exists, it will be overwritten. File permissions are set to 0644.
func (fs *OSFileSystem) CopyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

// WriteFile writes data to a file.
// The file is created if it doesn't exist, and truncated if it does.
// File permissions are set to 0644 (rw-r--r--).
func (fs *OSFileSystem) WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}

// ReadFile reads the contents of a file.
// Returns the file data as bytes.
func (fs *OSFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// FileExists checks if a file exists and is accessible.
// Returns false if the path is a directory or doesn't exist.
func (fs *OSFileSystem) FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// DirectoryExists checks if a directory exists and is accessible.
// Returns false if the path is a file or doesn't exist.
func (fs *OSFileSystem) DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// ListDirectories lists all directories in a path.
// Returns only directory names, not files. The path must exist and be readable.
func (fs *OSFileSystem) ListDirectories(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	return dirs, nil
}

// ListFiles lists all files in a path.
// Returns only file names, not directories. The path must exist and be readable.
func (fs *OSFileSystem) ListFiles(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

// MoveDirectory moves a directory from src to dst.
// This is equivalent to renaming the directory. Both src and dst must be on the same filesystem.
func (fs *OSFileSystem) MoveDirectory(src, dst string) error {
	return os.Rename(src, dst)
}

// TemplateProcessor handles template processing for work items.
// It copies template files and replaces placeholders with work item data.
type TemplateProcessor struct {
	fs     FileSystem
	config Config
}

// NewTemplateProcessor creates a new template processor.
// It requires a FileSystem implementation for file operations.
func NewTemplateProcessor(fs FileSystem, config Config) *TemplateProcessor {
	return &TemplateProcessor{fs: fs, config: config}
}

// ProcessTemplate processes an embedded template for a work item.
// It replaces {{name}} placeholders with the work item name.
// Templates are always sourced from embedded resources.
func (tp *TemplateProcessor) ProcessTemplate(targetPath, name string, itemType ItemType) error {
	// Get embedded template content
	var embeddedContent string
	switch itemType {
	case TypeFeature:
		embeddedContent = embeddedTemplateWorkItemFeature
	case TypeBug:
		embeddedContent = embeddedTemplateWorkItemBug
	case TypeExperiment:
		embeddedContent = embeddedTemplateWorkItemExperiment
	default:
		return fmt.Errorf("unsupported item type: %s", itemType)
	}

	// Process template placeholders
	processed := strings.ReplaceAll(embeddedContent, "{{name}}", name)

	// Write the processed content directly to target
	return tp.fs.WriteFile(targetPath, []byte(processed))
}

// WorkItemParser parses work item metadata from README files.
// It extracts status, phase, progress, and task information from markdown.
type WorkItemParser struct {
	fs FileSystem
}

// NewWorkItemParser creates a new work item parser.
// Requires a FileSystem implementation for file operations.
func NewWorkItemParser(fs FileSystem) *WorkItemParser {
	return &WorkItemParser{fs: fs}
}

// ParseWorkItem extracts metadata from a work item README file.
// It parses status, phase, progress, assignee, and tasks from the markdown content.
// Returns a WorkItem struct with all parsed information.
func (p *WorkItemParser) ParseWorkItem(name, path string) (WorkItem, error) {
	item := WorkItem{
		Name:   name,
		Path:   path,
		Status: "UNKNOWN",
		Phase:  PhaseDiscovery, // Default phase
	}

	content, err := p.fs.ReadFile(path)
	if err != nil {
		return item, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	var statusRegex = regexp.MustCompile(`##\s*Status:\s*(\w+(?:_\w+)*)`)
	var titleRegex = regexp.MustCompile(`^#\s+(?:Feature|Bug|Experiment):\s*(.+)$`)
	var phaseRegex = regexp.MustCompile(`##\s*Phase:\s*(\w+)`)
	var progressRegex = regexp.MustCompile(`##\s*Progress:\s*(\d+)%`)
	var assigneeRegex = regexp.MustCompile(`##\s*Assigned\s+To:\s*(.+)`)
	var phaseSectionRegex = regexp.MustCompile(`##\s+(\w+)\s+Phase`)
	var taskRegex = regexp.MustCompile(`^\s*-\s*\[([ x])\]\s*(.+)$`)

	currentPhase := PhaseDiscovery // Default to discovery

	for scanner.Scan() {
		line := scanner.Text()

		// Extract title from first heading
		if matches := titleRegex.FindStringSubmatch(line); len(matches) > 1 {
			item.Title = strings.TrimSpace(matches[1])
		}

		// Extract status
		if matches := statusRegex.FindStringSubmatch(line); len(matches) > 1 {
			item.Status = ItemStatus(strings.TrimSpace(matches[1]))
		}

		// Extract phase
		if matches := phaseRegex.FindStringSubmatch(line); len(matches) > 1 {
			item.Phase = WorkPhase(strings.TrimSpace(matches[1]))
		}

		// Extract progress
		if matches := progressRegex.FindStringSubmatch(line); len(matches) > 1 {
			if progress, err := strconv.Atoi(matches[1]); err == nil {
				item.Progress = progress
			}
		}

		// Extract assignee
		if matches := assigneeRegex.FindStringSubmatch(line); len(matches) > 1 {
			item.AssignedTo = strings.TrimSpace(matches[1])
		}

		// Check for phase section headers
		if matches := phaseSectionRegex.FindStringSubmatch(line); len(matches) > 1 {
			phaseName := strings.ToLower(matches[1])
			switch phaseName {
			case "discovery":
				currentPhase = PhaseDiscovery
			case "planning":
				currentPhase = PhasePlanning
			case "execution":
				currentPhase = PhaseExecution
			case "cleanup":
				currentPhase = PhaseCleanup
			}
		}

		// Extract tasks
		if matches := taskRegex.FindStringSubmatch(line); len(matches) > 1 {
			completed := matches[1] == "x"
			description := strings.TrimSpace(matches[2])
			task := Task{
				Description: description,
				Completed:   completed,
				Phase:       currentPhase,
				AssignedTo:  item.AssignedTo, // Default to work item assignee
			}
			item.Tasks = append(item.Tasks, task)
		}
	}

	if err := scanner.Err(); err != nil {
		return item, err
	}

	// Infer type from directory name
	if strings.HasPrefix(name, "feature-") {
		item.Type = TypeFeature
	} else if strings.HasPrefix(name, "bug-") {
		item.Type = TypeBug
	} else if strings.HasPrefix(name, "experiment-") {
		item.Type = TypeExperiment
	}

	// Set timestamps based on file information
	if fileInfo, err := os.Stat(path); err == nil {
		item.CreatedAt = fileInfo.ModTime() // Use file modification time as proxy for creation
		item.UpdatedAt = fileInfo.ModTime() // Use file modification time as last update
	}

	return item, nil
}

// StatusUpdater updates work item status in README files.
// It modifies the status, phase, progress, and assignee fields in markdown.
type StatusUpdater struct {
	fs FileSystem
}

// NewStatusUpdater creates a new status updater.
// Requires a FileSystem implementation for file operations.
func NewStatusUpdater(fs FileSystem) *StatusUpdater {
	return &StatusUpdater{fs: fs}
}

// UpdateStatus updates the status in a README file.
// It replaces the existing status line or adds one if none exists.
func (su *StatusUpdater) UpdateStatus(filePath string, newStatus ItemStatus) error {
	data, err := su.fs.ReadFile(filePath)
	if err != nil {
		return err
	}

	content := string(data)
	statusRegex := regexp.MustCompile(`(?i)(##\s*Status:\s*)(\w+)`)

	if statusRegex.MatchString(content) {
		content = statusRegex.ReplaceAllString(content, fmt.Sprintf("${1}%s", newStatus))
	} else {
		// If no status line found, add one after the first heading
		lines := strings.Split(content, "\n")
		if len(lines) > 0 && strings.HasPrefix(lines[0], "#") {
			lines = append(lines[:1], append([]string{fmt.Sprintf("\n## Status: %s", newStatus)}, lines[1:]...)...)
			content = strings.Join(lines, "\n")
		}
	}

	return su.fs.WriteFile(filePath, []byte(content))
}

// UpdateProgress updates the progress in a README file
func (su *StatusUpdater) UpdateProgress(filePath string, progress int) error {
	data, err := su.fs.ReadFile(filePath)
	if err != nil {
		return err
	}

	content := string(data)
	progressRegex := regexp.MustCompile(`(?i)(##\s*Progress:\s*)(\d+)%`)

	if progressRegex.MatchString(content) {
		content = progressRegex.ReplaceAllString(content, fmt.Sprintf("${1}%d%%", progress))
	} else {
		// If no progress line found, add one after status
		statusRegex := regexp.MustCompile(`(?i)(##\s*Status:\s*\w+)`)
		if statusRegex.MatchString(content) {
			content = statusRegex.ReplaceAllString(content, fmt.Sprintf("${1}\n\n## Progress: %d%%", progress))
		}
	}

	return su.fs.WriteFile(filePath, []byte(content))
}

// UpdateAssignee updates the assignee in a README file
func (su *StatusUpdater) UpdateAssignee(filePath string, assignee string) error {
	data, err := su.fs.ReadFile(filePath)
	if err != nil {
		return err
	}

	content := string(data)
	assigneeRegex := regexp.MustCompile(`(?i)(##\s*Assigned\s+To:\s*)(.+)`)
	phaseRegex := regexp.MustCompile(`(?i)(##\s*Phase:\s*\w+)`)

	if assigneeRegex.MatchString(content) {
		content = assigneeRegex.ReplaceAllString(content, fmt.Sprintf("${1}%s", assignee))
	} else {
		// If no assignee line found, add one after phase
		if phaseRegex.MatchString(content) {
			content = phaseRegex.ReplaceAllString(content, fmt.Sprintf("${1}\n\n## Assigned To: %s", assignee))
		}
	}

	return su.fs.WriteFile(filePath, []byte(content))
}

// UpdatePhaseAndStatus updates both phase and status in a README file
func (su *StatusUpdater) UpdatePhaseAndStatus(filePath string, phase WorkPhase, status ItemStatus) error {
	data, err := su.fs.ReadFile(filePath)
	if err != nil {
		return err
	}

	content := string(data)

	// Update phase
	phaseRegex := regexp.MustCompile(`(?i)(##\s*Phase:\s*)(\w+)`)
	if phaseRegex.MatchString(content) {
		content = phaseRegex.ReplaceAllString(content, fmt.Sprintf("${1}%s", phase))
	} else {
		// Add phase after title if not found
		titleRegex := regexp.MustCompile(`(^# .+\n)`)
		if titleRegex.MatchString(content) {
			content = titleRegex.ReplaceAllString(content, fmt.Sprintf("${1}\n## Phase: %s", phase))
		}
	}

	// Update status
	statusRegex := regexp.MustCompile(`(?i)(##\s*Status:\s*)(\w+(?:_\w+)*)`)
	if statusRegex.MatchString(content) {
		content = statusRegex.ReplaceAllString(content, fmt.Sprintf("${1}%s", status))
	} else {
		// Add status after phase if not found
		phaseRegex = regexp.MustCompile(`(?i)(##\s*Phase:\s*\w+)`)
		if phaseRegex.MatchString(content) {
			content = phaseRegex.ReplaceAllString(content, fmt.Sprintf("${1}\n\n## Status: %s", status))
		}
	}

	return su.fs.WriteFile(filePath, []byte(content))
}

// UpdatePhase updates the phase in a README file
func (su *StatusUpdater) UpdatePhase(filePath string, phase WorkPhase) error {
	data, err := su.fs.ReadFile(filePath)
	if err != nil {
		return err
	}

	content := string(data)
	phaseRegex := regexp.MustCompile(`(?i)(##\s*Phase:\s*)(\w+)`)

	if phaseRegex.MatchString(content) {
		content = phaseRegex.ReplaceAllString(content, fmt.Sprintf("${1}%s", phase))
	} else {
		// If no phase line found, add one after title
		titleRegex := regexp.MustCompile(`(^# .+\n)`)
		if titleRegex.MatchString(content) {
			content = titleRegex.ReplaceAllString(content, fmt.Sprintf("${1}\n## Phase: %s", phase))
		}
	}

	return su.fs.WriteFile(filePath, []byte(content))
}

// CompleteTask marks a task as completed in a README file
func (su *StatusUpdater) CompleteTask(filePath string, taskId int) error {
	data, err := su.fs.ReadFile(filePath)
	if err != nil {
		return err
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	taskRegex := regexp.MustCompile(`^\s*-\s*\[([ x])\]`)
	completeRegex := regexp.MustCompile(`^\s*-\s*\[\s*\]`)

	taskCount := 0
	for i, line := range lines {
		if taskRegex.MatchString(line) {
			if taskCount == taskId {
				// Mark this task as completed
				lines[i] = completeRegex.ReplaceAllString(line, "- [x]")
				break
			}
			taskCount++
		}
	}

	content = strings.Join(lines, "\n")
	return su.fs.WriteFile(filePath, []byte(content))
}

// TaskParser parses task completion status from README files.
// It counts completed and total tasks in markdown checklists.
type TaskParser struct {
	fs FileSystem
}

// NewTaskParser creates a new task parser.
// Requires a FileSystem implementation for file operations.
func NewTaskParser(fs FileSystem) *TaskParser {
	return &TaskParser{fs: fs}
}

// ParseTaskList counts total and completed tasks in a README.
// Returns the total number of tasks and the number that are completed.
func (tp *TaskParser) ParseTaskList(filePath string) (total, completed int, err error) {
	content, err := tp.fs.ReadFile(filePath)
	if err != nil {
		return 0, 0, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	taskRegex := regexp.MustCompile(`^\s*-\s*\[([ x])\]`)

	for scanner.Scan() {
		line := scanner.Text()
		if matches := taskRegex.FindStringSubmatch(line); len(matches) > 1 {
			total++
			if matches[1] == "x" {
				completed++
			}
		}
	}

	return total, completed, scanner.Err()
}

// PostmortemGenerator generates postmortem templates for completed work items.
// It creates structured templates for retrospective analysis.
type PostmortemGenerator struct {
	fs FileSystem
}

// NewPostmortemGenerator creates a new postmortem generator.
// Requires a FileSystem implementation for file operations.
func NewPostmortemGenerator(fs FileSystem) *PostmortemGenerator {
	return &PostmortemGenerator{fs: fs}
}

// GeneratePostmortem creates a postmortem template for a completed work item.
// It generates a structured markdown template for retrospective analysis.
func (pg *PostmortemGenerator) GeneratePostmortem(path, name string) error {
	template := fmt.Sprintf(`# Postmortem: %s

## Completion Date
%s

## Summary
- [ ] What was accomplished?
- [ ] Key challenges faced?
- [ ] Lessons learned?

## Metrics
- Development time:
- Lines of code added/modified:
- Tests added:

## What Went Well
-

## What Could Be Improved
-

## Follow-up Items
- [ ] Documentation updates needed
- [ ] Technical debt created
- [ ] Future enhancements identified
`, name, time.Now().Format("2006-01-02"))

	postmortemPath := filepath.Join(path, "POSTMORTEM.md")
	return pg.fs.WriteFile(postmortemPath, []byte(template))
}
