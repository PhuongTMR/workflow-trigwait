# Contributing Guide

Thank you for considering contributing to Workflow Trigger and Wait! This guide will help you understand how to contribute effectively.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Release Process](#release-process)

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). By participating, you are expected to uphold this code.

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Git
- GitHub account
- Basic understanding of GitHub Actions

### Repository Structure

```
workflow-trigwait/
├── cmd/
│   ├── main.go           # Main application code
│   └── main_test.go      # Unit tests
├── dist/                 # Pre-built binaries for distribution
├── docs/                 # Documentation
├── scripts/
│   └── build.sh          # Build script for all platforms
├── action.yml            # GitHub Action definition
├── Dockerfile            # Docker container (legacy)
├── go.mod                # Go module definition
├── CLAUDE.md            # Claude Code documentation
└── README.md            # Main documentation
```

## Development Setup

### 1. Fork and Clone

```bash
# Fork the repository on GitHub, then:
git clone https://github.com/YOUR_USERNAME/workflow-trigwait.git
cd workflow-trigwait
```

### 2. Install Dependencies

The project has no external Go dependencies (uses only standard library).

```bash
# Verify Go installation
go version

# Download any tooling dependencies
go mod download
```

### 3. Verify Setup

```bash
# Run tests
go test -v ./cmd/...

# Build binary
go build -o workflow-trigwait ./cmd/main.go

# Run binary (will fail without env vars, but verifies it builds)
./workflow-trigwait
```

## Making Changes

### Branch Naming

Use descriptive branch names:

- `feature/add-retry-logic` - New features
- `fix/timeout-error` - Bug fixes
- `docs/update-readme` - Documentation
- `refactor/simplify-polling` - Code refactoring
- `test/add-integration-tests` - Test improvements

### Code Style

This project follows standard Go conventions:

1. **Use `gofmt`**
   ```bash
   gofmt -w cmd/main.go
   ```

2. **Run `go vet`**
   ```bash
   go vet ./cmd/...
   ```

3. **Follow Go best practices**
   - Short, descriptive variable names
   - Error handling at every step
   - Clear function names
   - Comments for exported functions

### Design Principles

1. **Zero External Dependencies**
   - Only use Go standard library
   - Keeps binary size small
   - Reduces security surface area

2. **Minimal Output**
   - Beautiful but concise logging
   - Use emojis for quick visual scanning
   - Avoid verbose messages

3. **Backward Compatibility**
   - Don't break existing configurations
   - Add new features as optional parameters
   - Deprecate gracefully with warnings

4. **Performance**
   - Optimize for fast GitHub Actions execution
   - Small binary size for quick downloads
   - Efficient API polling

### Common Changes

#### Adding a New Input Parameter

1. **Update `action.yml`:**
   ```yaml
   inputs:
     new_parameter:
       description: 'Description of the parameter'
       required: false
       default: 'default_value'
   ```

2. **Update `Config` struct in `cmd/main.go`:**
   ```go
   type Config struct {
       // ... existing fields
       NewParameter string
   }
   ```

3. **Load in `loadConfig()`:**
   ```go
   config := &Config{
       // ... existing fields
       NewParameter: getEnvOrDefault("INPUT_NEW_PARAMETER", "default_value"),
   }
   ```

4. **Add tests:**
   ```go
   func TestLoadConfig_NewParameter(t *testing.T) {
       os.Setenv("INPUT_NEW_PARAMETER", "test_value")
       defer os.Unsetenv("INPUT_NEW_PARAMETER")

       config, err := loadConfig()
       if err != nil {
           t.Fatal(err)
       }

       if config.NewParameter != "test_value" {
           t.Errorf("expected 'test_value', got '%s'", config.NewParameter)
       }
   }
   ```

5. **Update documentation:**
   - README.md inputs table
   - docs/USAGE_GUIDE.md examples
   - CLAUDE.md if relevant

#### Fixing a Bug

1. **Create a failing test:**
   ```go
   func TestBugFix_IssueXXX(t *testing.T) {
       // Test that reproduces the bug
   }
   ```

2. **Fix the bug**

3. **Verify test passes:**
   ```bash
   go test -v -run TestBugFix_IssueXXX ./cmd/...
   ```

4. **Run all tests:**
   ```bash
   go test -v ./cmd/...
   ```

#### Improving Output Messages

Update format strings in `cmd/main.go`:

```go
// Before
fmt.Println("Triggering workflow...")

// After - more beautiful and concise
fmt.Printf("🚀 Triggering %s/%s → %s\n", owner, repo, workflow)
```

## Testing

### Running Tests

```bash
# Run all tests
go test -v ./cmd/...

# Run with race detector
go test -v -race ./cmd/...

# Run specific test
go test -v -run TestTriggerWorkflow ./cmd/...

# Run benchmarks
go test -bench=. ./cmd/...

# Check coverage
go test -cover ./cmd/...
```

### Writing Tests

Follow the existing test patterns:

1. **Use `httptest` for API mocking:**
   ```go
   func TestAPIRequest(t *testing.T) {
       server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
           w.WriteHeader(http.StatusOK)
           json.NewEncoder(w).Encode(response)
       }))
       defer server.Close()

       // Test with server.URL
   }
   ```

2. **Set up and clean up environment:**
   ```go
   func TestFunction(t *testing.T) {
       os.Setenv("INPUT_OWNER", "test")
       defer os.Unsetenv("INPUT_OWNER")

       // Test code
   }
   ```

3. **Use table-driven tests:**
   ```go
   func TestMultipleCases(t *testing.T) {
       tests := []struct {
           name     string
           input    string
           expected string
       }{
           {"case1", "input1", "output1"},
           {"case2", "input2", "output2"},
       }

       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               // Test each case
           })
       }
   }
   ```

### Manual Testing

Test locally with environment variables:

```bash
# Build binary
go build -o workflow-trigwait ./cmd/main.go

# Test trigger and wait
INPUT_OWNER="my-org" \
INPUT_REPO="my-repo" \
INPUT_GITHUB_TOKEN="ghp_xxxxxxxxxxxx" \
INPUT_WORKFLOW_FILE_NAME="test.yml" \
INPUT_REF="main" \
INPUT_WAIT_INTERVAL=10 \
INPUT_TRIGGER_TIMEOUT=60 \
./workflow-trigwait
```

### Integration Testing

Create a test workflow in your fork:

```yaml
# .github/workflows/test-trigger.yml
name: Test Trigger

on:
  workflow_dispatch:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Test action
        uses: ./  # Use local action
        with:
          owner: ${{ github.repository_owner }}
          repo: ${{ github.event.repository.name }}
          github_token: ${{ secrets.GITHUB_TOKEN }}
          workflow_file_name: test-target.yml
```

## Submitting Changes

### Pull Request Process

1. **Create a branch:**
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make your changes**

3. **Run tests:**
   ```bash
   go test -v -race ./cmd/...
   ```

4. **Commit with clear messages:**
   ```bash
   git commit -m "feat: add retry logic for API failures"
   ```

   Use conventional commit prefixes:
   - `feat:` - New feature
   - `fix:` - Bug fix
   - `docs:` - Documentation changes
   - `test:` - Test changes
   - `refactor:` - Code refactoring
   - `perf:` - Performance improvements
   - `chore:` - Build/tooling changes

5. **Push to your fork:**
   ```bash
   git push origin feature/my-feature
   ```

6. **Create Pull Request:**
   - Go to GitHub and create a PR
   - Fill out the PR template
   - Link related issues
   - Add screenshots for UI changes

### Pull Request Guidelines

**Good PR:**
- Clear title describing the change
- Description explaining why the change is needed
- Tests for new functionality
- Documentation updates
- Small, focused changes

**PR Title Examples:**
- ✅ `feat: add support for GitHub Enterprise Server`
- ✅ `fix: handle timeout errors gracefully`
- ✅ `docs: add examples for parallel workflows`
- ❌ `Update files`
- ❌ `Fix stuff`

### Pull Request Template

```markdown
## Description
Brief description of changes.

## Motivation
Why is this change needed? What problem does it solve?

## Changes
- Bullet point list of changes

## Testing
How was this tested?

## Screenshots
If applicable, add screenshots.

## Checklist
- [ ] Tests pass locally
- [ ] Added/updated tests for changes
- [ ] Updated documentation
- [ ] Follows code style guidelines
- [ ] No breaking changes (or documented)
```

## Release Process

Releases are managed by maintainers. The process:

1. **Update Version:**
   - Update version in documentation
   - Create changelog entry

2. **Build Binaries:**
   ```bash
   ./scripts/build.sh
   ```

3. **Create Git Tag:**
   ```bash
   git tag -a v1.7.0 -m "Release v1.7.0"
   git push origin v1.7.0
   ```

4. **Create GitHub Release:**
   - Draft new release
   - Attach binaries from `dist/`
   - Include changelog
   - Mark as latest release

5. **Update Major Version Tag:**
   ```bash
   git tag -fa v1 -m "Update v1 to v1.7.0"
   git push origin v1 --force
   ```

## Binary Size Optimization

When adding features, keep binary size in mind:

### Check Binary Size

```bash
# Build and check size
go build -o test-binary ./cmd/main.go
ls -lh test-binary

# Compare with main branch
git checkout main
go build -o main-binary ./cmd/main.go
ls -lh main-binary test-binary
```

### Optimization Tips

1. **Avoid new dependencies:**
   ```go
   // ❌ Avoid
   import "github.com/some/package"

   // ✅ Use standard library
   import "encoding/json"
   ```

2. **Use build flags:**
   ```bash
   go build -ldflags="-s -w -buildid=" -trimpath -o binary ./cmd/main.go
   ```

3. **Profile binary size:**
   ```bash
   go tool nm -size binary | sort -rn | head -20
   ```

## Code Review Process

### For Contributors

- Be open to feedback
- Respond to review comments promptly
- Make requested changes or explain why not
- Keep discussions professional and focused

### For Reviewers

- Review within 2-3 days if possible
- Provide constructive feedback
- Test the changes if possible
- Approve when satisfied

## Getting Help

- **Questions:** Use [GitHub Discussions](https://github.com/PhuongTMR/workflow-trigwait/discussions)
- **Bugs:** Create an [Issue](https://github.com/PhuongTMR/workflow-trigwait/issues)
- **Security:** Email maintainers privately (see SECURITY.md)

## Recognition

Contributors will be:
- Listed in release notes
- Mentioned in CHANGELOG.md
- Credited in commit history

Thank you for contributing! 🎉
