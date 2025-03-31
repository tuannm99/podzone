# E-commerce Microservices Monorepo

This is a Go monorepo containing a collection of microservices for an e-commerce platform, including code, k8s infrastructure setup on public Cloud, On-premise and local setup

## Getting Started

### Prerequisites

- Go 1.23
- Docker, Kind, kubectl, helm

```bash
Docker version 28.0.2, build 0442a73
Docker Compose version v2.34.0

helm version
version.BuildInfo{Version:"v3.17.0", GitCommit:"301108edc7ac2a8ba79e4ebf5701b0b6ce6a31e4", GitTreeState:"clean", GoVersion:"go1.23.4"}

kubectl version
Client Version: v1.32.0
Kustomize Version: v5.5.0
Server Version: v1.32.0
```

- Protocol Buffers compiler

```bash
google.golang.org/protobuf v1.36.6
github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3
google.golang.org/genproto/googleapis/api v0.0.0-20250303144028-a0af3efb3deb

```

- Swagger-gen

```bash
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

```

This will start all services and their dependencies (databases, caches, etc.) using Docker Compose.

### Development Workflow

#### Setup Development Environment

```bash
# assume that we are in ubuntu >= 22.04
# apt dependency
sudo apt update -y
sudo apt install dmsetup cryptsetup nfs-common open-iscsi -y # k8s longhorn storageclass

# using both kubectl and helm, each components I do easiest way to install that I believe
# docker

# kubectl

# helm

# for local machine using k3s (single node)
curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION="v1.31.0+k3s1" K3S_TOKEN=12345token sh -s - server --disable=traefik --disable=servicelb
sudo cat /etc/rancher/k3s/k3s.yaml > ~/.kubeconfig
export KUBECONFIG=~/.kubeconfig

# local cicd
curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64 && \
chmod +x skaffold && sudo mv skaffold /usr/local/bin
skaffold version

# this is my ingress nginx controller loadbalancer IP
# check it by run `kubectl get svc -n ingress-nginx`
NAME                                 TYPE           CLUSTER-IP      EXTERNAL-IP     PORT(S)                      AGE
ingress-nginx-controller             LoadBalancer   10.43.194.107   10.42.100.100   80:31249/TCP,443:30820/TCP   62m
ingress-nginx-controller-admission   ClusterIP      10.43.6.241     <none>          443/TCP                      62m

# after that we can setup host base routing by setting up ingress
# setting /etc/hosts for some UI Infrastructure
# -> update later as I go
sudo vi /etc/hosts
10.42.100.100 rancher.local.com harbor.local.com jenkin.local.com longhorn.local.com minio.local.com pg-ui.local.com

# installed local cluster
./deployments/kind-cluster-dev/install.sh
# get rancher UI password (or we can view the password is in the script)
kubectl get secret --namespace cattle-system bootstrap-secret -o go-template='{{.data.bootstrapPassword|base64decode}}{{"\n"}}'

# mongodb
kubectl get secret --namespace default mongodb -o jsonpath="{.data.mongodb-root-password}" | base64 -d

# elasticsearch + kibana
- username `elastic`
kubectl get secret esv8-es-elastic-user -o jsonpath='{.data.elastic}' | base64 -d

- we can login on kibana using this account -> kibana.local.com

#

```

## Architecture Overview

This project follows a microservices architecture with the following key components:

- **API Gateway**: Central entry point that routes requests to appropriate services
- **Catalog Service**: Manages product catalog and inventory
- **Order Service**: Handles order processing and management
- **Cart Service**: Manages shopping cart functionality
- **User Service**: Handles user authentication and profiles
- **Payment Service**: Processes payments

Services communicate with each other using gRPC for internal communication, while exposing REST APIs via gRPC-Gateway for external clients.

## Directory Structure

```
podzone/
├── .github/                       # GitHub workflows and CI/CD configurations
├── Makefile                       # Top-level make targets for common operations
├── docker-compose.yml             # Local development environment setup
├── go.mod                         # Root Go modules file
├── go.sum                         # Dependency checksums
├── tools/                         # Development tools and scripts
│   ├── protoc/                    # Protobuf compiler scripts
│   ├── db/                        # Database migration tools
│   └── ci/                        # CI/CD helper scripts
├── api/                           # API definitions (protobuf, OpenAPI specs)
│   ├── proto/                     # Protocol Buffers definitions
│   │   ├── catalog/               # Catalog service API definitions
│   │   ├── order/                 # Order service API definitions
│   │   ├── payment/               # Payment service API definitions
│   │   ├── user/                  # User service API definitions
│   │   └── cart/                  # Cart service API definitions
│   └── swagger/                   # OpenAPI/Swagger definitions
├── pkg/                           # Shared packages
│   ├── auth/                      # Authentication and authorization
│   ├── config/                    # Configuration management
│   ├── database/                  # Database utilities
│   ├── errors/                    # Common error types
│   ├── logging/                   # Logging utilities
│   ├── messaging/                 # Message queue utilities
│   ├── middleware/                # Common middleware
│   ├── models/                    # Shared data models
│   ├── telemetry/                 # Observability and monitoring
│   ├── testing/                   # Test utilities
│   └── validator/                 # Input validation
├── services/                      # Individual microservices
│   ├── catalog/                   # Catalog service
│   │   ├── cmd/                   # Service entry point
│   │   ├── internal/              # Service-specific private code
│   │   ├── Dockerfile             # Service-specific Docker build
│   │   └── README.md              # Service documentation
│   ├── order/                     # Order service
│   ├── payment/                   # Payment service
│   ├── user/                      # User service
│   ├── cart/                      # Cart service
│   └── gateway/                   # API Gateway
└── deployments/                   # Kubernetes and deployment manifests
    ├── kubernetes/                # K8s manifests
    ├── terraform/                 # Infrastructure as code
    └── helm/                      # Helm charts
```
