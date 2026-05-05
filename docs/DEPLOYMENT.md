# RCDS Deployment Guide

This guide covers the deployment paths currently supported by the repository: local binaries, containers, GitHub Pages documentation, and Kubernetes manifests.

## Local Binary

Build from source:

```bash
git clone https://github.com/String-Reconciliation-Ditributed-System/RCDS_GO.git
cd RCDS_GO
make build
./bin/rcds version
```

Run a set sync server:

```bash
./bin/rcds server \
  --host 0.0.0.0 \
  --port 8080 \
  --algorithm iblt \
  --expected-diff 100 \
  --input server-set.txt \
  --output server-result.txt
```

Run a set sync client:

```bash
./bin/rcds client \
  --host 127.0.0.1 \
  --port 8080 \
  --algorithm iblt \
  --expected-diff 100 \
  --input client-set.txt \
  --output client-result.txt
```

Run a file sync server:

```bash
./bin/rcds server --mode file --host 0.0.0.0 --port 8080 --file ./source.bin
```

Run a file sync client:

```bash
./bin/rcds client --mode file --host 127.0.0.1 --port 8080 --file ./copy.bin
```

## Container Image

Build the image:

```bash
docker build -t rcds:latest .
```

Set sync server:

```bash
docker run --rm \
  --name rcds-server \
  -p 8080:8080 \
  -v "$PWD/data:/data" \
  rcds:latest server \
  --host 0.0.0.0 \
  --port 8080 \
  --algorithm full \
  --input /data/server-set.txt \
  --output /data/server-result.txt
```

File sync server:

```bash
docker run --rm \
  --name rcds-server \
  -p 8080:8080 \
  -v "$PWD/data:/data" \
  rcds:latest server \
  --mode file \
  --host 0.0.0.0 \
  --port 8080 \
  --file /data/source.bin
```

Client containers can use `--network host` on Linux, a shared Docker network, or a Compose service name depending on the environment.

## Docker Compose Example

```yaml
services:
  rcds-server:
    image: rcds:latest
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
    command:
      - server
      - --host
      - 0.0.0.0
      - --port
      - "8080"
      - --algorithm
      - full
      - --input
      - /data/server-set.txt
      - --output
      - /data/server-result.txt

  rcds-client:
    image: rcds:latest
    depends_on:
      - rcds-server
    volumes:
      - ./data:/data
    command:
      - client
      - --host
      - rcds-server
      - --port
      - "8080"
      - --algorithm
      - full
      - --input
      - /data/client-set.txt
      - --output
      - /data/client-result.txt
```

## Kubernetes Manifests

The repository currently ships:

- `deploy/crds/rcds_v1_rcds_crd.yaml`
- `deploy/examples/rcds_sample.yaml`
- RBAC manifests under `deploy/operator/`

Apply the CRD and RBAC:

```bash
kubectl apply -f deploy/crds/rcds_v1_rcds_crd.yaml
kubectl apply -f deploy/operator/
```

Apply the sample custom resource:

```bash
kubectl apply -f deploy/examples/rcds_sample.yaml
```

Important: the repository does not yet include a full controller Deployment manifest. The CRD and RBAC are useful scaffolding, but a production Kubernetes deployment still needs a controller image, Deployment, Service, configuration, and observability wiring.

## GitHub Pages Website

The project website lives in `docs/index.html` with styles in `docs/assets/site.css`. The Pages workflow publishes the `docs/` directory as a static artifact on pushes to `master` or `main`.

To preview locally:

```bash
python3 -m http.server 8000 --directory docs
```

Then open `http://127.0.0.1:8000`.

## Production Notes

- Run set sync behind a trusted network boundary unless you add TLS and authentication.
- Use `--host 0.0.0.0` for container/server binds and a concrete host name for clients.
- Tune `--expected-diff` for IBLT. A poor estimate can require retries or fail probabilistic decoding.
- File sync is exact and checksum-verified, but it is not yet a streaming multi-gigabyte transfer system.
- Use `make test`, `make integration-test`, `make e2e-test`, and `go vet ./...` before release builds.

## Troubleshooting

### Client cannot connect

Check that the server is still running, the port is exposed, and the server was started with a bind address reachable by the client.

### IBLT sync fails

Increase `--expected-diff` and/or `--max-retries`, or use `--algorithm full` to verify the data path.

### File checksum mismatch

The client does not replace the destination if verification fails. Confirm both sides use the intended source/destination paths and the same `--chunk-size` if you are trying to maximize chunk reuse.
