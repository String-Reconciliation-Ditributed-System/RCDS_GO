# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2025-11-21

### Added
- Comprehensive README with usage examples and badges
- CONTRIBUTING.md guide for contributors
- Architecture documentation (docs/ARCHITECTURE.md)
- Deployment guide (docs/DEPLOYMENT.md)
- Kubernetes CRD for RCDS resources
- RBAC manifests for Kubernetes deployment
- Dockerfile for containerization
- GitHub Pages deployment workflow
- E2E testing workflow
- Security scanning workflow (gosec, govulncheck, dependency-review)
- Release automation workflow with multi-platform builds
- golangci-lint configuration
- Enhanced Makefile with comprehensive targets
- Code coverage reporting with codecov

### Changed
- Updated Go version from 1.14 to 1.21
- Modernized GitHub Actions workflows
- Updated all dependencies to latest versions
- Improved CI/CD pipeline with separate lint, test, and build jobs

### Removed
- Vendor directory (now using Go modules directly)
- Travis CI configuration (replaced with GitHub Actions)

### Fixed
- Makefile vendor target
- GitHub Actions workflow permissions (security fix)

### Security
- Added explicit permission blocks to all workflows
- Integrated security scanning tools

## [0.0.1] - 2019-10-XX

### Added
- Initial implementation of RCDS algorithm
- Support for IBLT, CPI, and Full Sync algorithms
- Basic TCP client/server implementation
- Unit tests for core functionality

[Unreleased]: https://github.com/String-Reconciliation-Ditributed-System/RCDS_GO/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/String-Reconciliation-Ditributed-System/RCDS_GO/compare/v0.0.1...v0.2.0
[0.0.1]: https://github.com/String-Reconciliation-Ditributed-System/RCDS_GO/releases/tag/v0.0.1
