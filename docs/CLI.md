# RCDS CLI Reference

The `rcds` binary supports one-shot server/client workflows for set reconciliation and exact file pull synchronization.

## Build

```bash
make build
./bin/rcds version
```

## Commands

```bash
rcds server [options]
rcds client [options]
rcds version
rcds help
```

Servers handle one client and then exit by default. Pass `--once=false` to keep serving sequential clients.

## Common Options

| Flag | Default | Applies to | Description |
| --- | --- | --- | --- |
| `--host` | `127.0.0.1` | server, client | Address to listen on or connect to. |
| `--port` | `8080` | server, client | TCP port. |
| `--mode` | `set` | server, client | Sync mode: `set` or `file`. |
| `--algorithm` | `iblt` | set mode | Set reconciliation algorithm: `rcds`, `iblt`, or `full`. |
| `--items` | empty | set mode | Comma-separated set elements. |
| `--input` | empty | set mode, file server | Line-delimited set input, or file source in file server mode. |
| `--output` | empty | set mode, file client | Reconciled set output, or file destination in file client mode. |
| `--file` | empty | file mode | File source for server, destination for client. |
| `--expected-diff` | `100` | IBLT set mode | Expected symmetric difference for IBLT sizing. |
| `--max-retries` | `3` | IBLT set mode | Additional IBLT retry tables. |
| `--timeout` | `30s` | file mode | Network timeout for file sync. |
| `--chunk-size` | `65536` | file mode | Fixed chunk size in bytes. |
| `--freeze-local` | `false` | set mode | Do not apply remote additions locally. |
| `--once` | `true` | server | Exit after one client. |

## Set Sync

Start a server:

```bash
./bin/rcds server \
  --algorithm full \
  --port 8080 \
  --items server-only,shared \
  --output server.out
```

Connect a client:

```bash
./bin/rcds client \
  --algorithm full \
  --port 8080 \
  --items client-only,shared \
  --output client.out
```

Both outputs contain the reconciled set:

```text
client-only
server-only
shared
```

For larger inputs, use line-delimited files:

```bash
./bin/rcds server --algorithm iblt --expected-diff 500 --input server.txt --output server.out
./bin/rcds client --algorithm iblt --expected-diff 500 --input client.txt --output client.out
```

## File Sync

The file protocol is a pull workflow. The server exposes an exact source file. The client reconstructs the destination from chunks it already has plus chunks sent by the server. The destination is replaced only after the final SHA-256 checksum matches.

Start a file server:

```bash
./bin/rcds server --mode file --file ./source.bin --chunk-size 65536
```

Pull to a destination:

```bash
./bin/rcds client --mode file --file ./copy.bin --chunk-size 65536
```

You can also use `--input` for the server source and `--output` for the client destination:

```bash
./bin/rcds server --mode file --input ./source.bin
./bin/rcds client --mode file --output ./copy.bin
```

## Algorithm Notes

- `full` is deterministic and useful as a baseline or for small datasets.
- `iblt` is efficient when the expected symmetric difference is accurate enough.
- `rcds` prepares content-dependent metadata while preserving the GenSync transport interface.

## Verification

```bash
make test
make integration-test
make e2e-test
go vet ./...
```

`make integration-test` and `make e2e-test` use Go build tags and exercise localhost TCP workflows.
