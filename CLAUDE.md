# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a GitHub Action that triggers workflows in repositories via `workflow_dispatch` and optionally waits for completion. The action is implemented as a standalone Go binary that runs in GitHub Actions composite action format.

## Documentation Structure

The project has comprehensive documentation:
- **README.md** - Main user-facing documentation with quick start
- **docs/USAGE_GUIDE.md** - Extensive usage examples and patterns
- **docs/TROUBLESHOOTING.md** - Common issues and solutions
- **CONTRIBUTING.md** - Development and contribution guidelines
- **docs/INDEX.md** - Documentation navigation

When making changes, ensure all relevant documentation is updated.

## Key Architecture

### Single-File Go Binary
- All logic is contained in `cmd/main.go` (~400 lines)
- No external Go dependencies - uses only standard library
- Configuration is loaded via environment variables prefixed with `INPUT_`
- Outputs are written to `$GITHUB_OUTPUT` file

### Execution Model
The tool has three main phases that can be independently controlled:
1. **Trigger**: Dispatch workflow via GitHub API (`POST /repos/:owner/:repo/actions/workflows/:workflow_file/dispatches`)
2. **Find Run**: Poll workflow runs API to locate the triggered run
3. **Wait**: Poll run status until completion

### Workflow Correlation Strategy
Two methods for identifying the triggered workflow run:
- **Time-based matching** (default): Finds runs created after trigger time
- **Distinct ID matching** (optional): Generates unique ID, passes as input, matches via `display_title` field in run metadata

Key implementation detail in `findWorkflowRun()` (cmd/main.go:228):
- If `distinct_id_name` is set, the action auto-generates a unique ID and adds it to `client_payload`
- The target workflow must include this ID in its `run-name` field
- Matching is done by checking if `WorkflowRun.DisplayTitle` contains the distinct ID

### Adaptive Polling
The action uses intelligent polling intervals (cmd/main.go:330-336):
- Slower polling (30s+) when workflow is queued/pending
- Faster polling (user-configured) when workflow is in progress
- Exponential backoff during trigger phase to reduce API calls

### Cross-Platform Distribution
The action ships pre-built binaries for multiple platforms (detected at runtime in action.yml:77-96):
- Linux: amd64, arm64
- macOS (Darwin): amd64, arm64
- Windows: amd64
- Falls back to building from source if binary not found

## Development Commands

### Run Tests
```bash
go test -v -race ./cmd/...
```

### Build Binary Locally
```bash
go build -o workflow-trigwait ./cmd/main.go
```

### Build All Platform Binaries
```bash
./scripts/build.sh
```
This creates optimized binaries in `dist/` with optional UPX compression.

### Test Locally
Set environment variables with `INPUT_` prefix:
```bash
INPUT_OWNER="my-org" \
INPUT_REPO="my-repo" \
INPUT_GITHUB_TOKEN="ghp_xxxx" \
INPUT_WORKFLOW_FILE_NAME="test.yml" \
INPUT_REF="main" \
INPUT_WAIT_INTERVAL=10 \
INPUT_TRIGGER_TIMEOUT=60 \
INPUT_CLIENT_PAYLOAD='{"test": "true"}' \
INPUT_PROPAGATE_FAILURE=true \
INPUT_TRIGGER_WORKFLOW=true \
INPUT_WAIT_WORKFLOW=true \
./workflow-trigwait
```

### Build Docker Image
```bash
docker build -t workflow-trigwait .
```

## Important Implementation Details

### Configuration Loading (cmd/main.go:81)
- All inputs come from environment variables with `INPUT_` prefix
- Boolean defaults: `propagate_failure=true`, `trigger_workflow=true`, `wait_workflow=true`
- String defaults: `ref=main`, `wait_interval=10`, `trigger_timeout=120`
- `client_payload` must be valid JSON if provided

### GitHub API Integration
- All API requests go through `apiRequest()` function (cmd/main.go:362)
- Uses standard GitHub REST API v3 (`application/vnd.github.v3+json`)
- Supports GitHub Enterprise via `GITHUB_API_URL` and `GITHUB_SERVER_URL` env vars
- API timeout is hardcoded to 30 seconds per request

### Error Handling & Output
- Exits with status 1 on any error
- If `propagate_failure=true` (default), exits with error when downstream workflow fails
- Always sets outputs: `workflow_id`, `workflow_url`, `conclusion`, `distinct_id` (if enabled)

### Testing Strategy
Tests use `httptest.NewServer()` to mock GitHub API responses. Key test scenarios:
- Configuration validation and defaults
- API request/response handling
- Workflow run discovery (time-based and distinct ID)
- Polling and completion detection
- Error propagation behavior

## Code Style & Patterns

### Minimalist Dependencies
The codebase deliberately avoids external dependencies to:
- Minimize binary size
- Reduce supply chain risk
- Simplify maintenance
- Speed up builds

### Structured Output
Uses formatted output with box drawing characters for better readability in GitHub Actions logs:
```
══════════════════════════════════════════════════════════════
  Workflow Trigger
══════════════════════════════════════════════════════════════
```

### Config Struct Pattern
All configuration is centralized in the `Config` struct (cmd/main.go:17), loaded once at startup, then passed to all functions.

## Binary Size Optimization

The build process is heavily optimized to minimize binary size for faster GitHub Actions downloads:

### Go Build Flags (`scripts/build.sh`)
- `-ldflags="-s -w -buildid="`: Strip debug info, symbol table, and build ID
- `-trimpath`: Remove file system paths from binary
- `CGO_ENABLED=0`: Static linking, no C dependencies

### UPX Compression
- Applied to Linux and Windows binaries (not macOS due to code signing issues)
- Uses LZMA compression algorithm (`--best --lzma`)
- Achieves ~70% size reduction (5.2 MB → 1.7 MB)

### Binary Sizes
- **Linux/Windows (compressed)**: 1.5-1.7 MB (with UPX)
- **macOS (uncompressed)**: 5.0-5.3 MB (UPX not applied)
- **Total repository size**: ~18 MB for all 5 platform binaries

The compression is critical for GitHub Actions performance - smaller binaries mean faster action initialization when users reference this action from other repositories.

## GitHub Actions Integration

The action is defined as a composite action (action.yml), not a container action. This means:
- The Go binary runs directly on the runner (faster startup)
- No Docker overhead
- Binaries are committed to the repository in `dist/`
- Platform detection happens at runtime using `uname`
