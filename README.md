# E-commerce Microservices Monorepo for Experimentation and Learning Purposes

This is a Go monorepo containing a collection of microservices for an e-commerce platform, including code, k8s infrastructure setup on public Cloud (AWS) and local setup (k3s)

## 🚀 Project Roadmap

**Development Environment Setup**

- Repository initialization ✅
- gRPC implementation ✅
- Swagger API documentation generation ✅
- Docker and docker-compose configuration ✅
- Persistence layer setup ✅
- API Gateway integration ✅
- Custom Gateway plugins development
- K8s local development ✅
- DI ✅
- Clean Architecture hands-on (only for auth service) ✅
- Testing

🔄 **Microservice Implementation**

- Gateway Service ✅
- Auth Service ✅
- Cart Service
- Catalog Service
- Order Service
- Payment Service
- User Service
- Onboarding Service
- Storefront

### Planned

📝 **Documentation**

- Requirements specification
- Architecture documentation
  - C4 model diagrams (to component level)
  - High-level architecture overview
  - Database design
  - API design guidelines
  - Low-level component design

🏭 **Production Environment**

- Kubernetes deployment manifests/Helm charts
- Infrastructure as Code (Terraform for AWS, Ansible for on-prem)
- Persistent storage configuration
- Observability stack
  - Prometheus for metrics
  - Grafana for visualization
  - Jaeger for distributed tracing
  - OpenTelemetry for instrumentation
  - ELK for centralized logging
- service mesh implementation ✅ consul
  - Header detection middleware
  - Sidecar deployment configuration
  - Kubernetes deployment updates
  - Distributed tracing exporters

### 🔥 Moreover

- Using database for one simple service that I am writing from scratch ([NovaSQL](https://github.com/tuannm99/novasql)) - in-progress

## 🏗️ Architecture Overview

- TODO

## 🔧 Additional Considerations

### Security

- Authentication and authorization
- API security testing
- Secret management (Vault/AWS KMS/etc.)
- Container security scanning

### CI/CD

- GitHub Actions/Jenkins/GitLab CI pipeline
- Automated testing
- Deployment automation
- Blue/Green deployment strategy

### Data Management

- Data migration strategies
- Backup and recovery processes
- Data governance policies

### Performance

- Load testing with tools like k6 or JMeter
- Performance monitoring
- Auto-scaling policies

### Compliance

- Logging standards
- Audit trails
- Compliance reporting

## Getting Started

### Prerequisites

- Go 1.23
- Docker, Kind, kubectl, helm

```bash
Docker version 28.0.2, build 0442a73
Docker Compose version v2.34.0

- helm version
version.BuildInfo{Version:"v3.17.0", GitCommit:"301108edc7ac2a8ba79e4ebf5701b0b6ce6a31e4", GitTreeState:"clean", GoVersion:"go1.23.4"}

- kubectl version
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

### Development Workflow

#### run dev service

```bash
# Hot reload with air
air --build.cmd "go build -o ./bin/${svc} ./cmd/${svc}/main.go" --build.bin "./bin/${svc}"

```

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

# install service mesh (consul)
helm repo add hashicorp https://helm.releases.hashicorp.com
helm install --values deployments/kubernetes/staging/consul-mesh/values.yml consul hashicorp/consul --create-namespace --namespace consul --version "1.2.0"

kubectl get --namespace consul secrets/consul-bootstrap-acl-token --template={{.data.token}} | base64 -d

# create secret from .env
kubectl create secret generic global-secrets --from-env-file=.env

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
```
