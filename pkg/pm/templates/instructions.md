# Guidelines for Project Contributors and Maintainers

All development work must be tracked through the Project Management tool ([go-pm](https://github.com/bryankaraffa/go-pm)) to ensure:

- Clear project visibility and progress tracking
- Proper documentation and knowledge sharing
- Structured handoffs between human and AI agents
- Quality assurance through phased development
- Consistent workflow across all team members

**NEVER start coding or make changes without first creating a work item in the PM tool.**

## When to Use the PM Tool

You **MUST** use the PM tool for:

- Starting **ANY** new feature, bug fix, or experiment
- Making significant changes (>1 hour of work)
- Working on user-facing functionality
- Implementing new APIs or modifying existing ones
- Database schema changes
- Configuration or deployment changes
- Documentation updates that affect functionality

You do **NOT** need PM tracking for:
- Minor code cleanup (< 30 minutes)
- Simple refactoring without functional changes
- Documentation typo fixes
- Test-only changes

## Creating Work Items

Before starting any work:

1. Choose the appropriate work item type:
   - `feature` - New functionality or enhancements
   - `bug` - Bug fixes and defect resolution
   - `experiment` - Research, prototyping, or exploratory work

2. Create the work item:
   ```bash
   go-pm new <type> <name>
   ```
   Example: `go-pm new feature user-authentication`

3. The tool will automatically:
   - Create a directory structure under `wiki/work-items/backlog/`
   - Generate a README.md template with phase-specific sections
   - Set initial status to PROPOSED
   - Assign to you (the agent) by default

4. **IMMEDIATELY** after creation, advance to the first phase:
   ```bash
   go-pm phase advance <name>
   ```

## Phase Workflow

Work items progress through four mandatory phases:

### DISCOVERY Phase (Human-led, Agent-assisted)
**Goal:** Understand the problem space

**Activities:**
- Analyze requirements and constraints
- Research existing code and documentation
- Identify stakeholders and dependencies
- Document problem statement and success criteria

**Commands:**
- `go-pm status show <name>` - Check current status
- `go-pm progress update <name> <percentage>` - Update progress (0-25%)
- `go-pm phase advance <name>` - Move to PLANNING when ready

### PLANNING Phase (Agent-led, Human-validated)
**Goal:** Design the solution

**Activities:**
- Create technical design specifications
- Define API contracts and interfaces
- Break down work into tasks
- Identify testing requirements
- Update documentation with design decisions

**Commands:**
- `go-pm phase tasks <name>` - View current phase tasks
- `go-pm phase complete <name> <task-id>` - Mark tasks complete
- `go-pm progress update <name> <percentage>` - Update progress (26-50%)
- `go-pm phase advance <name>` - Move to EXECUTION when design approved

### EXECUTION Phase (Agent-led, Human-oversight)
**Goal:** Implement the solution

**Activities:**
- Write production code
- Create comprehensive tests
- Update documentation
- Code review and validation
- Integration testing

**Commands:**
- `go-pm progress update <name> <percentage>` - Update progress (51-90%)
- `go-pm phase advance <name>` - Move to CLEANUP when implementation complete

### CLEANUP Phase (Human-led, Agent-assisted)
**Goal:** Finalize and archive

**Activities:**
- Final testing and validation
- Documentation completion
- Postmortem analysis
- Knowledge sharing
- Archive the work item

**Commands:**
- `go-pm progress update <name> <percentage>` - Update progress (91-100%)
- `go-pm archive <name>` - Archive when fully complete

## Progress Tracking

Update progress regularly (at least daily):

- Use `go-pm progress update <name> <percentage>` to track completion
- Progress should reflect actual work completed, not time spent
- Always update progress before ending a work session
- Use meaningful increments (e.g., 10-20% per significant milestone)

Status transitions happen automatically with phase advancement. Manual status updates are rarely needed.

## Managing Work Items

### Viewing Work Items
- `go-pm list active` - See all work in progress
- `go-pm list proposed` - See items waiting to start
- `go-pm status show <name>` - Detailed view of specific item

### Assignment
- Work items are automatically assigned to the creating agent
- Use `go-pm assign <name> <assignee>` to reassign (rarely needed)
- Valid assignees: "human", "agent", or specific agent IDs

### Completion
When work is fully done:
1. Ensure all tasks are completed
2. Update progress to 100%
3. Advance through CLEANUP phase
4. Use `go-pm archive <name>` to archive

## Development Workflow Integration

### Before Starting Work
1. Check for existing work items: `go-pm list active`
2. If your task isn't tracked, create it: `go-pm new <type> <name>`
3. Advance to appropriate phase: `go-pm phase advance <name>`

### During Development
1. Work within the generated directory structure
2. Update documentation in the work item's README.md
3. Commit changes with meaningful messages
4. Update progress regularly: `go-pm progress update <name> <percentage>`

### Code Changes
- All code changes must relate to an active work item
- Update the work item's documentation with implementation details
- Include work item references in commit messages
- Run tests and validate before phase advancement

### Communication
- Use work item documentation for knowledge sharing
- Update README.md with decisions, challenges, and solutions
- Document API changes and breaking changes
- Include testing instructions and validation steps

## Best Practices and Rules

### Mandatory Rules
- **NEVER** start work without a PM work item
- **ALWAYS** update progress before ending work sessions
- **ALWAYS** advance phases when criteria are met
- **ALWAYS** document decisions and changes
- **ALWAYS** run tests before phase advancement

### Quality Assurance
- Complete all phase tasks before advancing
- Ensure documentation is current and accurate
- Test thoroughly before moving to CLEANUP
- Get human validation for design decisions

### Collaboration
- Use work items for clear handoffs
- Document context for human reviewers
- Update status to indicate when human input is needed
- Archive completed work promptly

### Efficiency
- Keep work items focused (single responsibility)
- Break large features into multiple work items
- Update progress meaningfully, not just daily
- Use templates provided by the tool

### Error Handling
- If blocked, document the issue in the work item
- Use appropriate status to indicate blocking conditions
- Communicate clearly with human collaborators
- Don't abandon work items - mark as blocked instead

## Common Scenarios

### Starting a New Feature
```bash
go-pm new feature user-login
go-pm phase advance user-login
# Work on discovery...
go-pm progress update user-login 25
go-pm phase advance user-login
# Continue through phases...
```

### Fixing a Bug
```bash
go-pm new bug null-pointer-crash
go-pm phase advance null-pointer-crash
# Investigate and fix...
go-pm progress update null-pointer-crash 100
go-pm phase advance null-pointer-crash
go-pm archive null-pointer-crash
```

### Working on Existing Items
```bash
go-pm list active
go-pm status show existing-feature
go-pm progress update existing-feature 60
# Continue working...
```

### Getting Help
- Use `go-pm instructions` anytime to review these guidelines
- Check `go-pm status show <name>` for current state
- Use `go-pm phase tasks <name>` to see what needs to be done

---

**Remember:** The PM tool is your project management system. Use it consistently to maintain project quality, visibility, and collaboration. All agents are expected to follow these guidelines without exception.