[![Go Report Card](https://goreportcard.com/badge/github.com/String-Reconciliation-Ditributed-System/RCDS_GO)](https://goreportcard.com/report/github.com/String-Reconciliation-Ditributed-System/RCDS_GO)
![Go CI](https://github.com/String-Reconciliation-Ditributed-System/RCDS_GO/workflows/Go%20CI/badge.svg)
[![codecov](https://codecov.io/gh/String-Reconciliation-Ditributed-System/RCDS_GO/branch/master/graph/badge.svg)](https://codecov.io/gh/String-Reconciliation-Ditributed-System/RCDS_GO)
[![GoDoc](https://godoc.org/github.com/String-Reconciliation-Ditributed-System/RCDS_GO?status.svg)](https://godoc.org/github.com/String-Reconciliation-Ditributed-System/RCDS_GO)

# Recursive Content-Dependent Shingling (RCDS)

RCDS is a scalable string reconciliation protocol designed for distributed systems. This Go implementation provides efficient file synchronization using set reconciliation primitives.

## Overview

The RCDS algorithm breaks files into content-dependent chunks (shingles) and uses set reconciliation to synchronize data between distributed nodes. This approach is significantly more efficient than traditional file synchronization methods, especially for large files with small differences.

### Key Features

- 🚀 **Scalable**: Logarithmic complexity with respect to file size
- 🔒 **Efficient**: Only transfers differences, not entire files
- 🔄 **Multiple Algorithms**: Supports CPI, Interactive CPI, and IBLT set reconciliation
- 🌐 **Distributed**: Designed for distributed systems
- 📦 **Go Modules**: Native Go module support
- ☸️ **Kubernetes Ready**: CRD support for Kubernetes deployments

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
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

- Go 1.21 or later
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

### Basic Usage

```go
package main

import (
    "github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/genSync"
    "github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/set"
)

func main() {
    // Create a new sync instance
    sync := // ... initialize your sync algorithm
    
    // Add elements to sync
    sync.AddElement("data1")
    sync.AddElement("data2")
    
    // Start server
    go sync.SyncServer("127.0.0.1", 8080)
    
    // Connect as client
    sync.SyncClient("127.0.0.1", 8080)
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
```

### Running Tests

```bash
# Run all tests
go test ./pkg/...

# Run with verbose output
go test -v ./pkg/...

# Run with coverage
go test -coverprofile=coverage.out ./pkg/...
go tool cover -html=coverage.out
```

## Architecture

RCDS uses a layered architecture:

1. **Application Layer**: Command-line interface and user-facing APIs
2. **Reconciliation Layer**: RCDS, Full Sync, and IBLT implementations
3. **Core Libraries**: GenSync interface, hash functions, dictionaries
4. **Utilities**: Set operations, file utilities, type conversions

For detailed architecture documentation, see [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## Algorithms

RCDS supports multiple set reconciliation algorithms:

### RCDS (Recursive Content-Dependent Shingling)

The main algorithm that uses content-dependent chunking and hash shingling.

- **Complexity**: O(log n) with respect to file size
- **Best for**: Large files with small differences
- **Use case**: File synchronization in distributed systems

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

## Kubernetes Deployment

RCDS can be deployed on Kubernetes using Custom Resource Definitions (CRDs).

### Installing the CRD

```bash
kubectl apply -f deploy/crds/
```

### Deploying RCDS

```bash
kubectl apply -f deploy/operator.yaml
```

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for detailed deployment instructions.

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
57th Annual Allerton Conference on Communication, Control, and Computing, 2019  
[(Allerton)](https://proceedings.allerton.csl.illinois.edu/media/files/0073.pdf)

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
