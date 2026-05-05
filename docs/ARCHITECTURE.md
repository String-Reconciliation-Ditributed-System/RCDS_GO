# RCDS_GO Architecture

RCDS_GO is organized as a small CLI plus reusable Go packages for set reconciliation, TCP transport, and file synchronization. The project is based on Bowen Song and Ari Trachtenberg's Allerton 2019 paper, ["Scalable String Reconciliation by Recursive Content-Dependent Shingling"](https://ieeexplore.ieee.org/document/8919901); see [PAPER.md](PAPER.md) for the research-to-code map.

## Layers

```text
cmd/
  main.go                 CLI entry point for set and file sync

pkg/file/
  sync.go                 Chunked file pull protocol with SHA-256 verification

pkg/lib/algorithm/
  full_sync/              Deterministic full-set reconciliation baseline
  iblt/                   Invertible Bloom Lookup Table reconciliation
  rcds/                   Recursive Content-Dependent Shingling adapter

pkg/lib/genSync/
  interface.go            Shared GenSync contract
  connection.go           Length-prefixed TCP connection implementation

pkg/set/
  set.go                  Hashable set abstraction used by algorithms

pkg/util/
  auxiliary.go            Stable integer and byte conversions
```

## CLI Flow

The CLI supports two modes:

- `--mode set`: reconcile elements supplied with `--items` and/or `--input`.
- `--mode file`: serve a source file and let a client pull an exact verified copy.

Set mode creates a `genSync.GenSync` implementation from `--algorithm`:

```text
rcds server/client
  -> parse flags
  -> create full, iblt, or rcds syncer
  -> add []byte elements
  -> run SyncServer or SyncClient over TCP
  -> optionally write sorted output
```

File mode uses `pkg/file` instead of the set algorithms:

```text
server source file
  -> fixed-size chunks
  -> content hashes and full SHA-256
  -> manifest plus missing chunk payloads

client destination file
  -> local chunk hashes, if the file exists
  -> request missing chunks
  -> reconstruct to temp file
  -> verify checksum
  -> atomically replace destination
```

## GenSync Contract

All set algorithms implement the same interface:

```go
type GenSync interface {
    SetFreezeLocal(freezeLocal bool)
    AddElement(elem interface{}) error
    DeleteElement(elem interface{}) error
    SyncClient(ip string, port int) error
    SyncServer(ip string, port int) error
    GetLocalSet() *set.Set
    GetSetAdditions() *set.Set
    GetSentBytes() int
    GetReceivedBytes() int
    GetTotalBytes() int
}
```

The current implementations expect byte elements (`[]byte`) for CLI and algorithm interoperability. The set package normalizes byte keys to strings so map keys remain hashable.

## Transport

`pkg/lib/genSync/connection.go` provides the TCP framing used by the set algorithms:

- `net.JoinHostPort` address construction
- port validation
- 8-byte length prefixes
- full writes via `io.Copy`
- bounded payload and slice sizes
- safe close before and after connection establishment

This transport is intentionally minimal. TLS, authentication, and multi-client concurrency are deployment concerns that should be added above or around this layer for production networks.

## Algorithms

### Full Sync

`pkg/lib/algorithm/full_sync` exchanges complete sets when digests differ. It is deterministic and useful as a correctness baseline.

### IBLT

`pkg/lib/algorithm/iblt` uses Invertible Bloom Lookup Tables to exchange compact set summaries. It performs best when `--expected-diff` is close to the actual symmetric difference and can retry with larger tables.

### RCDS

`pkg/lib/algorithm/rcds` implements the paper-inspired metadata pipeline:

```text
input bytes
  -> rolling content hashes
  -> local-minimum content-dependent chunks
  -> adjacent hash shingles
  -> dictionary-backed reconstruction metadata
```

The current adapter then uses a full-sync backend for the wire exchange. That is an intentional implementation boundary: it keeps the GenSync workflow compatible and tested while preserving a path toward a pure RCDS exchange that reconciles shingle differences and transfers only missing terminal partitions.

The paper's strongest fit is similar ordered data with small or clustered edits. It is less favorable when edits are numerous and sparsely distributed because many content-dependent partitions may be affected.

## File Protocol

The file protocol is exact and conservative:

1. The client sends hashes of chunks it already has.
2. The server sends a full manifest plus data for missing unique chunks.
3. The client reconstructs the file in manifest order.
4. The client verifies the full SHA-256 checksum.
5. The client renames the temporary file into place.

The current file implementation is designed for correctness and CLI usability. Very large files still need streaming and bounded-memory payload exchange before this should be used for multi-gigabyte production transfers.

## Testing Strategy

- Unit tests cover CLI parsing, file reconstruction, integer encoding, set operations, logger configuration, and algorithm edge cases.
- Package tests exercise local TCP sync for full sync, IBLT, and RCDS.
- Integration tests run all algorithms through the same reconciliation scenario.
- E2E tests build the binary and run real CLI server/client workflows for sets, files, large input files, startup, and connection failure.

## References

- B. Song and A. Trachtenberg, "Scalable String Reconciliation by Recursive Content-Dependent Shingling", 2019 57th Annual Allerton Conference on Communication, Control, and Computing, pp. 623-630. DOI: [10.1109/ALLERTON.2019.8919901](https://ieeexplore.ieee.org/document/8919901). Open preprint: [arXiv:1910.00536](https://arxiv.org/abs/1910.00536).
- Y. Minsky, A. Trachtenberg, and R. Zippel, "Set Reconciliation with Nearly Optimal Communication Complexity", IEEE Transactions on Information Theory, 2003.
- M. T. Goodrich and M. Mitzenmacher, "Invertible Bloom Lookup Tables", Allerton 2011.
