# Release Notes

## Current Unreleased

### Major Changes

- Real CLI server/client workflows for set reconciliation.
- File sync mode with chunk reuse, SHA-256 verification, and atomic replacement.
- Static project website under `docs/`, published by GitHub Pages.
- CLI reference documentation in `docs/CLI.md`.
- Expanded unit, integration, and e2e tests.
- CI now uses Go 1.24 and fails build-tagged test workflows on failures.
- GA readiness documentation in `docs/GA_READINESS.md`.

### Fixes

- File sync rejects manifest chunks when the bytes do not match the declared chunk hash.
- File sync rejects manifests whose declared file size does not match the reconstructed manifest bytes.
- IBLT parameter mismatches now return explicit client/server errors.
- Active set-sync sessions now honor CLI `--timeout` through the shared GenSync transport.
- Set-sync listeners now honor configured timeouts while waiting for the first client.
- Invalid or unavailable crypto hash selections now fail cleanly.
- RCDS content-dependent partitioning rejects zero hash-space values.
- Set difference now preserves values from the left-hand set.
- Set intersection now preserves values from the left-hand set.
- File chunk reads allocate less memory by reusing the read buffer.
- File chunk metadata allocation is pre-sized when file size is known.
- Runtime logrus usage, unused controller-runtime/zap logging dependencies, and the direct client-go retry dependency were removed.
- IBLT tests avoid fixed ports and correct randomized symmetric-difference sizing.
- Removed the stale Travis CI release configuration.
- Release tags now run fmt, vet, race, integration, e2e, govulncheck, and gosec gates before artifact builds.

### Verification

```bash
go vet ./...
go test -race ./...
go test -tags integration ./test/integration -v
go test -tags e2e ./test/e2e -v
```

GA release status and remaining blockers are tracked in `docs/GA_READINESS.md`.

## Version 0.2.0

### Major Changes

#### Modernization
- Updated Go version from 1.14 to 1.21
- Modernized all dependencies
- Updated GitHub Actions to latest versions
- Removed vendored dependencies in favor of Go modules

#### Documentation
- **NEW**: Comprehensive README with usage examples and installation instructions
- **NEW**: CONTRIBUTING.md guide for contributors
- **NEW**: Architecture documentation (docs/ARCHITECTURE.md)
- **NEW**: Deployment guide (docs/DEPLOYMENT.md) covering standalone, Docker, and Kubernetes
- **NEW**: API documentation with godoc

#### Kubernetes Support
- **NEW**: Custom Resource Definition (CRD) for Kubernetes deployment
- **NEW**: RBAC manifests (ServiceAccount, Role, RoleBinding)
- **NEW**: Example Kubernetes resources
- **NEW**: Dockerfile for containerization

#### CI/CD Improvements
- **NEW**: Comprehensive CI workflow with separate lint, test, and build jobs
- **NEW**: E2E testing workflow
- **NEW**: Security scanning workflow (gosec, govulncheck)
- **NEW**: Release automation workflow with multi-platform builds
- **NEW**: GitHub Pages deployment for documentation
- **NEW**: Code coverage reporting with codecov integration
- Added golangci-lint configuration for consistent code quality

#### Build System
- Enhanced Makefile with comprehensive targets:
  - `make build` - Build the binary
  - `make test` - Run all tests
  - `make test-coverage` - Run tests with coverage report
  - `make fmt` - Format code
  - `make vet` - Run go vet
  - `make lint` - Run golangci-lint
  - `make clean` - Clean build artifacts

### Breaking Changes

- Go 1.14 is no longer supported; minimum version is now Go 1.24
- Vendor directory removed; use `go mod download` instead

### Security

- All GitHub Actions workflows now have explicit permission blocks
- Added security scanning to CI/CD pipeline
- No known security vulnerabilities

### Bug Fixes

- Fixed Makefile vendor target
- Updated .gitignore to exclude vendor directory

### Known Limitations

- Kubernetes operator/controller not yet implemented (planned for v0.3.0)
- E2E tests are now implemented for CLI workflows in the unreleased branch
- Integration tests are now implemented for reconciliation workflows in the unreleased branch

### Migration Guide

#### From v0.0.1 to v0.2.0

1. Update Go version:
   ```bash
   # Ensure you have Go 1.24 or later
   go version
   ```

2. Remove vendor directory:
   ```bash
   rm -rf vendor/
   ```

3. Update dependencies:
   ```bash
   go mod tidy
   go mod download
   ```

4. Rebuild:
   ```bash
   make build
   ```

### Installation

#### Binary Installation

Download the appropriate binary for your platform from the [releases page](https://github.com/String-Reconciliation-Ditributed-System/RCDS_GO/releases).

**Linux (amd64)**:
```bash
wget https://github.com/String-Reconciliation-Ditributed-System/RCDS_GO/releases/download/v0.2.0/rcds-linux-amd64
chmod +x rcds-linux-amd64
sudo mv rcds-linux-amd64 /usr/local/bin/rcds
```

**macOS (amd64)**:
```bash
wget https://github.com/String-Reconciliation-Ditributed-System/RCDS_GO/releases/download/v0.2.0/rcds-darwin-amd64
chmod +x rcds-darwin-amd64
sudo mv rcds-darwin-amd64 /usr/local/bin/rcds
```

**macOS (arm64)**:
```bash
wget https://github.com/String-Reconciliation-Ditributed-System/RCDS_GO/releases/download/v0.2.0/rcds-darwin-arm64
chmod +x rcds-darwin-arm64
sudo mv rcds-darwin-arm64 /usr/local/bin/rcds
```

**Windows (amd64)**:
Download `rcds-windows-amd64.exe` and add to your PATH.

#### Docker Installation

```bash
docker pull rcds/rcds:0.2.0
docker pull rcds/rcds:latest
```

#### Go Library

```bash
go get github.com/String-Reconciliation-Ditributed-System/RCDS_GO@v0.2.0
```

### Contributors

Special thanks to all contributors who helped with this release!

### What's Next?

Version 0.3.0 will focus on:
- Kubernetes operator implementation
- Integration tests
- Enhanced E2E tests
- Performance optimizations
- Additional documentation and examples

---

For full changelog, see [CHANGELOG.md](CHANGELOG.md)
