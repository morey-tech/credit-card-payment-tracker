# Wingspan Scoring - Development Guidelines

## Git Workflow

- **Always create a new feature branch before making any code changes**
- Never commit directly to `main` branch
- Branch naming: use descriptive names (e.g., `feature/score-calculator`, `fix/round-goals-bug`)

## GitHub CLI Workflow

The GitHub CLI (`gh`) is available in the devcontainer for streamlined PR management.

### Authentication

First-time setup requires authentication:
```bash
gh auth login
```

Follow the interactive prompts to authenticate via browser or token.

### Creating Pull Requests

After committing your changes to a feature branch, create a PR:

```bash
# Basic PR creation (opens editor for title/description)
gh pr create

# Create PR with inline title and body
gh pr create --title "feat: add score validation" --body "Implements input validation for bird scores"

# Create PR and open in browser
gh pr create --web
```

### Auto-Merge with Squash Strategy

Enable auto-merge to automatically merge PRs when all checks pass:

```bash
# Enable auto-merge with squash strategy on current branch
gh pr merge --auto --squash

# Enable auto-merge on a specific PR number
gh pr merge 42 --auto --squash

# Combine PR creation with auto-merge
gh pr create --title "fix: correct bonus tile calculation" \
  --body "Fixes issue with end-of-round bonus scoring" && \
gh pr merge --auto --squash
```

### Example Workflow

Complete workflow from feature to auto-merged PR:

```bash
# 1. Create and switch to feature branch
git checkout -b feature/improve-ui

# 2. Make changes and commit
git add .
git commit -m "feat: improve score entry form layout"

# 3. Push branch to remote
git push -u origin feature/improve-ui

# 4. Create PR with auto-merge enabled
gh pr create \
  --title "feat: improve score entry form layout" \
  --body "Reorganizes the score entry form for better usability" && \
gh pr merge --auto --squash
```

### Additional Commands

```bash
# View PR status
gh pr status

# List all PRs
gh pr list

# View PR details
gh pr view [PR-number]

# Check PR checks status
gh pr checks
```

## Commit Standards

- Use conventional commit format: `type: description`
- Types: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`
- Keep commits focused and atomic
- Write clear, descriptive commit messages
- Don't include Claude Code in the commit message.

## Code Standards

- Follow Go best practices and idioms
- Run `go fmt` before committing
- Ensure code builds successfully with `go build`
- Run tests and ensure they pass before committing
- Write tests for new functionality
- Maintain test coverage above 70%

## Testing

### Running Tests

Run all tests:
```bash
go test ./...
```

Run tests with verbose output:
```bash
go test -v ./...
```

Run tests with coverage:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

View coverage in browser:
```bash
go tool cover -html=coverage.out
```

Run tests with race detection:
```bash
go test -race ./...
```

### Test Structure

Tests are organized alongside source files:
- `goals/scorer_test.go` - Round goal scoring tests
- `goals/selector_test.go` - Goal selection tests
- `goals/goals_test.go` - Goal definition tests
- `scoring/scoring_test.go` - End-game calculation tests
- `db/db_test.go` - Database initialization tests
- `db/game_results_test.go` - CRUD operation tests
- `main_test.go` - HTTP handler tests

### Writing Tests

- Use the `testify/assert` library for assertions
- Name test functions with the pattern `Test<FunctionName>_<Scenario>`
- Use table-driven tests for multiple scenarios
- Test both success and error cases
- Use temporary databases for database tests

### CI/CD Integration

Tests run automatically on every push to main via GitHub Actions. The build will fail if:
- Any test fails
- Race conditions are detected
- Code doesn't compile

## Project Context

This is a Wingspan board game scoring application built with Go, featuring:
- SQLite database for game results storage
- Web-based UI for score tracking
- Podman (Docker) containerization support
    - Deployed to Kubernetes in production.