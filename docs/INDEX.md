# Documentation Index

Welcome to the Workflow Trigger and Wait documentation!

## Getting Started

- **[README.md](../README.md)** - Start here for a quick overview and basic examples
- **[Usage Guide](USAGE_GUIDE.md)** - Comprehensive usage patterns and advanced examples

## Reference Documentation

### For Users

- **[Troubleshooting Guide](TROUBLESHOOTING.md)** - Diagnose and resolve common issues
  - Common errors and solutions
  - Performance optimization
  - Token and permission problems
  - Debugging tips

### For Developers

- **[Contributing Guide](../CONTRIBUTING.md)** - How to contribute to the project
  - Development setup
  - Code style and conventions
  - Testing procedures
  - Pull request process

## Quick Navigation

### Common Tasks

| Task | Documentation |
|------|---------------|
| First time setup | [README - Quick Start](../README.md) |
| Trigger cross-repo workflow | [Usage Guide - Cross-Repository](USAGE_GUIDE.md#1-cross-repository-deployment-pipeline) |
| Handle workflow failures | [Usage Guide - Failure Propagation](USAGE_GUIDE.md#5-conditional-failure-propagation) |
| Fix timeout errors | [Troubleshooting - Timeout](TROUBLESHOOTING.md#1-timeout-workflow-run-did-not-appear) |
| Fix permission errors | [Troubleshooting - 403 Forbidden](TROUBLESHOOTING.md#2-api-request-failed-403-forbidden) |
| Enable correlation | [README - Workflow Correlation](../README.md#workflow-correlation-optional) |
| Build from source | [Contributing - Development Setup](../CONTRIBUTING.md#development-setup) |

### By Use Case

| Use Case | Example |
|----------|---------|
| Basic workflow trigger | [README - Basic Usage](../README.md#basic-usage) |
| Parallel workflows | [Usage Guide - Parallel Execution](USAGE_GUIDE.md#2-parallel-workflow-execution) |
| Multi-environment deployment | [Usage Guide - Multi-Environment](USAGE_GUIDE.md#3-multi-environment-deployment) |
| Fire-and-forget | [Usage Guide - Fire-and-Forget](USAGE_GUIDE.md#4-fire-and-forget-workflow-trigger) |
| Conditional logic | [Usage Guide - Output Usage](USAGE_GUIDE.md#conditional-logic-based-on-results) |

## Documentation Structure

```
workflow-trigwait/
├── README.md                  # Overview and quick start
├── CONTRIBUTING.md            # Contribution guidelines
├── LICENSE                    # License information
└── docs/
    ├── INDEX.md             # This file
    ├── USAGE_GUIDE.md       # Comprehensive usage guide
    └── TROUBLESHOOTING.md   # Troubleshooting guide
```

## Getting Help

1. **Search the documentation** - Use Ctrl+F to search within docs
2. **Check existing issues** - [GitHub Issues](https://github.com/PhuongTMR/workflow-trigwait/issues)
3. **Ask questions** - [GitHub Discussions](https://github.com/PhuongTMR/workflow-trigwait/discussions)
4. **Report bugs** - [Create an issue](https://github.com/PhuongTMR/workflow-trigwait/issues/new)

## Contributing to Documentation

Documentation improvements are welcome! See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

When contributing documentation:
- Keep examples clear and concise
- Test all code examples
- Follow existing formatting
- Update this index if adding new docs
