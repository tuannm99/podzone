.PHONY: all proto swagger build test lint dev down clean help

SERVICES := catalog order payment user cart gateway
GO := go
DOCKER_COMPOSE := docker-compose
PROTOC := protoc

# Colors for output
COLOR_RESET = \033[0m
COLOR_GREEN = \033[32m
COLOR_YELLOW = \033[33m

all: proto swagger build

# this run only first when created project, or run when a updated happens
init_proto:
	@echo "$(COLOR_GREEN)Setting up Proto dependencies...$(COLOR_RESET)"
	@mkdir -p third_party/google/api
	@mkdir -p third_party/google/protobuf
	@echo "$(COLOR_YELLOW)Downloading Google API proto files...$(COLOR_RESET)"
	@curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto > third_party/google/api/annotations.proto
	@curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > third_party/google/api/http.proto
	@echo "$(COLOR_YELLOW)Downloading Google Protobuf proto files...$(COLOR_RESET)"
	@curl -sSL https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/timestamp.proto > third_party/google/protobuf/timestamp.proto
	@curl -sSL https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/empty.proto > third_party/google/protobuf/empty.proto
	@curl -sSL https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/any.proto > third_party/google/protobuf/any.proto
	@curl -sSL https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/struct.proto > third_party/google/protobuf/struct.proto
	@curl -sSL https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/wrappers.proto > third_party/google/protobuf/wrappers.proto
	@echo "$(COLOR_GREEN)Proto dependencies setup complete.$(COLOR_RESET)"


proto:
	@echo "$(COLOR_GREEN)Generating protobuf code...$(COLOR_RESET)"
	@mkdir -p pkg/api/proto
	@for service in $(SERVICES); do \
		if [ -d api/proto/$$service ]; then \
			echo "Generating protobuf for $$service"; \
			$(PROTOC) \
				--proto_path=api/proto \
				--proto_path=third_party \
				--go_out=pkg/api/proto \
				--go_opt=paths=source_relative \
				--go-grpc_out=pkg/api/proto \
				--go-grpc_opt=paths=source_relative \
				--grpc-gateway_out=pkg/api/proto \
				--grpc-gateway_opt=paths=source_relative \
				api/proto/$$service/*.proto || echo "Error generating proto for $$service"; \
		fi; \
	done

# Generate swagger docs
swagger:
	@echo "$(COLOR_GREEN)Generating OpenAPI/Swagger specs from proto files...$(COLOR_RESET)"
	@for service in $(SERVICES); do \
		if [ -d api/proto/$$service ]; then \
			echo "Generating OpenAPI spec for $$service"; \
			mkdir -p api/swagger-gen/$$service; \
			$(PROTOC) \
				--proto_path=api/proto \
				--proto_path=third_party \
				--openapiv2_out=api/swagger-gen/$$service \
				--openapiv2_opt=logtostderr=true \
				--openapiv2_opt=json_names_for_fields=true \
				api/proto/$$service/*.proto || echo "Error generating OpenAPI spec for $$service"; \
		fi; \
	done

# Generate both proto code and Swagger documentation
api: proto swagger
	@echo "$(COLOR_GREEN)API generation complete.$(COLOR_RESET)"

# Build all services
build:
	@echo "$(COLOR_GREEN)Building all services...$(COLOR_RESET)"
	@for service in $(SERVICES); do \
		if [ -d services/$$service ]; then \
			echo "Building $$service"; \
			cd services/$$service && $(GO) build -o bin/$$service cmd/*.go && cd ../..; \
		fi; \
	done

# Run tests for all packages
test:
	@echo "$(COLOR_GREEN)Running tests...$(COLOR_RESET)"
	$(GO) test -v ./pkg/... ./services/...

# Run linter
lint:
	@echo "$(COLOR_GREEN)Running linter...$(COLOR_RESET)"
	golangci-lint run ./...

# Start development environment
dev:
	@echo "$(COLOR_GREEN)Starting development environment...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) up -d

# Stop development environment
down:
	@echo "$(COLOR_GREEN)Stopping development environment...$(COLOR_RESET)"
	$(DOCKER_COMPOSE) down

# Clean generated files and binaries
clean:
	@echo "$(COLOR_GREEN)Cleaning up...$(COLOR_RESET)"
	@find . -name "bin" -type d -exec rm -rf {} +
	@find . -name "*.pb.go" -type f -delete
	@find . -name "*.swagger.json" -type f -delete

# Generate a new service 
generate-service:
	@read -p "Enter service name: " service; \
	mkdir -p services/$$service/cmd services/$$service/internal; \
	mkdir -p services/$$service/internal/handler services/$$service/internal/repository services/$$service/internal/service; \
	mkdir -p api/proto/$$service; \
	cp tools/templates/service_template/* services/$$service/ -r; \
	sed -i 's/{{SERVICE_NAME}}/'"$$service"'/g' services/$$service/cmd/main.go; \
	echo "Service $$service created"

# Help 
help:
	@echo "$(COLOR_YELLOW)Available commands:$(COLOR_RESET)"
	@echo "  make proto          - Generate protobuf code"
	@echo "  make swagger        - Generate Swagger documentation"
	@echo "  make build          - Build all services"
	@echo "  make test           - Run tests"
	@echo "  make lint           - Run linter"
	@echo "  make dev            - Start development environment"
	@echo "  make down           - Stop development environment"
	@echo "  make clean          - Clean generated files"
	@echo "  make generate-service - Create a new service"
