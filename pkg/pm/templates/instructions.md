---
applyTo: '**'
---

# Guidelines for Project Contributors and Maintainers

All development work must be tracked through the Project Management CLI tool ([go-pm](https://github.com/bryankaraffa/go-pm)) to ensure:

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
   - Create a directory structure under `{{backlog_dir}}/`
   - Generate a README.md template with phase-specific sections
   - Set initial status to PROPOSED
   - Assign to you (the agent) by default

4. **IMMEDIATELY** after creation, proceed to the **Discovery** phase by following this prompt:
   <prompt>
        Work with the user to build out a full description of the newly created feature.
        Ask questions to understand:
          • What the feature/bug/experiment should do
          • Who will use it
          • Technical requirements
          • Dependencies and integrations
          • Success criteria

        IMPORTANT: You are in a planning phase. Do NOT recommend creating new project code at this stage.
        Success looks like getting a really great README.md document at the 
        
        When you're done, update the feature file and ask the user to clarify their goals for the feature.
        Update the work item file as you gather more information.
   </prompt>

## Phase Workflow

Work items progress through three mandatory phases:

### Planning Phase (Agent-led, Human-validated)
**Goal:** Understand the problem and design the solution

**Activities:**
- Analyze requirements and constraints
- Research existing code and documentation
- Create technical design specifications
- Define API contracts and interfaces
- Break down work into tasks
- Identify testing requirements
- Document problem statement and success criteria

**Commands:**
- `go-pm status show <name>` - Check current status
- `go-pm progress update <name> <percentage>` - Update progress (0-25%)
- `go-pm phase advance <name>` - Move to IMPLEMENTATION when design approved

### Implementation Phase (Agent-led, Human-oversight)
**Goal:** Build the solution

**Activities:**
- Checkpoint frequently (every 30 minutes recommended)
- Use `go-pm checkpoint <name> <message>` to save progress
- Update work item documentation with implementation details

**Commands:**
- `go-pm checkpoint <name> <message>` - Save progress without advancing phase
- `go-pm progress update <name> <percentage>` - Update progress (26-90%)
- `go-pm review request <name>` - Move to REVIEW when implementation complete

### Review Phase (Human-led, Agent-assisted)
**Goal:** Finalize and validate the solution

**Activities:**
- Final testing and validation
- Code review and feedback incorporation
- Documentation completion
- Postmortem analysis
- Knowledge sharing
- Final review and approval

**Commands:**
- `go-pm status show <name>` - Check review status
- `go-pm review approve <name>` - Mark as COMPLETED (100% progress)
- `go-pm archive <name>` - Archive when in completed status

## Progress Tracking

Update progress regularly using checkpoints and meaningful increments:

- Use `go-pm checkpoint <name> <message>` to save progress without advancing phase
- Use `go-pm progress update <name> <percentage>` to track completion
- Progress should reflect actual work completed, not time spent
- Checkpoint frequently (every 30 minutes recommended) during implementation
- Always update progress before ending a work session

Status transitions happen automatically with phase advancement and review commands. Manual status updates are rarely needed.

## Managing Work Items

### Viewing Work Items
- `go-pm list active` - See all work in progress
- `go-pm list proposed` - See items waiting to start
- `go-pm status show <name>` - Detailed view of specific item

### Assignment
- Use `go-pm assign <name> <assignee>` to reassign
- Valid assignees: "human", "agent", or specific agent IDs

### Completion
When work is fully done:
1. Ensure all tasks are completed
2. Update progress to 100%
3. Advance through REVIEW phase
4. Use `go-pm archive <name>` to archive

## Development Workflow Integration

### Before Starting Work
1. Check for existing work items: `go-pm list active`
2. If your task isn't tracked, create it: `go-pm new <type> <name>`
3. Advance to appropriate phase: `go-pm phase advance <name>`

### During Development
1. Work within the generated directory structure
2. Checkpoint changes frequently (every 30 minutes recommended)
3. Use `go-pm checkpoint <name> <message>` to save progress
4. Update documentation in the work item's README.md
5. Update progress regularly: `go-pm progress update <name> <percentage>`

### Code Changes
- All code changes must relate to an active work item
- Update the work item's documentation with implementation details
- Include work item references in commit messages
- Run tests and validate before requesting review

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
- **ALWAYS** run tests before requesting review

### Quality Assurance
- Complete all phase tasks before advancing
- Ensure documentation is current and accurate
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
# Work on planning...
go-pm progress update user-login 25
go-pm phase advance user-login
# Implement using TDD: RED → GREEN → REFACTOR
go-pm checkpoint user-login "Completed user authentication logic"
go-pm progress update user-login 75
go-pm review request user-login
# Human reviews and approves
go-pm review approve user-login
go-pm archive user-login
```

### Fixing a Bug
```bash
go-pm new bug null-pointer-crash
go-pm phase advance null-pointer-crash
# Plan the fix...
go-pm progress update null-pointer-crash 25
go-pm phase advance null-pointer-crash
# Implement fix using TDD
go-pm checkpoint null-pointer-crash "Added null check"
go-pm progress update null-pointer-crash 75
go-pm review request null-pointer-crash
go-pm review approve null-pointer-crash
go-pm archive null-pointer-crash
```

### Working on Existing Items
```bash
go-pm list active
go-pm status show existing-feature
go-pm checkpoint existing-feature "Refactored database layer"
go-pm progress update existing-feature 60
# Continue working...
```

### Getting Help
- Use `go-pm instructions` anytime to review these guidelines
- Check `go-pm status show <name>` for current state
- Use `go-pm phase tasks <name>` to see what needs to be done

---

**Remember:** The PM tool is your project management system. Use it consistently to maintain project quality, visibility, and collaboration. All agents are expected to follow these guidelines without exception.