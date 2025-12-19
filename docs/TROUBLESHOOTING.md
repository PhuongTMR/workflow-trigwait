# Troubleshooting Guide

This guide helps you diagnose and resolve common issues when using the Workflow Trigger and Wait action.

## Table of Contents

- [Quick Diagnostics](#quick-diagnostics)
- [Common Errors](#common-errors)
- [Performance Issues](#performance-issues)
- [Token and Permission Issues](#token-and-permission-issues)
- [Workflow Correlation Issues](#workflow-correlation-issues)
- [Debugging Tips](#debugging-tips)

## Quick Diagnostics

### Check Status Checklist

Before diving into specific issues, verify these basics:

- [ ] GitHub token is valid and not expired
- [ ] Token has `repo`, `actions:write`, and `actions:read` permissions
- [ ] Target workflow exists at `.github/workflows/<workflow_file_name>`
- [ ] Target workflow has `workflow_dispatch` trigger configured
- [ ] Repository and owner names are correct
- [ ] Branch/ref exists in the target repository

### Enable Debug Mode

Add debug logging to get more detailed information:

```yaml
- name: Enable debug logging
  run: echo "ACTIONS_STEP_DEBUG=true" >> $GITHUB_ENV

- name: Trigger workflow with debugging
  uses: PhuongTMR/workflow-trigwait@v1
  with:
    owner: my-org
    repo: my-repo
    github_token: ${{ secrets.GITHUB_TOKEN }}
    workflow_file_name: deploy.yml
```

## Common Errors

### 1. Timeout: Workflow Run Did Not Appear

**Error Message:**
```
timeout: workflow run did not appear within 120s
```

**Causes and Solutions:**

#### Cause 1: Concurrent Workflow Triggers (Race Condition)

Multiple workflows triggered at the same time can cause confusion in time-based matching.

**Solution:** Enable distinct ID correlation

```yaml
- uses: PhuongTMR/workflow-trigwait@v1
  with:
    owner: my-org
    repo: my-repo
    github_token: ${{ secrets.GITHUB_TOKEN }}
    workflow_file_name: deploy.yml
    distinct_id_name: correlation_id  # Add this
```

And update your target workflow:

```yaml
name: Deploy
run-name: Deploy [${{ inputs.correlation_id }}]

on:
  workflow_dispatch:
    inputs:
      correlation_id:
        required: false
        type: string
```

#### Cause 2: Workflow Takes Time to Appear in API

GitHub's API may have a delay before the workflow run appears.

**Solution:** Increase timeout

```yaml
trigger_timeout: 300  # Wait 5 minutes instead of default 2 minutes
```

#### Cause 3: Workflow File Name Mismatch

The workflow file name might be incorrect or not match exactly.

**Solution:** Verify the exact file name

```bash
# Check your workflow file name
ls .github/workflows/
```

```yaml
# Use exact file name including extension
workflow_file_name: deploy.yml  # NOT deploy or deploy.yaml
```

#### Cause 4: Branch Filtering in Target Workflow

Target workflow might have branch filters that exclude your ref.

**Check your target workflow:**
```yaml
on:
  workflow_dispatch:
  push:
    branches: [main, develop]  # ⚠️ Might prevent manual triggers on other branches
```

**Solution:** Ensure workflow_dispatch has no branch restrictions

```yaml
on:
  workflow_dispatch:  # ✅ No branch restrictions
    inputs: {}
  push:
    branches: [main]
```

### 2. API Request Failed: 403 Forbidden

**Error Message:**
```
API request failed: 403 Forbidden
```

**Causes and Solutions:**

#### Cause 1: Insufficient Token Permissions

**Solution:** Verify token permissions

For Personal Access Tokens (classic):
- ✅ `repo` (Full control of private repositories)
- ✅ `workflow` (Update GitHub Action workflows)

For Fine-grained Personal Access Tokens:
- ✅ Repository access (the target repository)
- ✅ Actions: Read and write
- ✅ Contents: Read

#### Cause 2: Using GITHUB_TOKEN for Cross-Repository Triggers

The default `GITHUB_TOKEN` cannot trigger workflows in other repositories.

**Solution:** Use a Personal Access Token

```yaml
# ❌ Won't work for cross-repo
github_token: ${{ secrets.GITHUB_TOKEN }}

# ✅ Use PAT for cross-repo
github_token: ${{ secrets.PERSONAL_ACCESS_TOKEN }}
```

#### Cause 3: Repository Settings Restrict Actions

**Solution:** Check repository settings

1. Go to repository Settings → Actions → General
2. Verify "Allow all actions and reusable workflows" is selected
3. Check "Fork pull request workflows from outside collaborators"

#### Cause 4: Organization Security Policies

Some organizations restrict workflow triggers.

**Solution:** Contact your organization admin to verify:
- Actions are enabled for the repository
- External actions are allowed
- Workflow permissions are properly configured

### 3. API Request Failed: 404 Not Found

**Error Message:**
```
API request failed: 404 Not Found
```

**Causes and Solutions:**

#### Cause 1: Repository Name or Owner Incorrect

**Solution:** Double-check the repository details

```yaml
# Verify on GitHub: https://github.com/{owner}/{repo}
owner: my-org        # Organization or user name
repo: my-repository  # Repository name (case-sensitive)
```

#### Cause 2: Workflow File Does Not Exist

**Solution:** Verify the workflow file exists

```bash
# Check file path
ls .github/workflows/deploy.yml
```

#### Cause 3: Token Does Not Have Access to Repository

**Solution:** Verify token has repository access

- For private repos: Token must have `repo` scope
- For public repos: Token needs at least `public_repo` scope
- Token owner must have access to the repository

### 4. Workflow Failed with Conclusion: failure

**Error Message:**
```
workflow failed with conclusion: failure
```

This means the triggered workflow ran but failed.

**Solutions:**

#### Solution 1: Disable Failure Propagation

Continue your workflow even if the triggered workflow fails:

```yaml
- uses: PhuongTMR/workflow-trigwait@v1
  with:
    owner: my-org
    repo: my-repo
    github_token: ${{ secrets.GITHUB_TOKEN }}
    workflow_file_name: deploy.yml
    propagate_failure: false  # Don't fail this job
```

#### Solution 2: Check Workflow Output

Access the triggered workflow to diagnose:

```yaml
- name: Trigger workflow
  id: trigger
  uses: PhuongTMR/workflow-trigwait@v1
  with:
    owner: my-org
    repo: my-repo
    github_token: ${{ secrets.GITHUB_TOKEN }}
    workflow_file_name: deploy.yml
    propagate_failure: false

- name: Report failure
  if: steps.trigger.outputs.conclusion != 'success'
  run: |
    echo "Workflow failed: ${{ steps.trigger.outputs.workflow_url }}"
    echo "Conclusion: ${{ steps.trigger.outputs.conclusion }}"
    exit 1
```

### 5. Invalid client_payload JSON

**Error Message:**
```
invalid client_payload JSON: invalid character...
```

**Causes and Solutions:**

#### Cause 1: Invalid JSON Syntax

**Solution:** Validate your JSON

```yaml
# ❌ Invalid JSON
client_payload: '{"key": value}'  # Missing quotes around value

# ✅ Valid JSON
client_payload: '{"key": "value"}'

# ✅ Using YAML multiline for complex payloads
client_payload: |
  {
    "environment": "production",
    "version": "1.2.3",
    "options": {
      "deploy": true
    }
  }
```

#### Cause 2: Special Characters Not Escaped

**Solution:** Escape special characters properly

```yaml
# Use JSON.stringify or escape manually
client_payload: '{"message": "Deploy \"production\" environment"}'
```

## Performance Issues

### Slow Workflow Detection

**Issue:** Takes too long to find the triggered workflow.

**Solutions:**

1. **Enable Distinct ID Correlation** (Recommended)
   ```yaml
   distinct_id_name: correlation_id
   ```

2. **Reduce Polling Interval** (for quick workflows)
   ```yaml
   wait_interval: 5  # Poll every 5 seconds
   ```

3. **Optimize Target Workflow**
   - Ensure workflow starts quickly
   - Avoid heavy checkout operations before workflow_dispatch input processing

### Rate Limiting

**Issue:** Too many API requests causing rate limits.

**Solutions:**

1. **Increase Polling Interval**
   ```yaml
   wait_interval: 30  # Poll less frequently
   ```

2. **Use Distinct ID Correlation**
   - Reduces need for aggressive polling
   - More reliable matching

3. **Batch Workflow Triggers**
   - Instead of triggering many workflows, consider combining into one

## Token and Permission Issues

### Token Expiration

**Issue:** Token expires during long-running workflows.

**Solution:** Use a token with longer validity:

```yaml
# Use GitHub App token (recommended) or PAT with longer expiration
github_token: ${{ secrets.GITHUB_APP_TOKEN }}
```

### Cross-Organization Access

**Issue:** Cannot trigger workflows in different organization.

**Solution:** Ensure token has access:

1. Create PAT from account with access to both organizations
2. Grant token access to target organization
3. Repository must allow external collaborators

### GitHub Enterprise Server

**Issue:** Custom GitHub Enterprise Server URL.

**Solution:** Action automatically detects `GITHUB_API_URL` and `GITHUB_SERVER_URL` environment variables. These are set by GitHub Actions.

For manual testing:
```bash
export GITHUB_API_URL="https://github.company.com/api/v3"
export GITHUB_SERVER_URL="https://github.company.com"
```

## Workflow Correlation Issues

### Distinct ID Not Working

**Issue:** Still getting timeout even with distinct_id_name set.

**Checklist:**

1. **Verify target workflow includes ID in run-name:**
   ```yaml
   run-name: Deploy [${{ inputs.correlation_id }}]
   ```

2. **Verify input name matches:**
   ```yaml
   # In trigger
   distinct_id_name: correlation_id

   # In target workflow
   inputs:
     correlation_id:  # Must match
       type: string
   ```

3. **Check that run-name is not empty:**
   ```yaml
   # Use default value if needed
   run-name: Deploy [${{ inputs.correlation_id || 'manual' }}]
   ```

### Distinct ID Not in Target Workflow

**Issue:** You enabled `distinct_id_name` but the target workflow doesn't include it in `run-name`.

**Behavior:** The action will NOT fall back to time-based matching. It will keep polling until timeout.

**Why:** This is intentional. If you explicitly enable distinct ID correlation, the action assumes you want reliable matching and won't risk matching the wrong workflow run.

**Solution:** Either:

1. **Add distinct_id to target workflow** (Recommended):
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

2. **Remove distinct_id_name from trigger** (Falls back to time-based):
   ```yaml
   - uses: PhuongTMR/workflow-trigwait@v1
     with:
       owner: my-org
       repo: my-repo
       github_token: ${{ secrets.GITHUB_TOKEN }}
       workflow_file_name: deploy.yml
       # distinct_id_name: correlation_id  # Remove this
   ```

**Testing:** Use the test workflow below to verify your setup works correctly.

### Multiple Workflows Match

**Issue:** Multiple workflow runs match the distinct ID.

**Solution:** Ensure distinct ID is truly unique:
- Don't reuse distinct IDs
- The action auto-generates unique IDs
- Only override if you have a specific need

## Debugging Tips

### 1. Test with Simple Workflow

Create a minimal test workflow to verify setup:

```yaml
# .github/workflows/test.yml
name: Test Workflow
run-name: Test [${{ inputs.test_id }}]

on:
  workflow_dispatch:
    inputs:
      test_id:
        type: string

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Test successful!"
```

Trigger it:
```yaml
- uses: PhuongTMR/workflow-trigwait@v1
  with:
    owner: my-org
    repo: my-repo
    github_token: ${{ secrets.GITHUB_TOKEN }}
    workflow_file_name: test.yml
    distinct_id_name: test_id
```

### 2. Check GitHub API Directly

Test the GitHub API manually:

```bash
# Trigger workflow
curl -X POST \
  -H "Authorization: Bearer $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/OWNER/REPO/actions/workflows/deploy.yml/dispatches \
  -d '{"ref":"main","inputs":{}}'

# Check workflow runs
curl -H "Authorization: Bearer $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/OWNER/REPO/actions/workflows/deploy.yml/runs
```

### 3. Verify Workflow Dispatch Configuration

Check your target workflow accepts workflow_dispatch:

```bash
# Get workflow details
curl -H "Authorization: Bearer $GITHUB_TOKEN" \
  https://api.github.com/repos/OWNER/REPO/actions/workflows/deploy.yml
```

Look for `"workflow_dispatch"` in the response.

### 4. Use GitHub CLI

Debug with GitHub CLI:

```bash
# Trigger workflow
gh workflow run deploy.yml -R owner/repo --ref main

# List recent runs
gh run list -R owner/repo --workflow=deploy.yml

# View run details
gh run view RUN_ID -R owner/repo
```

### 5. Check Workflow Run Logs

Access the triggered workflow's logs:

```yaml
- name: Trigger workflow
  id: trigger
  uses: PhuongTMR/workflow-trigwait@v1
  with:
    owner: my-org
    repo: my-repo
    github_token: ${{ secrets.GITHUB_TOKEN }}
    workflow_file_name: deploy.yml
    propagate_failure: false

- name: Show workflow URL
  run: |
    echo "Check logs at: ${{ steps.trigger.outputs.workflow_url }}"
```

### 6. Test Locally

Test the action binary locally:

```bash
# Build binary
go build -o workflow-trigwait ./cmd/main.go

# Test with environment variables
INPUT_OWNER="my-org" \
INPUT_REPO="my-repo" \
INPUT_GITHUB_TOKEN="ghp_xxxx" \
INPUT_WORKFLOW_FILE_NAME="test.yml" \
INPUT_REF="main" \
./workflow-trigwait
```

## Getting Help

If you're still experiencing issues:

1. **Check Existing Issues:** [GitHub Issues](https://github.com/PhuongTMR/workflow-trigwait/issues)
2. **Search Discussions:** [GitHub Discussions](https://github.com/PhuongTMR/workflow-trigwait/discussions)
3. **Create New Issue:** Include:
   - Error message (full output)
   - Workflow configuration (sanitized)
   - Target workflow configuration
   - Steps to reproduce
   - Expected vs actual behavior

### Issue Template

```markdown
**Describe the issue**
A clear description of what happened.

**Configuration**
```yaml
- uses: PhuongTMR/workflow-trigwait@v1
  with:
    owner: xxx
    repo: xxx
    # ... sanitized config
```

**Target Workflow**
```yaml
name: My Workflow
on:
  workflow_dispatch:
    # ... workflow config
```

**Error Output**
```
Error message here
```

**Expected Behavior**
What you expected to happen.

**Environment**
- Action version: v1.x.x
- GitHub: Cloud / Enterprise Server
- Runner: ubuntu-latest / self-hosted
```

## See Also

- [USAGE_GUIDE.md](USAGE_GUIDE.md) - Comprehensive usage examples
- [README.md](../README.md) - Quick start guide
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Contributing guidelines
