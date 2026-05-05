[![Go Report Card](https://goreportcard.com/badge/github.com/String-Reconciliation-Ditributed-System/RCDS_GO)](https://goreportcard.com/report/github.com/String-Reconciliation-Ditributed-System/RCDS_GO)
![Go CI](https://github.com/String-Reconciliation-Ditributed-System/RCDS_GO/workflows/Go%20CI/badge.svg)
[![codecov](https://codecov.io/gh/String-Reconciliation-Ditributed-System/RCDS_GO/branch/master/graph/badge.svg)](https://codecov.io/gh/String-Reconciliation-Ditributed-System/RCDS_GO)
[![GoDoc](https://godoc.org/github.com/String-Reconciliation-Ditributed-System/RCDS_GO?status.svg)](https://godoc.org/github.com/String-Reconciliation-Ditributed-System/RCDS_GO)

# Recursive Content-Dependent Shingling (RCDS)

RCDS_GO is a Go implementation and engineering sandbox for Recursive Content-Dependent Shingling, the string reconciliation protocol introduced by Bowen Song and Ari Trachtenberg in ["Scalable String Reconciliation by Recursive Content-Dependent Shingling"](https://ieeexplore.ieee.org/document/8919901), Allerton 2019. An open preprint is available as [arXiv:1910.00536](https://arxiv.org/abs/1910.00536).

Project website: [GitHub Pages](https://string-reconciliation-ditributed-system.github.io/RCDS_GO/)

## Overview

The paper frames string reconciliation as the problem of bringing similar remote strings into agreement while communicating the differences, not the whole object. RCDS reduces that problem to set reconciliation by turning strings into recursively partitioned, content-dependent shingles that can be reconciled and reconstructed.

This repository turns those ideas into reusable Go packages and a CLI. It includes:

- RCDS metadata construction through content-dependent chunking and hash shingles.
- IBLT and full-sync set reconciliation backends behind a shared GenSync interface.
- Exact chunked file pull workflows with SHA-256 verification.
- Tests, CI, Docker, Kubernetes scaffolding, and a GitHub Pages project website.

See [docs/PAPER.md](docs/PAPER.md) for the research-to-code map, current implementation boundaries, and citation details.

### Key Features

- **CLI server/client**: Real `rcds server` and `rcds client` commands for local and containerized workflows.
- **Multiple algorithms**: `rcds`, `iblt`, and `full` set reconciliation behind one GenSync-compatible interface.
- **Chunked file pull**: Transfers missing chunks and verifies the final SHA-256 checksum before replacing the destination.
- **TCP transport**: Length-prefixed payloads with bounded reads, full writes, and byte counters.
- **Tested workflows**: Unit, integration, e2e, coverage, and vet checks cover the critical paths.
- **Deployment scaffolding**: Dockerfile, Kubernetes CRD/RBAC manifests, and a GitHub Pages project website.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [CLI Reference](docs/CLI.md)
- [Paper Background](docs/PAPER.md)
- [Usage](#usage)
- [Architecture](#architecture)
- [Algorithms](#algorithms)
- [API Documentation](#api-documentation)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Contributing](#contributing)
- [References](#references)
- [License](#license)

## Installation

### Prerequisites

- Go 1.24 or later
- Make (optional, for using Makefile commands)

### Install from Source

```bash
git clone https://github.com/String-Reconciliation-Ditributed-System/RCDS_GO.git
cd RCDS_GO
make build
```

The binary will be available at `bin/rcds`.

### Install as a Library

```bash
go get github.com/String-Reconciliation-Ditributed-System/RCDS_GO
```

## Quick Start

### CLI Set Sync

Start a one-shot server with a line-delimited or comma-delimited set:

```bash
./bin/rcds server --algorithm full --items server-only,shared --output server.out
```

Then connect a client:

```bash
./bin/rcds client --algorithm full --items client-only,shared --output client.out
```

Both output files will contain the reconciled set:

```text
client-only
server-only
shared
```

### CLI File Sync

Serve an exact source file:

```bash
./bin/rcds server --mode file --file ./source.bin
```

Pull it from a client:

```bash
./bin/rcds client --mode file --file ./copy.bin
```

The file protocol transfers only chunks the client does not already have and verifies the final SHA-256 checksum before replacing the destination file.

### Library Usage

```go
package main

import (
    "log"
    "sync"

    "github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/algorithm/full_sync"
    "github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/genSync"
)

func main() {
    server, err := full_sync.NewFullSetSync()
    if err != nil {
        log.Fatal(err)
    }
    client, err := full_sync.NewFullSetSync()
    if err != nil {
        log.Fatal(err)
    }
    
    server.AddElement([]byte("server-only"))
    server.AddElement([]byte("shared"))
    client.AddElement([]byte("client-only"))
    client.AddElement([]byte("shared"))
    
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := server.SyncServer("127.0.0.1", 8080); err != nil {
            log.Fatal(err)
        }
    }()
    
    if err := client.SyncClient("127.0.0.1", 8080); err != nil {
        log.Fatal(err)
    }
    wg.Wait()

    var _ genSync.GenSync = client
}
```

## Usage

### Building the Project

```bash
# Build the binary
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Run linter
make lint

# Run all checks
make all

# Run integration and e2e tests
make integration-test
make e2e-test
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run CLI e2e tests
go test -tags e2e ./test/e2e

# Run integration tests
go test -tags integration ./test/integration
```

## Architecture

RCDS uses a layered architecture:

1. **Application Layer**: Command-line interface and user-facing APIs
2. **Reconciliation Layer**: RCDS, Full Sync, and IBLT implementations
3. **Core Libraries**: GenSync interface, hash functions, dictionaries
4. **Utilities**: Set operations, file utilities, type conversions

For detailed architecture documentation, see [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

For command-line details, see [docs/CLI.md](docs/CLI.md).

## Algorithms

RCDS supports multiple set reconciliation algorithms:

### RCDS (Recursive Content-Dependent Shingling)

The RCDS adapter follows the paper's front half: it builds content-dependent partitions, creates hash-shingle metadata, and keeps those structures compatible with the shared `GenSync` workflow. The current network exchange is intentionally conservative and uses the full-sync backend while the pure RCDS wire protocol is being hardened.

- **Paper goal**: sublinear communication for many similar strings, especially when edits are small or clustered.
- **Current implementation**: content-dependent metadata plus full-sync wire exchange.
- **Best for**: research, testing content-dependent shingling behavior, and comparing reconciliation backends.
- **Caution**: do not treat `--algorithm rcds` as a production rsync replacement yet.

### IBLT (Invertible Bloom Lookup Tables)

A probabilistic data structure for set reconciliation.

- **Complexity**: O(d) where d is the number of differences
- **Best for**: Sets with small symmetric difference
- **Use case**: Network-efficient reconciliation

### Full Sync

Traditional full synchronization (baseline for comparison).

- **Complexity**: O(n)
- **Best for**: Small datasets or complete synchronization
- **Use case**: Initial sync or fallback method

## API Documentation

### GenSync Interface

The core interface for all synchronization algorithms:

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

For complete API documentation, run:

```bash
godoc -http=:6060
```

Then visit http://localhost:6060/pkg/github.com/String-Reconciliation-Ditributed-System/RCDS_GO/

## Project Website

The website source lives in [docs/index.html](docs/index.html) and [docs/assets/site.css](docs/assets/site.css). GitHub Pages publishes the `docs/` directory on pushes to `master` or `main`.

Preview locally:

```bash
python3 -m http.server 8000 --directory docs
```

Then open http://127.0.0.1:8000.

## Kubernetes Deployment

The repository includes Kubernetes CRD/RBAC scaffolding.

### Installing the CRD

```bash
kubectl apply -f deploy/crds/
```

### Installing RBAC Scaffolding

```bash
kubectl apply -f deploy/operator/
```

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for detailed deployment instructions.

Note: a complete production controller Deployment is not included yet.

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linters
5. Submit a pull request

## References

If you use this work, please cite the relevant papers:

**[1]** B. Song and A. Trachtenberg, "Scalable String Reconciliation by Recursive Content-Dependent Shingling"
2019 57th Annual Allerton Conference on Communication, Control, and Computing, pp. 623-630.
DOI: [10.1109/ALLERTON.2019.8919901](https://ieeexplore.ieee.org/document/8919901). Open preprint: [arXiv:1910.00536](https://arxiv.org/abs/1910.00536).

**[2]** Y. Minsky, A. Trachtenberg, and R. Zippel,  
"Set Reconciliation with Nearly Optimal Communication Complexity",  
IEEE Transactions on Information Theory, 49:9.  
<http://ipsit.bu.edu/documents/ieee-it3-web.pdf>

**[3]** Y. Minsky and A. Trachtenberg,  
"Scalable set reconciliation"  
40th Annual Allerton Conference on Communication, Control, and Computing, 2002.  
<http://ipsit.bu.edu/documents/BUTR2002-01.pdf>

**[4]** Goodrich, Michael T., and Michael Mitzenmacher. "Invertible bloom lookup tables."  
49th Annual Allerton Conference on Communication, Control, and Computing (Allerton), 2011.  
[arXiv](https://arxiv.org/abs/1101.2245)

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

This implementation is based on the [cpisync](https://github.com/trachten/cpisync) project. The original C++ implementation is available at [forked cpisync](https://github.com/Bowenislandsong/cpisync).

## Contact

For questions, issues, or contributions, please open an issue on GitHub.

---

**Note**: This is an active research project. APIs may change as the project evolves.
