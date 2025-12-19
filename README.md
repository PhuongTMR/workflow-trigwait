# Trigger Workflow and Wait

A GitHub Action that triggers a workflow in another repository (or the same repository) and waits for it to complete. Built with Go for fast execution and minimal dependencies.

[![GitHub release](https://img.shields.io/github/v/release/PhuongTMR/workflow-trigwait)](https://github.com/PhuongTMR/workflow-trigwait/releases)
[![Build](https://github.com/PhuongTMR/workflow-trigwait/workflows/Build/badge.svg)](https://github.com/PhuongTMR/workflow-trigwait/actions)
[![Go Report](https://goreportcard.com/badge/github.com/PhuongTMR/workflow-trigwait)](https://goreportcard.com/report/github.com/PhuongTMR/workflow-trigwait)

## Features

- 🚀 **Trigger workflows** via `workflow_dispatch` event
- ⏳ **Wait for completion** with configurable polling interval
- 📊 **Propagate failures** from downstream workflows (optional)
- 🔧 **Flexible configuration** - trigger only, wait only, or both
- ⚡ **Fast execution** - pre-built binaries (1.5-5.3 MB)
- 🎯 **Reliable correlation** - optional distinct ID matching for concurrent triggers
- 🌐 **Cross-repository** - trigger workflows in any accessible repository

## Quick Links

- 📖 [Usage Guide](docs/USAGE_GUIDE.md) - Comprehensive examples and patterns
- 🔧 [Troubleshooting](docs/TROUBLESHOOTING.md) - Common issues and solutions
- 🤝 [Contributing](CONTRIBUTING.md) - How to contribute to this project

## Inputs

| Input                | Required | Default | Description |
| -------------------- | -------- | ------- | ----------- |
| `owner`              | ✅       | -       | Repository owner where the workflow is located |
| `repo`               | ✅       | -       | Repository name where the workflow is located |
| `github_token`       | ✅       | -       | GitHub access token with `repo` and `actions` permissions |
| `workflow_file_name` | ✅       | -       | Workflow file name (e.g., `deploy.yml`) |
| `ref`                | ❌       | `main`  | Branch, tag, or commit SHA to run the workflow on |
| `wait_interval`      | ❌       | `10`    | Seconds between status checks (adaptive polling: slower when queued) |
| `trigger_timeout`    | ❌       | `120`   | Seconds to wait for triggered workflow to appear |
| `client_payload`     | ❌       | `{}`    | JSON string of inputs to pass to the workflow |
| `propagate_failure`  | ❌       | `true`  | Fail this job if the downstream workflow fails |
| `trigger_workflow`   | ❌       | `true`  | Whether to trigger the workflow |
| `wait_workflow`      | ❌       | `true`  | Whether to wait for the workflow to complete |
| `distinct_id_name`   | ❌       | -       | Input field name for workflow correlation (enables reliable run identification) |

## Outputs

| Output         | Description |
| -------------- | ----------- |
| `workflow_id`  | The ID of the triggered workflow run |
| `workflow_url` | URL to the workflow run in GitHub Actions |
| `conclusion`   | Final status of the workflow (`success`, `failure`, `cancelled`, etc.) |
| `distinct_id`  | Unique identifier used to correlate the trigger with the workflow run |

## Workflow Correlation (Optional)

By default, this action uses **time-based matching** to find the triggered workflow run. This works well for most cases but can be unreliable with concurrent triggers.

For reliable identification, enable **distinct_id correlation** by setting `distinct_id_name`:

**How it works:**
1. A unique ID is auto-generated and passed as an input to the target workflow
2. The target workflow includes the ID in its `run-name` field
3. The action matches workflow runs by checking the `display_title` - no extra API calls needed

**Required setup in your target workflow:**

```yaml
name: My Workflow
run-name: My Workflow [${{ inputs.distinct_id }}]

on:
  workflow_dispatch:
    inputs:
      distinct_id:
        description: 'Unique identifier for correlating workflow runs'
        required: false
        type: string

jobs:
  your-job:
    runs-on: ubuntu-latest
    steps:
      # ... your steps
```

**Using a custom input name (e.g., reusing an existing `id` field):**

If your workflow already has an identifier input like `id`, you can reuse it:

```yaml
# Your trigger configuration
- uses: PhuongTMR/workflow-trigwait@master
  with:
    owner: my-org
    repo: my-repo
    github_token: ${{ secrets.PERSONAL_ACCESS_TOKEN }}
    workflow_file_name: build.yml
    distinct_id_name: id  # Use 'id' instead of 'distinct_id'
```

```yaml
# Your target workflow
name: Build
run-name: Build [${{ inputs.id }}]

on:
  workflow_dispatch:
    inputs:
      id:
        description: 'run identifier'
        required: false

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      # ... your steps
```

> **Note:** Without this setup, the action falls back to time-based matching which may be unreliable with concurrent triggers.

## Examples

### Basic Usage

Trigger a workflow and wait for completion:

```yaml
- uses: PhuongTMR/workflow-trigwait@v1
  with:
    owner: my-org
    repo: my-repo
    github_token: ${{ secrets.PERSONAL_ACCESS_TOKEN }}
    workflow_file_name: deploy.yml
```

### With Inputs and Branch

Pass inputs to the triggered workflow:

```yaml
- uses: PhuongTMR/workflow-trigwait@v1
  with:
    owner: my-org
    repo: my-repo
    github_token: ${{ secrets.PERSONAL_ACCESS_TOKEN }}
    workflow_file_name: deploy.yml
    ref: develop
    client_payload: '{"environment": "staging", "version": "1.2.3"}'
```

### Reliable Correlation (Recommended for Production)

Enable distinct ID correlation for concurrent triggers:

```yaml
- uses: PhuongTMR/workflow-trigwait@v1
  with:
    owner: my-org
    repo: my-repo
    github_token: ${{ secrets.PERSONAL_ACCESS_TOKEN }}
    workflow_file_name: deploy.yml
    distinct_id_name: correlation_id  # Enable correlation
```

Your target workflow must include the ID in `run-name`:

```yaml
name: Deploy
run-name: Deploy [${{ inputs.correlation_id }}]

on:
  workflow_dispatch:
    inputs:
      correlation_id:
        type: string
        required: false
```

### Using Outputs

Access workflow results:

```yaml
- name: Trigger deployment
  id: deploy
  uses: PhuongTMR/workflow-trigwait@v1
  with:
    owner: my-org
    repo: my-repo
    github_token: ${{ secrets.GITHUB_TOKEN }}
    workflow_file_name: deploy.yml

- name: Check result
  run: |
    echo "Workflow ID: ${{ steps.deploy.outputs.workflow_id }}"
    echo "Workflow URL: ${{ steps.deploy.outputs.workflow_url }}"
    echo "Conclusion: ${{ steps.deploy.outputs.conclusion }}"
```

### More Examples

For more advanced use cases, see the [Usage Guide](docs/USAGE_GUIDE.md):
- Cross-repository deployment pipelines
- Parallel workflow execution
- Multi-environment deployments
- Fire-and-forget triggers
- Conditional failure handling

## Target Workflow Requirements

The target workflow must be configured to accept `workflow_dispatch` events. If using `distinct_id_name` for correlation, include the input in the `run-name`:

```yaml
name: Deploy
run-name: Deploy [${{ inputs.distinct_id }}]

on:
  workflow_dispatch:
    inputs:
      distinct_id:
        type: string
        description: Correlation ID (auto-generated by trigger action)
      environment:
        type: string
        description: Target environment
        default: 'staging'
      version:
        type: string
        description: Version to deploy

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Deploy
        run: |
          echo "Deploying ${{ inputs.version }} to ${{ inputs.environment }}"
```

## Token Permissions

The `github_token` requires these permissions:

| Permission | Scope | Purpose |
|------------|-------|---------|
| `repo` | Full control | Access private repositories |
| `actions:write` | Write | Trigger workflows |
| `actions:read` | Read | Check workflow status |

### Cross-Repository Triggers

The default `GITHUB_TOKEN` only works within the same repository. For cross-repository triggers, use:

- **Personal Access Token (PAT)** - Classic or fine-grained
- **GitHub App Token** - Recommended for organization-level automation

**Example: Fine-grained PAT permissions**
- Repository access: Select the target repository
- Permissions:
  - Actions: Read and write
  - Contents: Read (if workflow needs to check out code)

## How It Works

```
┌─────────────┐
│   Trigger   │  Send workflow_dispatch event to GitHub API
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Find Run   │  Poll workflow runs API to locate the triggered run
└──────┬──────┘  (Uses distinct ID or time-based matching)
       │
       ▼
┌─────────────┐
│    Wait     │  Poll run status until completion
└──────┬──────┘  (Adaptive polling: faster when running, slower when queued)
       │
       ▼
┌─────────────┐
│   Report    │  Set outputs: workflow_id, workflow_url, conclusion
└─────────────┘
```

## Troubleshooting

### Common Issues

**Timeout finding workflow run**
- Enable distinct ID correlation: `distinct_id_name: correlation_id`
- Increase timeout: `trigger_timeout: 300`
- See [Troubleshooting Guide](docs/TROUBLESHOOTING.md#timeout-workflow-run-did-not-appear)

**Permission denied (403)**
- Use PAT instead of `GITHUB_TOKEN` for cross-repo triggers
- Verify token has `repo`, `actions:write`, and `actions:read` permissions
- See [Troubleshooting Guide](docs/TROUBLESHOOTING.md#api-request-failed-403-forbidden)

**Workflow not found (404)**
- Check repository owner and name are correct
- Verify workflow file name matches exactly
- Ensure token has access to the repository

For more issues and solutions, see the [Troubleshooting Guide](docs/TROUBLESHOOTING.md).

## Local Development

### Prerequisites

- Go 1.21+
- UPX (optional, for binary compression)

### Quick Start

```bash
# Clone repository
git clone https://github.com/PhuongTMR/workflow-trigwait.git
cd workflow-trigwait

# Run tests
go test -v -race ./cmd/...

# Build binary
go build -o workflow-trigwait ./cmd/main.go

# Test locally
INPUT_OWNER="my-org" \
INPUT_REPO="my-repo" \
INPUT_GITHUB_TOKEN="ghp_xxxx" \
INPUT_WORKFLOW_FILE_NAME="test.yml" \
./workflow-trigwait
```

### Build All Platform Binaries

```bash
# Build optimized binaries for all platforms
./scripts/build.sh

# Outputs to dist/
ls -lh dist/
```

For detailed development instructions, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Versioning

Use specific version tags for stability:

```yaml
# ✅ Recommended: Specific version
- uses: PhuongTMR/workflow-trigwait@v1.7.0

# ✅ Good: Major version (auto-updates minor/patch)
- uses: PhuongTMR/workflow-trigwait@v1

# ⚠️ Not recommended: Latest from main branch
- uses: PhuongTMR/workflow-trigwait@master
```

See [all releases](https://github.com/PhuongTMR/workflow-trigwait/releases) for changelogs.

## Performance

- **Binary size**: 1.5-1.7 MB (Linux/Windows, with UPX), 5.0-5.3 MB (macOS)
- **Startup time**: < 100ms
- **Memory usage**: ~10 MB
- **API calls**: Optimized with exponential backoff

The action uses pre-built binaries for fast initialization in GitHub Actions runners.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Development setup
- Code style guidelines
- Testing procedures
- Pull request process

## License

See [LICENSE](LICENSE) for details.
