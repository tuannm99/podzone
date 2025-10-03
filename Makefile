.PHONY: all proto swagger build test lint dev down clean help

SERVICES := catalog order billing user cart auth grpcgateway
GO := go
DOCKER_COMPOSE := docker-compose
PROTOC := protoc

# Colors for output
COLOR_RESET = \033[0m
COLOR_GREEN = \033[32m
COLOR_YELLOW = \033[33m

SVC ?= none
.PHONY: svc help

# Generate both proto code and Swagger documentation
api: proto swagger
	@echo "$(COLOR_GREEN)API generation complete.$(COLOR_RESET)"

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
	@curl -sSL https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/descriptor.proto > third_party/google/protobuf/descriptor.proto
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

# Run tests for all packages
test:
	@echo "$(COLOR_GREEN)Running tests...$(COLOR_RESET)"
	$(GO) test -v ./pkg/... ./internal/... ./cmd/...

# Run linter
lint:
	@echo "$(COLOR_GREEN)Running linter...$(COLOR_RESET)"
	golangci-lint run ./...

gql-backoffice:
	go run github.com/99designs/gqlgen generate

dev:
	@echo "🔁 Starting services in parallel..."
	@for svc in $(SVC); do \
		echo "▶ Starting $$svc..."; \
		CONFIG_PATH=cmd/$$svc/config.yml air --build.cmd "go build -o ./bin/$$svc ./cmd/$$svc/main.go" \
			--build.bin "./bin/$$svc" & \
	done; \
	wait

migrate:
	export GOOSE_DRIVER=postgres
	export GOOSE_DBSTRING="postgres://user:pass@localhost:5432/dbname?sslmode=disable"
	goose -dir ./internal/auth/migrations/sql status
	goose -dir ./internal/auth/migrations/sql up
	goose -dir ./internal/auth/migrations/sql down

sonar:
	@echo "🔁 Starting quality check..."
	go test ./internal/... ./pkg/... ./cmd/... -coverprofile=coverage.out
	docker run --rm \
	  -e SONAR_HOST_URL="http://sonarqube.tuannm.uk" \
	  -v $(shell pwd):/usr/src \
	  -v /etc/hosts:/etc/hosts:ro \
	  sonarsource/sonar-scanner-cli \
	  -Dsonar.projectKey=podzone \
	  -Dsonar.sources=. \
	  -Dsonar.exclusions=**/*_test.go,**/pkg/api/**,**/docs/**,**/api/**,**/third_party/**,**/scripts/**,**/deployments/**,**/node_modules/**,**/generated/**,**/mocks/**,**/migrations/** \
	  -Dsonar.test.inclusions=**/*_test.go,**/generated/**,**/migrations/** \
	  -Dsonar.go.coverage.reportPaths=coverage.out \
	  -Dsonar.login=admin \
	  -Dsonar.password=1

k8s:
	@echo "📦 Building and deploying api services..."
	@for svc in $(SVC); do \
		echo "🚀 Building $$svc..."; \
		docker build -t tuannm99/podzone-$$svc:$(ENV) \
			--build-arg SERVICE_NAME=$$svc \
			-f Dockerfile .; \
		docker push tuannm99/podzone-$$svc:$(ENV); \
		kubectl delete -f deployments/kubernetes/$(ENV)/services/$$svc.yml --ignore-not-found; \
		kubectl apply -f deployments/kubernetes/$(ENV)/services/$$svc.yml; \
	done

k8s-ui:
	@echo "📦 Building and deploying ui services..."
	set -a; source .env; set +a;
	for svc in $(SVC); do \ echo "🚀 Building $$svc..."; \
		docker build -t tuannm99/podzone-$$svc:$(ENV) \
			--build-arg SERVICE_NAME=$$svc \
			--build-arg VITE_ADMIN_API_URL=$$VITE_ADMIN_API_URL \
			-f Dockerfile-ui .; \
		docker push tuannm99/podzone-$$svc:$(ENV); \
		kubectl delete -f deployments/kubernetes/$(ENV)/services/$$svc.yml --ignore-not-found; \
		kubectl apply -f deployments/kubernetes/$(ENV)/services/$$svc.yml; \
	done

portfw:
	@bash -c '\
		set -e; \
		pids=(); \
		cleanup() { \
			echo "Cleaning up..."; \
			for pid in "$${pids[@]}"; do \
				echo "Killing PID $$pid"; \
				kill $$pid 2>/dev/null || true; \
			done; \
		}; \
		trap cleanup EXIT INT TERM; \
		\
		kubectl port-forward svc/redis 6379:6379 -n default & pids+=($$!); \
		kubectl port-forward svc/postgres 5432:5432 -n default & pids+=($$!); \
		kubectl port-forward svc/mongodb-internal 27017:27017 -n default & pids+=($$!); \
		kubectl port-forward svc/kafka 9092:9092 -n default & pids+=($$!); \
		kubectl port-forward svc/elasticsearch 9200:9200 -n default & pids+=($$!); \
		# kubectl port-forward svc/redisinsight-service 8888:80 -n default & pids+=($$!); \
		# kubectl port-forward svc/kibana 5601:5601 -n default & pids+=($$!); \
		# kubectl port-forward svc/sonarqube 9000:9000 -n devops --address=0.0.0.0 & pids+=($$!); \
		# kubectl port-forward svc/pgadmin 8889:80 -n default & pids+=($$!); \
		# kubectl port-forward svc/kubernetes-dashboard-kong-proxy 8001:443 -n kubernetes-dashboard & pids+=($$!); \
		\
		wait -n || (echo "One process failed"; exit 1); \
	'

help:
	@echo "$(COLOR_YELLOW)Available commands:$(COLOR_RESET)"
	@echo "  make api                              - Generate protobuf code, swagger api"
	@echo "  make test                             - Run tests"
	@echo "  make lint                             - Run linter"
	@echo "  make portfw                           - Portfowrding"
	@echo "  make dev SVC=${service}               - Run service"
	@echo "  make gql-backoffice                   - Generate backoffice graphql"
	@echo "  make k8s ENV=${env} SVC=${service}    - Deploy service to k8s dev EG: make k8s ENV="staging" SVC="grpcgateway catalog auth storefront backoffice""
	@echo "  make k8s-ui ENV=${env} SVC=${service} - Deploy service to k8s dev EG: make k8s-ui ENV="staging" SVC="ui-admin""

