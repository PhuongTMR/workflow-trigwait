# Usage Guide

This guide provides comprehensive examples and best practices for using the Workflow Trigger and Wait action.

## Table of Contents

- [Quick Start](#quick-start)
- [Common Use Cases](#common-use-cases)
- [Advanced Configuration](#advanced-configuration)
- [Best Practices](#best-practices)
- [Output Usage](#output-usage)
- [Troubleshooting](#troubleshooting)

## Quick Start

### Minimal Setup

The simplest way to trigger and wait for a workflow:

```yaml
name: Trigger Deployment

on:
  push:
    branches: [main]

jobs:
  trigger-deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: PhuongTMR/workflow-trigwait@v1
        with:
          owner: my-org
          repo: my-repo
          github_token: ${{ secrets.PERSONAL_ACCESS_TOKEN }}
          workflow_file_name: deploy.yml
```

### What You Need

1. **GitHub Token**: A Personal Access Token (PAT) or GitHub App token with:
   - `repo` scope (for private repositories)
   - `actions:write` permission (to trigger workflows)
   - `actions:read` permission (to check workflow status)

2. **Target Workflow**: Must support `workflow_dispatch` event

## Common Use Cases

### 1. Cross-Repository Deployment Pipeline

Trigger a deployment in another repository and wait for completion:

```yaml
name: Production Deployment Pipeline

on:
  release:
    types: [published]

jobs:
  deploy-frontend:
    runs-on: ubuntu-latest
    steps:
      - name: Trigger frontend deployment
        uses: PhuongTMR/workflow-trigwait@v1
        with:
          owner: acme-corp
          repo: frontend
          github_token: ${{ secrets.DEPLOYMENT_TOKEN }}
          workflow_file_name: deploy.yml
          ref: main
          client_payload: |
            {
              "environment": "production",
              "version": "${{ github.event.release.tag_name }}"
            }

  deploy-backend:
    runs-on: ubuntu-latest
    needs: deploy-frontend
    steps:
      - name: Trigger backend deployment
        uses: PhuongTMR/workflow-trigwait@v1
        with:
          owner: acme-corp
          repo: backend
          github_token: ${{ secrets.DEPLOYMENT_TOKEN }}
          workflow_file_name: deploy.yml
          client_payload: '{"environment": "production"}'
```

### 2. Parallel Workflow Execution

Trigger multiple workflows simultaneously:

```yaml
name: Run Test Suites

on:
  pull_request:

jobs:
  trigger-tests:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        test-suite:
          - unit-tests
          - integration-tests
          - e2e-tests
    steps:
      - name: Trigger ${{ matrix.test-suite }}
        uses: PhuongTMR/workflow-trigwait@v1
        with:
          owner: my-org
          repo: my-repo
          github_token: ${{ secrets.GITHUB_TOKEN }}
          workflow_file_name: ${{ matrix.test-suite }}.yml
          ref: ${{ github.head_ref }}
```

### 3. Multi-Environment Deployment

Deploy to multiple environments sequentially:

```yaml
name: Multi-Environment Deployment

on:
  workflow_dispatch:
    inputs:
      version:
        description: 'Version to deploy'
        required: true

jobs:
  deploy-staging:
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to staging
        id: staging
        uses: PhuongTMR/workflow-trigwait@v1
        with:
          owner: my-org
          repo: deployment-repo
          github_token: ${{ secrets.DEPLOYMENT_TOKEN }}
          workflow_file_name: deploy.yml
          client_payload: |
            {
              "environment": "staging",
              "version": "${{ github.event.inputs.version }}"
            }

      - name: Verify staging
        run: |
          echo "Staging deployment: ${{ steps.staging.outputs.conclusion }}"
          curl -f https://staging.example.com/health

  deploy-production:
    runs-on: ubuntu-latest
    needs: deploy-staging
    steps:
      - name: Deploy to production
        uses: PhuongTMR/workflow-trigwait@v1
        with:
          owner: my-org
          repo: deployment-repo
          github_token: ${{ secrets.DEPLOYMENT_TOKEN }}
          workflow_file_name: deploy.yml
          client_payload: |
            {
              "environment": "production",
              "version": "${{ github.event.inputs.version }}"
            }
```

### 4. Fire-and-Forget Workflow Trigger

Trigger a workflow without waiting for it to complete:

```yaml
name: Trigger Background Job

on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM

jobs:
  trigger-cleanup:
    runs-on: ubuntu-latest
    steps:
      - name: Trigger cleanup job
        uses: PhuongTMR/workflow-trigwait@v1
        with:
          owner: my-org
          repo: maintenance-repo
          github_token: ${{ secrets.GITHUB_TOKEN }}
          workflow_file_name: cleanup.yml
          trigger_workflow: true
          wait_workflow: false  # Don't wait for completion
```

### 5. Conditional Failure Propagation

Continue even if downstream workflow fails:

```yaml
name: Run Optional Tests

on:
  pull_request:

jobs:
  required-tests:
    runs-on: ubuntu-latest
    steps:
      - name: Run required tests
        uses: PhuongTMR/workflow-trigwait@v1
        with:
          owner: my-org
          repo: my-repo
          github_token: ${{ secrets.GITHUB_TOKEN }}
          workflow_file_name: required-tests.yml
          propagate_failure: true  # Fail if tests fail

  optional-tests:
    runs-on: ubuntu-latest
    steps:
      - name: Run optional tests
        id: optional
        uses: PhuongTMR/workflow-trigwait@v1
        with:
          owner: my-org
          repo: my-repo
          github_token: ${{ secrets.GITHUB_TOKEN }}
          workflow_file_name: optional-tests.yml
          propagate_failure: false  # Continue even if tests fail

      - name: Report optional test results
        run: |
          echo "Optional tests: ${{ steps.optional.outputs.conclusion }}"
```

## Advanced Configuration

### Dynamic/Optional Inputs

The action automatically removes empty values from `client_payload` before sending the request. This is useful for optional or conditional inputs:

```yaml
- name: Trigger with optional inputs
  uses: PhuongTMR/workflow-trigwait@v1
  with:
    owner: my-org
    repo: my-repo
    github_token: ${{ secrets.GITHUB_TOKEN }}
    workflow_file_name: deploy.yml
    client_payload: |
      {
        "environment": "production",
        "version": "${{ github.event.inputs.version }}",
        "hotfix": "${{ github.event.inputs.hotfix }}",
        "rollback_version": ""
      }
```

**Behavior:**
- Empty strings (`""`) are removed
- Null values (`null`) are removed
- False booleans (`false`) and zero (`0`) are kept
- Nested empty objects are removed

**Example:**

Input payload:
```json
{
  "environment": "production",
  "version": "v1.2.3",
  "hotfix": "",
  "rollback_version": ""
}
```

Sent to target workflow:
```json
{
  "environment": "production",
  "version": "v1.2.3"
}
```

This allows you to use dynamic GitHub expressions without worrying about passing empty values:

```yaml
client_payload: |
  {
    "pr_number": "${{ github.event.pull_request.number }}",
    "branch": "${{ github.head_ref }}",
    "tag": "${{ github.ref_name }}"
  }
```

If running on a push (not PR), `pr_number` will be empty and automatically removed.

### Reliable Workflow Correlation

For production environments with concurrent workflow triggers, use distinct ID correlation:

**Trigger Workflow:**
```yaml
- name: Trigger with correlation
  uses: PhuongTMR/workflow-trigwait@v1
  with:
    owner: my-org
    repo: my-repo
    github_token: ${{ secrets.GITHUB_TOKEN }}
    workflow_file_name: deploy.yml
    distinct_id_name: correlation_id  # Enable correlation
    client_payload: '{"environment": "production"}'
```

**Target Workflow:**
```yaml
name: Deploy
run-name: Deploy [${{ inputs.correlation_id }}]

on:
  workflow_dispatch:
    inputs:
      correlation_id:
        description: 'Correlation ID for tracking'
        required: false
        type: string
      environment:
        description: 'Target environment'
        required: true
        type: string

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Deploy
        run: echo "Deploying to ${{ inputs.environment }}"
```

### Custom Polling Configuration

Adjust polling behavior for different scenarios:

```yaml
# Fast polling for quick jobs
- uses: PhuongTMR/workflow-trigwait@v1
  with:
    owner: my-org
    repo: my-repo
    github_token: ${{ secrets.GITHUB_TOKEN }}
    workflow_file_name: quick-job.yml
    wait_interval: 5  # Poll every 5 seconds
    trigger_timeout: 30  # Fail if not found in 30 seconds

# Slow polling for long-running jobs
- uses: PhuongTMR/workflow-trigwait@v1
  with:
    owner: my-org
    repo: my-repo
    github_token: ${{ secrets.GITHUB_TOKEN }}
    workflow_file_name: long-job.yml
    wait_interval: 60  # Poll every 60 seconds (queued jobs automatically poll less frequently)
    trigger_timeout: 300  # Wait 5 minutes to find workflow
```

### Working with Different Branches

Trigger workflows on specific branches or tags:

```yaml
# Trigger on specific branch
- uses: PhuongTMR/workflow-trigwait@v1
  with:
    owner: my-org
    repo: my-repo
    github_token: ${{ secrets.GITHUB_TOKEN }}
    workflow_file_name: test.yml
    ref: develop

# Trigger on specific tag
- uses: PhuongTMR/workflow-trigwait@v1
  with:
    owner: my-org
    repo: my-repo
    github_token: ${{ secrets.GITHUB_TOKEN }}
    workflow_file_name: release.yml
    ref: v1.2.3

# Trigger on specific commit SHA
- uses: PhuongTMR/workflow-trigwait@v1
  with:
    owner: my-org
    repo: my-repo
    github_token: ${{ secrets.GITHUB_TOKEN }}
    workflow_file_name: build.yml
    ref: ${{ github.sha }}
```

## Best Practices

### 1. Use Specific Versions

Always pin to a specific version for production:

```yaml
# ✅ Good - specific version
- uses: PhuongTMR/workflow-trigwait@v1.6.1

# ⚠️ Acceptable - major version
- uses: PhuongTMR/workflow-trigwait@v1

# ❌ Bad - unstable reference
- uses: PhuongTMR/workflow-trigwait@master
```

### 2. Store Tokens Securely

Never hardcode tokens; always use secrets:

```yaml
# ✅ Good
github_token: ${{ secrets.DEPLOYMENT_TOKEN }}

# ❌ Bad
github_token: ghp_xxxxxxxxxxxxxxxxxxxx
```

### 3. Use Descriptive Step IDs

Name steps clearly for easier output reference:

```yaml
- name: Deploy to production
  id: prod-deploy
  uses: PhuongTMR/workflow-trigwait@v1
  with:
    # ... configuration

- name: Notify Slack
  run: |
    echo "Deployment: ${{ steps.prod-deploy.outputs.conclusion }}"
```

### 4. Enable Distinct ID for Production

Use distinct ID correlation for environments with concurrent triggers:

```yaml
- uses: PhuongTMR/workflow-trigwait@v1
  with:
    # ... other config
    distinct_id_name: correlation_id  # Prevents race conditions
```

### 5. Handle Failures Appropriately

Decide whether to propagate failures based on criticality:

```yaml
# Critical workflows - propagate failures
- uses: PhuongTMR/workflow-trigwait@v1
  with:
    workflow_file_name: security-scan.yml
    propagate_failure: true

# Non-critical workflows - continue on failure
- uses: PhuongTMR/workflow-trigwait@v1
  with:
    workflow_file_name: performance-test.yml
    propagate_failure: false
```

## Output Usage

### Accessing Workflow Results

```yaml
- name: Trigger workflow
  id: trigger
  uses: PhuongTMR/workflow-trigwait@v1
  with:
    owner: my-org
    repo: my-repo
    github_token: ${{ secrets.GITHUB_TOKEN }}
    workflow_file_name: build.yml

- name: Use outputs
  run: |
    echo "Workflow ID: ${{ steps.trigger.outputs.workflow_id }}"
    echo "Workflow URL: ${{ steps.trigger.outputs.workflow_url }}"
    echo "Conclusion: ${{ steps.trigger.outputs.conclusion }}"
    echo "Distinct ID: ${{ steps.trigger.outputs.distinct_id }}"
```

### Conditional Logic Based on Results

```yaml
- name: Run deployment
  id: deploy
  uses: PhuongTMR/workflow-trigwait@v1
  with:
    owner: my-org
    repo: my-repo
    github_token: ${{ secrets.GITHUB_TOKEN }}
    workflow_file_name: deploy.yml
    propagate_failure: false

- name: Rollback on failure
  if: steps.deploy.outputs.conclusion != 'success'
  run: |
    echo "Deployment failed, initiating rollback..."
    # Rollback logic here

- name: Send notification
  if: always()
  run: |
    STATUS="${{ steps.deploy.outputs.conclusion }}"
    URL="${{ steps.deploy.outputs.workflow_url }}"
    curl -X POST $WEBHOOK_URL \
      -d "Deployment ${STATUS}: ${URL}"
```

### Create GitHub Deployment

```yaml
- name: Create deployment
  id: deployment
  uses: actions/github-script@v7
  with:
    script: |
      const deployment = await github.rest.repos.createDeployment({
        owner: context.repo.owner,
        repo: context.repo.repo,
        ref: context.sha,
        environment: 'production',
        required_contexts: []
      });
      return deployment.data.id;

- name: Trigger workflow
  id: deploy
  uses: PhuongTMR/workflow-trigwait@v1
  with:
    owner: my-org
    repo: my-repo
    github_token: ${{ secrets.GITHUB_TOKEN }}
    workflow_file_name: deploy.yml

- name: Update deployment status
  uses: actions/github-script@v7
  if: always()
  with:
    script: |
      await github.rest.repos.createDeploymentStatus({
        owner: context.repo.owner,
        repo: context.repo.repo,
        deployment_id: ${{ steps.deployment.outputs.result }},
        state: '${{ steps.deploy.outputs.conclusion }}' === 'success' ? 'success' : 'failure',
        log_url: '${{ steps.deploy.outputs.workflow_url }}'
      });
```

## Troubleshooting

### Common Issues

#### Workflow Not Found

**Error:** `timeout: workflow run did not appear within 120s`

**Solutions:**
1. Increase `trigger_timeout`:
   ```yaml
   trigger_timeout: 300  # Wait 5 minutes
   ```

2. Verify workflow file name matches exactly:
   ```yaml
   workflow_file_name: deploy.yml  # Must match .github/workflows/deploy.yml
   ```

3. Enable distinct ID correlation:
   ```yaml
   distinct_id_name: correlation_id
   ```

#### Permission Denied

**Error:** `API request failed: 403 Forbidden`

**Solutions:**
1. Verify token has required permissions:
   - `repo` scope
   - `actions:write`
   - `actions:read`

2. For cross-repo triggers, use a PAT instead of `GITHUB_TOKEN`

3. Check repository settings allow workflow triggers

#### Target Workflow Not Accepting workflow_dispatch

**Error:** `Workflow does not accept workflow_dispatch event`

**Solution:** Add `workflow_dispatch` to target workflow:
```yaml
on:
  workflow_dispatch:
    inputs: {}  # Add any required inputs
```

### Getting Help

1. Check the [GitHub Discussions](https://github.com/PhuongTMR/workflow-trigwait/discussions)
2. Review [existing issues](https://github.com/PhuongTMR/workflow-trigwait/issues)
3. Enable debug logging:
   ```yaml
   - name: Enable debug logging
     run: echo "ACTIONS_STEP_DEBUG=true" >> $GITHUB_ENV
   ```

## See Also

- [README.md](../README.md) - Overview and quick start
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Detailed troubleshooting guide
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Contributing guidelines
