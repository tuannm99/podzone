# E-commerce Microservices Monorepo for Learning Purposes

This is a Go monorepo containing a collection of microservices for an e-commerce platform.

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

## Getting Started

### Prerequisites

- Go 1.23
- Docker and Docker Compose

```bash
Docker version 28.0.2, build 0442a73
Docker Compose version v2.34.0
```

- Protocol Buffers compiler

```bash
google.golang.org/protobuf v1.36.6
github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3
google.golang.org/genproto/googleapis/api v0.0.0-20250303144028-a0af3efb3deb

- Make

```

- Swagger-gen

```bash
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

```

### Setup Development Environment

1. Clone the repository:

   ```bash
   git clone https://github.com/tuannm99/podzone.git
   cd podzone
   ```

2. Install dependencies:

   ```bash
   go mod download
   ```

3. Generate code from Protocol Buffers:

   ```bash
   make proto
   ```

4. Start the development environment:
   ```bash
   make dev
   ```

This will start all services and their dependencies (databases, caches, etc.) using Docker Compose.

### Development Workflow

#### Creating a New Service

```bash
make generate-service
# Enter service name when prompted
```

This will create a new service with the basic structure in place.

#### Building Services

```bash
# Build all services
make build

# Build a specific service
make build SERVICE=catalog
```

#### Running Tests

```bash
# Run all tests
make test

# Run tests for a specific package
go test ./services/catalog/...
```

## Working with Protocol Buffers

1. Define your service API in the `api/proto/<service>` directory
2. Generate code:
   ```bash
   make proto
   ```
3. Implement the generated interfaces in your service

## Shared Packages

The `pkg/` directory contains shared code that can be used across services:

- `config`: Configuration management
- `logging`: Structured logging
- `database`: Database utilities
- `auth`: Authentication and authorization
- `telemetry`: Observability and monitoring
- `middleware`: Common middleware for HTTP and gRPC
- `errors`: Common error types

## Deployment

### Local Development

```bash
make dev
```

### Kubernetes

Kubernetes manifests are provided in the `deployments/kubernetes` directory:

```bash
kubectl apply -f deployments/kubernetes/
```

### CI/CD

A GitHub Actions workflow is included that:

1. Builds and tests all services
2. Builds Docker images
3. Deploys to Kubernetes

## Documentation

- API documentation is available at http://localhost:8080 when running locally
- Service documentation is in each service's README.md file

## Contributing

1. Create a new feature branch
2. Make your changes
3. Add tests
4. Submit a pull request

## License

This project is licensed under the MIT License.
