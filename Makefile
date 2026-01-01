.PHONY: all proto swagger build test lint dev down clean help

GO := go

# Colors for output
COLOR_RESET = \033[0m
COLOR_GREEN = \033[32m
COLOR_YELLOW = \033[33m

SVC ?= none
.PHONY: svc help

BUF ?= buf

proto:
	@echo "Generating protobuf via buf..."
	@$(BUF) generate

proto_lint:
	@$(BUF) lint

proto_svc:
	@echo "$(COLOR_GREEN)Generating protobuf for $(SVC)...$(COLOR_RESET)"
	@$(BUF) generate --path api/proto/$(SVC)


# Run tests for all packages
test:
	@echo "$(COLOR_GREEN)Running tests...$(COLOR_RESET)"
	$(GO) test -v -cover ./pkg/... ./internal/... ./cmd/...

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
	@set -e; \
	export GOOSE_DRIVER=postgres; \
	export GOOSE_DBSTRING="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"; \
	goose -dir ./internal/auth/migrations/sql status; \
	goose -dir ./internal/auth/migrations/sql up
# 	goose -dir ./internal/auth/migrations/sql down

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
	  -Dsonar.exclusions=**/*_test.go,**/pkg/api/**,**/docs/**,**/api/**,**/third_party/**,**/scripts/**,**/deployments/**,**/node_modules/**,**/generated/**,**/mocks/**,**/migrations/**,**/*_templ.go,**/*.jsx,**/build/**,**/dist/**,**/*.config.* \
	  -Dsonar.test.inclusions=**/*_test.go,**/generated/**,**/migrations/**,**/*_templ.go,**/*.jsx,**/build/**,**/dist/**,**/*.config.* \
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
		kubectl port-forward svc/kafka 29092:29092 -n default & pids+=($$!); \
		kubectl port-forward svc/elasticsearch 9200:9200 -n default & pids+=($$!); \
		kubectl port-forward svc/consul 8500:8500 -n default & pids+=($$!); \
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
	@echo "  make k8s-ui ENV=${env} SVC=${service} - Deploy service to k8s dev EG: make k8s-ui ENV="staging" SVC="ui-podzone""

