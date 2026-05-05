# Paper Background

RCDS_GO is based on the paper:

> Bowen Song and Ari Trachtenberg, "Scalable String Reconciliation by Recursive Content-Dependent Shingling", 2019 57th Annual Allerton Conference on Communication, Control, and Computing, pp. 623-630.

- IEEE Xplore: <https://ieeexplore.ieee.org/document/8919901>
- DOI: `10.1109/ALLERTON.2019.8919901`
- Open preprint: <https://arxiv.org/abs/1910.00536>

## Problem

String reconciliation starts with two hosts that hold similar but not identical ordered data. The goal is to make the strings agree while sending as little data as possible. This is useful for distributed file systems, cloud storage clients, repository synchronization, and other systems where disconnected clients may make small edits to otherwise similar data.

Full sync sends the whole remote object. That is simple and deterministic, but its communication cost grows with input size even when the actual edit is tiny. Rsync improves this by matching fixed-size blocks with rolling checksums, but the paper shows cases where its communication still scales with the input length.

## Core RCDS Idea

The paper reduces string reconciliation to set reconciliation:

1. Partition each string by content, not by fixed offsets.
2. Recursively repeat that partitioning so long strings become a tree of content-dependent chunks.
3. Represent neighboring chunks as hash shingles with enough ordering information to reconstruct the string.
4. Reconcile the shingle sets using a set reconciliation protocol.
5. Exchange any missing terminal partitions and reconstruct the target string.

Content-dependent partitioning is the key. When an insertion or deletion shifts bytes, fixed offsets become misaligned. Content-derived cut points are more likely to remain aligned around unchanged regions, which keeps the symmetric difference closer to the edit distance.

## What The Paper Claims

The paper's contribution is a protocol that combines low communication for many similar strings with scalable computation for long strings.

Important results and caveats:

- RCDS is designed for similar strings, especially workloads with small or clustered edits.
- The paper reports sublinear communication growth for single-file experiments as input length increases under fixed burst edits.
- In experiments over the 5000 top-starred GitHub repositories available on April 17, 2019, RCDS used less communication than rsync for 51% of tested repository updates.
- RCDS can lose to rsync when updates contain many sparsely clustered edits.
- Parameters matter: partition depth, rolling window size, hash space, and the set reconciliation backend affect both communication and CPU cost.

## How This Repository Maps To The Paper

The repository has three related pieces:

- `pkg/lib/algorithm/rcds` implements content-dependent chunking and hash-shingle metadata inspired by the paper.
- `pkg/lib/algorithm/iblt` provides a probabilistic set reconciliation backend.
- `pkg/lib/algorithm/full_sync` provides a deterministic correctness baseline.

The current `rcds` adapter prepares the RCDS metadata, but it still uses the shared GenSync/full-sync backend for the wire exchange. That makes the CLI stable and testable while preserving a clear path toward a pure RCDS reconciliation exchange.

## Current Implementation Boundaries

Use this repository as a working implementation and comparison harness, not as a claim that every paper optimization is complete.

Implemented today:

- Content-dependent chunking with local-minimum cut points.
- Hash-shingle construction over chunk sequences.
- Shared `GenSync` interface for RCDS, IBLT, and full sync.
- CLI server/client flows for set reconciliation.
- Exact chunked file pull with checksum verification.

Still future work:

- A pure RCDS wire protocol that exchanges only reconciled hash-shingle differences and missing terminal partitions.
- Direct CLI controls for RCDS partition parameters.
- Benchmark harnesses that reproduce the paper's rsync and repository experiments.
- Streaming payload exchange for very large files.
- TLS, authentication, and multi-client production serving.

## Choosing An Algorithm

Use `--algorithm full` when correctness and simplicity matter more than bandwidth.

Use `--algorithm iblt` when you are reconciling sets and can estimate the symmetric difference with `--expected-diff`.

Use `--algorithm rcds` when you want to exercise the content-dependent shingling path or build on the paper's approach. Treat it as research-grade until the pure RCDS wire exchange is complete.

## Citation

```bibtex
@inproceedings{song2019scalable,
  title={Scalable String Reconciliation by Recursive Content-Dependent Shingling},
  author={Song, Bowen and Trachtenberg, Ari},
  booktitle={2019 57th Annual Allerton Conference on Communication, Control, and Computing (Allerton)},
  pages={623--630},
  year={2019},
  doi={10.1109/ALLERTON.2019.8919901}
}
```
