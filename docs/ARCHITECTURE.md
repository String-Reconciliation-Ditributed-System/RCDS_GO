# RCDS_GO Architecture

## Overview

RCDS (Recursive Content-Dependent Shingling) is a scalable string reconciliation protocol designed for distributed systems. This document describes the architecture of the RCDS_GO implementation.

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                        │
│                    (cmd/root.go)                             │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                  Reconciliation Layer                        │
│  ┌────────────┐  ┌──────────────┐  ┌─────────────────┐     │
│  │  RCDS Sync │  │  Full Sync   │  │   IBLT Sync     │     │
│  │ (pkg/lib/  │  │ (pkg/lib/    │  │  (pkg/lib/      │     │
│  │ algorithm/ │  │ algorithm/   │  │  algorithm/     │     │
│  │ rcds/)     │  │ full_sync/)  │  │  iblt/)         │     │
│  └────────────┘  └──────────────┘  └─────────────────┘     │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                    Core Libraries                            │
│  ┌────────────┐  ┌──────────────┐  ┌─────────────────┐     │
│  │  GenSync   │  │  Dictionary  │  │  Hash Functions │     │
│  │ (pkg/lib/  │  │ (pkg/lib/    │  │  (pkg/lib/      │     │
│  │ genSync/)  │  │ algorithm/)  │  │  algorithm/)    │     │
│  └────────────┘  └──────────────┘  └─────────────────┘     │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                  Utilities Layer                             │
│  ┌────────────┐  ┌──────────────┐  ┌─────────────────┐     │
│  │   Set      │  │  File Utils  │  │  Conversion     │     │
│  │ (pkg/set/) │  │ (pkg/file/)  │  │  (pkg/util/)    │     │
│  └────────────┘  └──────────────┘  └─────────────────┘     │
└─────────────────────────────────────────────────────────────┘
```

## Key Components

### 1. RCDS Algorithm (`pkg/lib/algorithm/rcds/`)

The core RCDS implementation that performs recursive content-dependent shingling:

- **Content-Dependent Partitioning**: Breaks data into chunks based on content
- **Hash Shingling**: Creates fingerprints of data chunks
- **Backtracking**: Reconstructs data from reconciled shingles

### 2. Set Reconciliation Primitives

Multiple set reconciliation algorithms are supported:

- **CPI (CPISync)**: Characteristic Polynomial Interpolation
- **Interactive CPI**: Interactive version of CPI
- **IBLT**: Invertible Bloom Lookup Tables

### 3. GenSync Interface (`pkg/lib/genSync/`)

Generic synchronization interface that abstracts:

- Connection management (TCP)
- Data type conversions
- Common sync operations

### 4. Network Layer

- TCP-based communication
- Server/Client model
- Binary protocol for efficient data transfer

## Data Flow

### Synchronization Process

1. **Client Initialization**
   - Load local data
   - Create shingles using content-dependent partitioning
   - Build hash shingle set

2. **Connection Establishment**
   - Client connects to server
   - Server starts listening

3. **Set Reconciliation**
   - Exchange set metadata
   - Use IBLT/CPI to find differences
   - Transfer only missing elements

4. **Data Reconstruction**
   - Backtrack from shingles to original data
   - Verify integrity
   - Update local copy

## Design Patterns

### Interface-Based Design

The `GenSync` interface allows different reconciliation algorithms to be used interchangeably:

```go
type GenSync interface {
    SyncClient(ip string, port int) error
    SyncServer(ip string, port int) error
    AddElement(elem interface{}) error
    DeleteElement(elem interface{}) error
    GetLocalSet() *set.Set
}
```

### Content-Dependent Shingling

Uses rolling hash to create content-dependent boundaries:

```
Input: "The quick brown fox jumps over the lazy dog"
       ↓
Shingling: ["The quick", "brown fox", "jumps over", "the lazy dog"]
       ↓
Hashing: [hash1, hash2, hash3, hash4]
```

## Performance Considerations

1. **Scalability**: RCDS scales logarithmically with file size
2. **Network Efficiency**: Only sends differences, not entire files
3. **Memory Usage**: Uses bloom filters and IBLT for space efficiency
4. **Hash Functions**: Multiple hash functions supported for flexibility

## References

- Song, B., & Trachtenberg, A. (2019). "Scalable String Reconciliation by Recursive Content-Dependent Shingling"
- Minsky, Y., Trachtenberg, A., & Zippel, R. (2003). "Set Reconciliation with Nearly Optimal Communication Complexity"
- Goodrich, M. T., & Mitzenmacher, M. (2011). "Invertible Bloom Lookup Tables"
