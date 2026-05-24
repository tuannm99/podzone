# E-commerce Microservices Monorepo for Experimentation and Learning Purposes

This is a Go monorepo containing a collection of microservices for an e-commerce platform, including code, k8s infrastructure setup on public Cloud (AWS) and local setup (k3s)

## 🚀 Project Roadmap

**Development Environment Setup**

- Repository initialization ✅
- gRPC/gRPC gateway implementation ✅
- Swagger API documentation generation ✅
- Docker and docker-compose configuration ✅
- Persistence layer setup ✅
- API Gateway integration ✅
- Custom Gateway plugins development
- K8s local development ✅
- DI ✅
- Clean Architecture hands-on (now auth service) ✅
- Testing ✅
- Code Quality using Sonarqube Community ✅
- Installation/Initialization docs and scripts

🔄 **Microservice Implementation**

- Gateway Service (apisix ✅) (TODO: replace it by using ./nextgen-gateway - if I can complete)
- Auth Service (gRPC) ✅
- User/Tenant Service (gRPC)
- Cart Service (gRPC)
- Catalog Service (gRPC)
- Order Service (gRPC)
- Payment Service (gRPC)
- gRPC-Gateway Service (thin layer transform from grpc to http for all these admin services) ✅
- Onboarding Service (Restful API, In-progress)
- Backoffice (Graphql, monolithic Seller Portal - domain driven design + clean-arch, In-progress)
- Storefront (Server Side Rendering, for surfing products of Seller Store)

🔄 **Our pkg**

- reused tools, use Fx as backbone, Koanf yaml file
- pkg/pdtenantdb - multi-tenant db manager, used for manage connection pool/db switching (In-progress)
- pkg/toolkit - utility

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
- service mesh implementation ✅ istio
  - Header detection middleware
  - Sidecar deployment configuration
  - Kubernetes deployment updates
  - Distributed tracing exporters

### 🔥 Moreover

- Using database for one simple service that I am writing from scratch ([NovaSQL](https://github.com/tuannm99/novasql)) - in-progress

## 🏗️ Architecture Overview

- Architecture docs live under [docs/architecture](./docs/architecture/README.md)
- C4 views:
  - [System Context](./docs/architecture/system-context.md)
  - [Containers](./docs/architecture/containers.md)
  - [Modules](./docs/architecture/modules.md)
  - [Frontend and Edge](./docs/architecture/frontend-edge.md)
  - [Transport and Contracts](./docs/architecture/transport-contracts.md)
  - [Data Ownership](./docs/architecture/data-ownership.md)
  - [Code Map](./docs/architecture/code-map.md)
  - [Deployment](./docs/architecture/deployment.md)
  - [Bounded Contexts](./docs/architecture/bounded-contexts.md)
  - [Platform Runtime](./docs/architecture/platform-runtime.md)
  - [Sequences](./docs/architecture/sequences.md)

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

- Go 1.25
- Docker, Kind, kubectl, helm

````bash
Docker version 28.0.2, build 0442a73
Docker Compose version v2.34.0

- helm version
version.BuildInfo{Version:"v3.17.0", GitCommit:"301108edc7ac2a8ba79e4ebf5701b0b6ce6a31e4", GitTreeState:"clean", GoVersion:"go1.23.4"}

- kubectl version
Client Version: v1.32.0
Kustomize Version: v5.5.0
Server Version: v1.32.0

- Golang CLI tools (go1.25.1)
go install github.com/99designs/gqlgen@v0.17.81
go install github.com/vektra/mockery/v3@v3.6.1
go install github.com/air-verse/air@v1.63.0
go install github.com/bufbuild/buf/cmd/buf@v1.61.0
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.7.2
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.27.3
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.27.3
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

### Development Workflow

#### run dev service

```bash
# Hot reload with air
air --build.cmd "go build -o ./bin/${svc} ./cmd/${svc}/main.go" --build.bin "./bin/${svc}"

````

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

# for dev development, on local machine using k3s (single node)
curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION="v1.31.0+k3s1" K3S_TOKEN=12345token sh -s - server --disable=traefik --disable=servicelb
sudo cat /etc/rancher/k3s/k3s.yaml > ~/.kubeconfig
export KUBECONFIG=~/.kubeconfig

# or istio
kubectl create namespace istio-system
helm install istiod istio/istiod -n istio-system --wait
helm install istiod istio/istiod -n istio-system --wait

kubectl get crd gateways.gateway.networking.k8s.io &> /dev/null || \
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.3.0/standard-install.yaml

kubectl label namespace default istio-injection=enabled

helm status istiod -n istio-system
helm get all istiod -n istio-system


# k8s dashboard
helm upgrade --install kubernetes-dashboard kubernetes-dashboard/kubernetes-dashboard --create-namespace --namespace kubernetes-dashboard

kubectl create serviceaccount -n kubernetes-dashboard admin-user
kubectl create clusterrolebinding -n kubernetes-dashboard admin-user --clusterrole cluster-admin --serviceaccount=kubernetes-dashboard:admin-user
token=$(kubectl -n kubernetes-dashboard create token admin-user)

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
10.42.100.100 rancher.tuannm.uk longhorn.tuannm.uk minio.tuannm.uk pg-ui.tuannm.uk
10.42.100.100 gateway.tuannm.uk podzone-ui.tuannm.uk kafka-ui.tuannm.uk
10.42.100.100 sonarqube.tuannm.uk harbor.tuannm.uk jenkins.tuannm.uk kibana.tuannm.uk redisinsight.tuannm.uk

```
