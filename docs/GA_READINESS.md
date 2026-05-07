# GA Readiness Plan

This repository is not GA-ready yet. The core CLI and library paths are useful for MVP validation, but a GA release needs the reliability, security, CI, and operational work below to be complete and verified.

## Current Supported Scope

Treat the current code as a GA candidate only for a narrow, controlled scope:

- CLI set reconciliation with `full`.
- CLI set reconciliation with `iblt` when both peers use matching options and `--expected-diff` is sized for the actual symmetric difference.
- Exact file pull sync for files that fit comfortably within process memory.
- Local binary and container smoke-test deployments inside trusted networks.

Do not claim GA support yet for public internet exposure, a complete low-bandwidth RCDS wire protocol, multi-gigabyte streaming file sync, resumable transfers, concurrent mutation of a shared syncer, or production Kubernetes operator behavior.

## Current Blockers

1. **Security boundary is incomplete.** Set sync and file sync use plaintext TCP with no authentication, authorization, or rate limiting. GA deployments must add TLS or require a documented trusted-network wrapper.
2. **Set server lifecycle is not fully context-aware.** Set-sync listeners now honor timeout deadlines before accepting a client, but the set server accept loop still does not receive a cancellation context. GA should make signal-triggered listener shutdown explicit and test signal/timeout behavior end to end.
3. **Large file sync is not streaming.** File sync reads and JSON-encodes chunks in memory. This is correct for small files, but it is not safe for multi-gigabyte production transfers.
4. **Configuration is CLI-only.** There is no config file or environment variable layer for repeatable production deployments.
5. **Observability is minimal.** There are byte counters, but no structured operation metrics, health checks, tracing, or error-rate reporting.
6. **Operational logging is sparse.** Runtime logrus usage has been removed and the local logger now uses standard-library `slog`, but the CLI and libraries still need one structured logging and metrics path that is actually wired through production commands.
7. **Kubernetes support is scaffolding only.** CRD and RBAC files exist, but the repo does not include a controller Deployment, Service, probes, or production configuration.
8. **RCDS wire protocol is research-grade.** `--algorithm rcds` builds RCDS metadata but currently uses the full-sync backend for exchange.
9. **Release governance must be enforced in GitHub settings.** Release tag automation now runs verification and security scans before building artifacts, but branch/tag protections and required status checks must be configured outside the repository files.
10. **CI must be the source of truth for full verification.** Local sandbox verification may not be able to bind TCP ports or download modules. The existing Go, e2e, integration, release, and security workflows must pass before release.

## Bugs Resolved In This Run

- Added chunk hash verification coverage so a malicious or corrupt file-sync response cannot satisfy a manifest with bytes that only match the declared size.
- Made crypto hashing return errors instead of panicking when an unavailable or invalid `crypto.Hash` is supplied, and linked the standard MD5/SHA1/SHA2 implementations used by current tests and options.
- Made IBLT hash-function option validation reject invalid hashes without panics.
- Added TCP operation deadlines to GenSync-backed set sync and wired CLI `--timeout` into `full`, `iblt`, and `rcds` set modes.
- Removed runtime `logrus` calls and the direct `client-go` retry dependency from the GenSync transport.
- Made IBLT's initial decode-success response skip resync even when `MaxSyncRetry` is zero.
- Made direct RCDS content hashing reject `hashSpace <= 0` before modulo arithmetic.
- Made set difference preserve values from the left-hand set.
- Made set intersection preserve values from the left-hand set.
- Reduced file chunk-read allocation overhead by reusing the read buffer.
- Tightened IBLT tests around dynamic ports, option mismatch errors, and symmetric-difference estimation.
- Added GenSync and full-sync timeout/payload-bound tests that compile locally and skip cleanly when localhost binding is blocked.
- Made CLI flag parsing reject negative timeout values.
- Removed the stale Travis CI release configuration so GitHub Actions is the release path of record.
- Made GenSync TCP listeners apply configured timeout deadlines while waiting for the first client and reject negative lower-level timeout values.
- Made file-sync clients reject response manifests whose declared `file_size` does not match the manifest byte count, and reject negative chunk sizes before writing.
- Reduced file chunk metadata allocations by preallocating the chunk slice from file size when available.
- Replaced the unused controller-runtime/zap logger dependency with standard-library `slog` and removed Kubernetes rand usage from algorithm tests.
- Added release-tag verification gates for formatting, vet, race tests, integration tests, e2e tests, `govulncheck`, and `gosec` before release artifacts are built.

## Execution Plan

1. **Stabilize correctness gates.**
   - Require `go test -race ./...`, `go vet ./...`, integration tests, and e2e tests in CI.
   - Remove flaky fixed-port tests and use dynamic localhost ports.
   - Keep protocol mismatch paths returning explicit errors instead of silently skipping.

2. **Harden network behavior.**
   - Add context-aware set server listener shutdown.
   - Keep `--timeout` wired through all set and file sync modes.
   - Add CI-backed tests for peer disconnects, protocol mismatch, unavailable server, and timeout behavior.

3. **Close security gaps.**
   - Add TLS configuration for server and client.
   - Add authentication or document a required authenticated proxy/service-mesh deployment.
   - Add payload and request-size limits for file protocol messages.

4. **Make production configuration repeatable.**
   - Add config file and environment variable support.
   - Define precedence as CLI flags > environment > config file > defaults.
   - Add `rcds config validate`.

5. **Bound file-sync memory use.**
   - Replace full-response JSON chunk payloads with a streaming chunk protocol.
   - Keep atomic destination replacement and SHA-256 verification.
   - Add benchmarks and memory ceilings for large-file transfers.

6. **Add operational visibility.**
   - Emit structured logs from one logging stack.
   - Add counters for sync attempts, bytes, duration, errors, and algorithm.
   - Add health/readiness probes for long-running deployments.

7. **Complete deployment artifacts.**
   - Add Docker Compose smoke tests.
   - Add Kubernetes Deployment, Service, probes, config, and documented rollout.
   - Decide whether the Kubernetes operator is in-scope for GA or explicitly post-GA.

8. **Clean release automation.**
   - Keep GitHub Actions as the release path of record and remove stale CI systems.
   - Keep release jobs requiring the same test, security, and smoke-test gates as PRs.
   - Configure branch and tag protections so required checks cannot be bypassed.

9. **Run release verification.**
   - Run the CI workflow on a clean branch.
   - Run manual CLI smoke tests for set and file sync.
   - Record benchmark baselines for full, IBLT, RCDS metadata, and file sync.
   - Update release notes with tested Go version, supported modes, and known limitations.

## Required Verification Matrix

| Area | Required command or check | GA expectation |
| --- | --- | --- |
| Formatting | `test -z "$(gofmt -l .)"` | No output |
| Vet | `go vet ./...` | Pass |
| Unit tests | `go test -race ./...` | Pass |
| Integration | `go test -tags integration ./test/integration -v` | Pass |
| E2E | `go test -tags e2e ./test/e2e -v` | Pass |
| Coverage | `go test ./... -coverprofile=coverage.out -covermode=atomic` | Critical paths covered |
| Performance | `go test ./... -bench=. -benchmem` | Baselines recorded |
| Security | TLS/auth review, dependency scan, payload limits | No unauthenticated production exposure |
| Deployment | Docker and Kubernetes smoke tests | Reproducible deploy and rollback |

## Current Local Verification Notes

The May 7, 2026 automation run could only execute packages that do not require uncached third-party modules. The sandbox blocks module downloads from `proxy.golang.org` and does not permit binding localhost TCP sockets, so full verification must run in CI or a developer environment with module and localhost access.

Commands run:

```bash
gofmt -l .
git diff --check
GOCACHE="$PWD/.cache/go-build" GOMODCACHE="$PWD/.cache/go-mod" go test ./log -count=1 -v
GOCACHE="$PWD/.cache/go-build" GOMODCACHE="$PWD/.cache/go-mod" go test ./pkg/file -run 'TestReadChunksAndWriteManifest|TestWriteManifestRejectsMissingChunk|TestWriteManifestRejectsSizeMismatch|TestWriteManifestRejectsChunkHashMismatch|TestReadChunksIfMissingUsesEmptyChecksum' -count=1
GOCACHE="$PWD/.cache/go-build" GOMODCACHE="$PWD/.cache/go-mod" go test ./pkg/lib/genSync ./pkg/file -count=1 -v
GOCACHE="$PWD/.cache/go-build" GOMODCACHE="$PWD/.cache/go-mod" go test ./pkg/lib/algorithm -count=1 -v
GOCACHE="$PWD/.cache/go-build" GOMODCACHE="$PWD/.cache/go-mod" go test ./pkg/lib/genSync -count=1 -v
GOCACHE="$PWD/.cache/go-build" GOMODCACHE="$PWD/.cache/go-mod" go test ./pkg/lib/algorithm/full_sync -count=1 -v
GOCACHE="$PWD/.cache/go-build" GOMODCACHE="$PWD/.cache/go-mod" go test ./pkg/set -count=1 -v
GOCACHE="$PWD/.cache/go-build" GOMODCACHE="$PWD/.cache/go-mod" go test ./pkg/util -count=1 -v
GOCACHE="$PWD/.cache/go-build" GOMODCACHE="$PWD/.cache/go-mod" go test ./pkg/file -count=1
GOCACHE="$PWD/.cache/go-build" GOMODCACHE="$PWD/.cache/go-mod" go test ./pkg/lib/algorithm ./pkg/lib/genSync ./pkg/lib/algorithm/full_sync ./pkg/set ./pkg/file ./pkg/util ./log -count=1
GOCACHE="$PWD/.cache/go-build" GOMODCACHE="$PWD/.cache/go-mod" go test -race ./pkg/file ./pkg/lib/algorithm ./pkg/lib/genSync ./pkg/lib/algorithm/full_sync ./pkg/set ./pkg/util ./log -count=1
GOCACHE="$PWD/.cache/go-build" GOMODCACHE="$PWD/.cache/go-mod" go test ./pkg/file ./pkg/lib/algorithm ./pkg/lib/genSync ./pkg/lib/algorithm/full_sync ./pkg/set ./pkg/util ./log -coverprofile=coverage.out -covermode=atomic -count=1
GOCACHE="$PWD/.cache/go-build" GOMODCACHE="$PWD/.cache/go-mod" go test ./pkg/file -bench=BenchmarkReadChunks1MiB -benchmem -run '^$' -count=3
GOCACHE="$PWD/.cache/go-build" GOMODCACHE="$PWD/.cache/go-mod" go vet ./pkg/file ./pkg/lib/algorithm ./pkg/lib/genSync ./pkg/lib/algorithm/full_sync ./pkg/set ./pkg/util ./log
GOCACHE="$PWD/.cache/go-build" GOMODCACHE="$PWD/.cache/go-mod" go test ./...
GOCACHE="$PWD/.cache/go-build" GOMODCACHE="$PWD/.cache/go-mod" go vet ./...
GOCACHE="$PWD/.cache/go-build" GOMODCACHE="$PWD/.cache/go-mod" go test -tags integration ./test/integration -v -count=1
GOCACHE="$PWD/.cache/go-build" GOMODCACHE="$PWD/.cache/go-mod" go test -tags e2e ./test/e2e -v -count=1
```

Results:

- `gofmt -l .` and `git diff --check` passed with no output.
- The targeted `pkg/file` and `pkg/lib/genSync` non-network regression subsets passed, including manifest size validation and listener timeout construction checks.
- The dependency-free `pkg/file`, `pkg/lib/algorithm`, `pkg/lib/genSync`, `pkg/lib/algorithm/full_sync`, `pkg/set`, `pkg/util`, and `log` suites passed under `-race`.
- `pkg/lib/genSync` and `pkg/lib/algorithm/full_sync` compile locally; localhost TCP cases skip when bind is not permitted by the environment.
- The latest local `BenchmarkReadChunks1MiB` baseline on darwin/arm64 Apple M4 is `750109-764066 ns/op`, `1372.36-1401.92 MB/s`, `1117641 B/op`, `57 allocs/op`.
- Targeted coverage for locally executable packages is `32.9%` overall, with package-level results of `pkg/file` 37.9%, `pkg/lib/algorithm` 71.4%, `pkg/lib/genSync` 14.6%, `pkg/lib/algorithm/full_sync` 0.0% because localhost paths skip in this sandbox, `pkg/set` 92.0%, `pkg/util` 88.9%, and `log` 100.0%.
- Targeted vet passed for the dependency-free package set.
- Full repo test, vet, integration, e2e, and build-tagged verification are blocked by missing module downloads because the sandbox cannot resolve `proxy.golang.org`; affected external modules are now limited to `github.com/stretchr/testify`, `github.com/SheldonZhong/go-IBLT`, and `github.com/emirpasic/gods`.
- Integration, e2e, and algorithm benchmark verification still need CI or a developer environment with module downloads and localhost TCP access.

Performance verification was only partially completed locally. Before GA, run `go test ./... -bench=. -benchmem` in CI or a developer environment and record baseline throughput, allocation count, and transferred byte counts for full sync, IBLT, RCDS metadata sync, and file sync.
