# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Real `rcds server` and `rcds client` CLI workflows for set sync.
- Chunked file pull mode with SHA-256 verification and atomic destination replacement.
- GitHub Pages project website under `docs/`.
- CLI reference documentation.
- GA readiness documentation covering blockers, execution order, and verification gates.
- Unit, integration, and e2e coverage for CLI, file sync, algorithms, transport, set operations, and utility conversions.

### Changed
- `make test` now runs all packages.
- CI uses Go 1.24 and runs the expanded package test suite.
- E2E and integration workflows now fail when tests fail.
- Deployment and architecture documentation now reflects the implemented CLI and file sync behavior.
- Release tag automation now gates artifact builds behind formatting, vet, race, integration, e2e, govulncheck, and gosec checks.

### Fixed
- Removed unresolved RCDS merge-conflict markers.
- Hardened TCP framing, payload bounds, close behavior, and byte counters.
- Added TCP operation deadlines for active set-sync sessions and wired CLI `--timeout` through set algorithms.
- Replaced unsafe integer conversion helpers with stable big-endian encoding.
- Made set key normalization nil-safe.
- Made IBLT reject non-byte elements instead of panicking.
- Made file manifest reconstruction reject chunks whose bytes do not match their declared hash.
- Made unavailable crypto hash functions return errors instead of panicking.
- Made invalid IBLT hash options fail during syncer construction.
- Made IBLT parameter mismatches return explicit client and server errors.
- Made IBLT skip unnecessary resync when initial decode succeeds with zero retries configured.
- Made IBLT sync tests use dynamic ports and correct randomized symmetric-difference sizing.
- Made set-sync listeners honor configured timeouts while waiting for the first client.
- Made lower-level TCP connection construction reject negative timeouts.
- Made content-dependent partitioning reject zero hash-space values.
- Made set difference preserve values from the left-hand set.
- Made set intersection preserve values from the left-hand set.
- Made file sync reject manifests whose declared file size does not match reconstructed manifest bytes.
- Made file sync reject negative chunk sizes before writing.
- Reduced file chunk-read allocations by reusing the read buffer and pre-sizing chunk metadata when file size is known.
- Applied log level validation before logger construction.
- Removed runtime logrus usage, unused controller-runtime/zap logging dependencies, and the direct client-go retry dependency from runtime paths.
- Removed the stale Travis CI release configuration so GitHub Actions is the release path of record.

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
