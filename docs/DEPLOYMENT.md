# RCDS Deployment Guide

This guide covers deploying RCDS in various environments.

## Table of Contents

- [Standalone Deployment](#standalone-deployment)
- [Docker Deployment](#docker-deployment)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Configuration](#configuration)

## Standalone Deployment

### Binary Installation

1. Download the latest release from [GitHub Releases](https://github.com/String-Reconciliation-Ditributed-System/RCDS_GO/releases)
2. Extract the binary:
   ```bash
   tar -xzf rcds-linux-amd64.tar.gz
   ```
3. Move to a directory in your PATH:
   ```bash
   sudo mv rcds /usr/local/bin/
   ```

### Build from Source

```bash
git clone https://github.com/String-Reconciliation-Ditributed-System/RCDS_GO.git
cd RCDS_GO
make build
sudo cp bin/rcds /usr/local/bin/
```

## Docker Deployment

### Building the Docker Image

```bash
docker build -t rcds:latest .
```

### Running as Server

```bash
docker run -d \
  --name rcds-server \
  -p 8080:8080 \
  rcds:latest server --port 8080
```

### Running as Client

```bash
docker run -d \
  --name rcds-client \
  rcds:latest client --server rcds-server:8080
```

### Docker Compose

Create a `docker-compose.yml`:

```yaml
version: '3.8'

services:
  rcds-server:
    image: rcds:latest
    ports:
      - "8080:8080"
    command: server --port 8080
    
  rcds-client:
    image: rcds:latest
    depends_on:
      - rcds-server
    command: client --server rcds-server:8080
```

Run with:

```bash
docker-compose up -d
```

## Kubernetes Deployment

### Prerequisites

- Kubernetes cluster (1.19+)
- kubectl configured
- Appropriate RBAC permissions

### Install CRD

```bash
kubectl apply -f deploy/crds/rcds_v1_rcds_crd.yaml
```

### Install Operator

```bash
kubectl apply -f deploy/operator.yaml
kubectl apply -f deploy/role.yaml
kubectl apply -f deploy/role_binding.yaml
kubectl apply -f deploy/service_account.yaml
```

### Deploy RCDS Instance

Create a custom resource:

```yaml
apiVersion: rcds.distributed-system.io/v1
kind: RCDS
metadata:
  name: rcds-sample
  namespace: default
spec:
  replicas: 3
  algorithm: "iblt"
  port: 8080
  resources:
    limits:
      cpu: "1"
      memory: "512Mi"
    requests:
      cpu: "100m"
      memory: "128Mi"
```

Apply it:

```bash
kubectl apply -f rcds-instance.yaml
```

### Verify Deployment

```bash
# Check pods
kubectl get pods -l app=rcds

# Check RCDS resources
kubectl get rcds

# Check logs
kubectl logs -l app=rcds -f
```

### Scaling

Scale the RCDS deployment:

```bash
kubectl scale rcds rcds-sample --replicas=5
```

Or edit the resource:

```bash
kubectl edit rcds rcds-sample
```

### Monitoring

RCDS exposes Prometheus metrics on `/metrics` endpoint.

1. Deploy Prometheus:
   ```bash
   kubectl apply -f deploy/monitoring/prometheus.yaml
   ```

2. Deploy ServiceMonitor:
   ```bash
   kubectl apply -f deploy/monitoring/service-monitor.yaml
   ```

3. Access metrics:
   ```bash
   kubectl port-forward svc/prometheus 9090:9090
   ```

### Troubleshooting

#### Pods not starting

Check events:
```bash
kubectl describe pod <pod-name>
kubectl get events --sort-by='.lastTimestamp'
```

#### CRD not found

Reinstall CRD:
```bash
kubectl delete crd rcds.rcds.distributed-system.io
kubectl apply -f deploy/crds/rcds_v1_rcds_crd.yaml
```

#### Connection issues

Check service:
```bash
kubectl get svc -l app=rcds
kubectl describe svc <service-name>
```

Check network policies:
```bash
kubectl get networkpolicy
```

## Configuration

### Environment Variables

- `RCDS_PORT`: Server listening port (default: 8080)
- `RCDS_ALGORITHM`: Reconciliation algorithm (iblt, cpi, full)
- `RCDS_LOG_LEVEL`: Log level (debug, info, warn, error)

### Configuration File

Create `rcds-config.yaml`:

```yaml
server:
  port: 8080
  host: "0.0.0.0"

algorithm:
  type: "iblt"
  parameters:
    tableSize: 1000
    hashCount: 3

logging:
  level: "info"
  format: "json"
```

Use with:

```bash
rcds --config rcds-config.yaml
```

## Security Considerations

### Network Security

1. Use TLS for production deployments
2. Implement network policies in Kubernetes
3. Restrict access using firewalls

### RBAC

In Kubernetes, ensure proper RBAC:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: rcds-operator
rules:
- apiGroups: ["rcds.distributed-system.io"]
  resources: ["rcds"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

## Performance Tuning

### Memory

Adjust based on dataset size:

```yaml
resources:
  limits:
    memory: "2Gi"
  requests:
    memory: "1Gi"
```

### CPU

For high-throughput scenarios:

```yaml
resources:
  limits:
    cpu: "2"
  requests:
    cpu: "500m"
```

### Network

Optimize for your environment:

- Use persistent connections
- Adjust buffer sizes
- Enable compression if supported

## Backup and Recovery

### Backup

```bash
# Backup CRD definitions
kubectl get rcds -o yaml > rcds-backup.yaml

# Backup operator configuration
kubectl get deployment rcds-operator -o yaml > operator-backup.yaml
```

### Recovery

```bash
# Restore CRD instances
kubectl apply -f rcds-backup.yaml

# Restore operator
kubectl apply -f operator-backup.yaml
```

## Upgrading

### Rolling Update

```bash
kubectl set image deployment/rcds-operator rcds=rcds:v0.2.0
kubectl rollout status deployment/rcds-operator
```

### Rollback

```bash
kubectl rollout undo deployment/rcds-operator
```

## Support

For issues or questions:

- Open an issue on [GitHub](https://github.com/String-Reconciliation-Ditributed-System/RCDS_GO/issues)
- Check the [documentation](https://github.com/String-Reconciliation-Ditributed-System/RCDS_GO/tree/master/docs)
- Review [examples](https://github.com/String-Reconciliation-Ditributed-System/RCDS_GO/tree/master/examples)
