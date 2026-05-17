.PHONY: all proto swagger build test lint dev down clean help docker-dev docker-dev-infra docker-dev-down mocks dev-backoffice-seed dev-backoffice-sample dev-auth-bootstrap dev-ui-auth-sync dev-pod-sample

GO := go

# Colors for output
COLOR_RESET = \033[0m
COLOR_GREEN = \033[32m
COLOR_YELLOW = \033[33m

SVC ?= none
.PHONY: svc help

BUF ?= buf
MOCKERY ?= mockery

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
	@set -e; \
	export PODZONE_TC_REUSE=1; \
	$(GO) test -v -cover ./pkg/... ./internal/... ./cmd/...

# Run linter
lint:
	@echo "$(COLOR_GREEN)Running linter...$(COLOR_RESET)"
	golangci-lint run ./...

docker-dev-infra:
	docker compose -f deployments/docker/infras.yml up -d

docker-dev:
	docker compose -f deployments/docker/infras.yml -f deployments/docker/services.yml up --build

docker-dev-down:
	docker compose -f deployments/docker/infras.yml -f deployments/docker/services.yml down

dev-backoffice-seed:
	@TENANT_ID=$(TENANT_ID) \
	STORE_NAME=$(STORE_NAME) \
	STORE_SUBDOMAIN=$(STORE_SUBDOMAIN) \
	CLUSTER_NAME=$(CLUSTER_NAME) \
	DB_NAME=$(DB_NAME) \
	SCHEMA_NAME=$(SCHEMA_NAME) \
	PG_HOST=$(PG_HOST) \
	PG_PORT=$(PG_PORT) \
	PG_USER=$(PG_USER) \
	PG_PASSWORD=$(PG_PASSWORD) \
	PG_SSL_MODE=$(PG_SSL_MODE) \
	CONSUL_URL=$(CONSUL_URL) \
	ONBOARDING_URL=$(ONBOARDING_URL) \
	CREATE_STORE=$(CREATE_STORE) \
	sh scripts/dev/seed_backoffice_tenant.sh

dev-backoffice-sample:
	@TENANT_ID=$(TENANT_ID) \
	STORE_NAME=$(STORE_NAME) \
	STORE_SUBDOMAIN=$(STORE_SUBDOMAIN) \
	CLUSTER_NAME=$(CLUSTER_NAME) \
	DB_NAME=$(DB_NAME) \
	SCHEMA_NAME=$(SCHEMA_NAME) \
	PG_HOST=$(PG_HOST) \
	PG_PORT=$(PG_PORT) \
	PG_USER=$(PG_USER) \
	PG_PASSWORD=$(PG_PASSWORD) \
	PG_SSL_MODE=$(PG_SSL_MODE) \
	CONSUL_URL=$(CONSUL_URL) \
	ONBOARDING_URL=$(ONBOARDING_URL) \
	CREATE_STORE=$(CREATE_STORE) \
	sh scripts/dev/seed_backoffice_tenant.sh
	@TENANT_ID=$(TENANT_ID) \
	STORE_NAME=$(STORE_NAME) \
	STORE_SUBDOMAIN=$(STORE_SUBDOMAIN) \
	DB_NAME=$(DB_NAME) \
	SCHEMA_NAME=$(SCHEMA_NAME) \
	PG_HOST=$(PG_HOST) \
	PG_PORT=$(PG_PORT) \
	PG_USER=$(PG_USER) \
	PG_PASSWORD=$(PG_PASSWORD) \
	PG_SSL_MODE=$(PG_SSL_MODE) \
	ONBOARDING_URL=$(ONBOARDING_URL) \
	$(GO) run ./scripts/dev/seed_backoffice_sample.go

dev-auth-bootstrap:
	@TENANT_ID=$(TENANT_ID) \
	TENANT_NAME=$(TENANT_NAME) \
	TENANT_SLUG=$(TENANT_SLUG) \
	DEV_USERNAME=$(DEV_USERNAME) \
	DEV_EMAIL=$(DEV_EMAIL) \
	DEV_PASSWORD=$(DEV_PASSWORD) \
	DEV_FULL_NAME=$(DEV_FULL_NAME) \
	PG_HOST=$(PG_HOST) \
	PG_PORT=$(PG_PORT) \
	PG_USER=$(PG_USER) \
	PG_PASSWORD=$(PG_PASSWORD) \
	PG_SSL_MODE=$(PG_SSL_MODE) \
	JWT_SECRET=$(JWT_SECRET) \
	JWT_KEY=$(JWT_KEY) \
	AUTH_BOOTSTRAP_OUTPUT=$(AUTH_BOOTSTRAP_OUTPUT) \
	$(GO) run ./scripts/dev/seed_auth_bootstrap.go

dev-ui-auth-sync:
	@AUTH_BOOTSTRAP_OUTPUT=$(AUTH_BOOTSTRAP_OUTPUT) \
	UI_AUTH_BOOTSTRAP_TARGET=$(UI_AUTH_BOOTSTRAP_TARGET) \
	sh scripts/dev/sync_ui_auth_bootstrap.sh

dev-pod-sample:
	@$(MAKE) dev-backoffice-sample \
		TENANT_ID=$(TENANT_ID) \
		STORE_NAME=$(STORE_NAME) \
		STORE_SUBDOMAIN=$(STORE_SUBDOMAIN) \
		TENANT_NAME=$(TENANT_NAME) \
		TENANT_SLUG=$(TENANT_SLUG) \
		CLUSTER_NAME=$(CLUSTER_NAME) \
		DB_NAME=$(DB_NAME) \
		SCHEMA_NAME=$(SCHEMA_NAME) \
		PG_HOST=$(PG_HOST) \
		PG_PORT=$(PG_PORT) \
		PG_USER=$(PG_USER) \
		PG_PASSWORD=$(PG_PASSWORD) \
		PG_SSL_MODE=$(PG_SSL_MODE) \
		CONSUL_URL=$(CONSUL_URL) \
		ONBOARDING_URL=$(ONBOARDING_URL) \
		CREATE_STORE=$(CREATE_STORE)
	@$(MAKE) dev-auth-bootstrap \
		TENANT_ID=$(TENANT_ID) \
		TENANT_NAME=$(TENANT_NAME) \
		TENANT_SLUG=$(TENANT_SLUG) \
		DEV_USERNAME=$(DEV_USERNAME) \
		DEV_EMAIL=$(DEV_EMAIL) \
		DEV_PASSWORD=$(DEV_PASSWORD) \
		DEV_FULL_NAME=$(DEV_FULL_NAME) \
		PG_HOST=$(PG_HOST) \
		PG_PORT=$(PG_PORT) \
		PG_USER=$(PG_USER) \
		PG_PASSWORD=$(PG_PASSWORD) \
		PG_SSL_MODE=$(PG_SSL_MODE) \
		JWT_SECRET=$(JWT_SECRET) \
		JWT_KEY=$(JWT_KEY) \
		AUTH_BOOTSTRAP_OUTPUT=$(AUTH_BOOTSTRAP_OUTPUT)
	@$(MAKE) dev-ui-auth-sync \
		AUTH_BOOTSTRAP_OUTPUT=$(AUTH_BOOTSTRAP_OUTPUT) \
		UI_AUTH_BOOTSTRAP_TARGET=$(UI_AUTH_BOOTSTRAP_TARGET)

gql-backoffice:
	go run github.com/99designs/gqlgen generate

mocks:
	@echo "$(COLOR_GREEN)Generating mocks via mockery...$(COLOR_RESET)"
	@GOCACHE=$${GOCACHE:-/tmp/podzone-mockery-cache} $(MOCKERY)

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
	go run github.com/pressly/goose/v3/cmd/goose -dir ./internal/auth/migrations/sql status; \
	go run github.com/pressly/goose/v3/cmd/goose -dir ./internal/auth/migrations/sql up
# 	go run github.com/pressly/goose/v3/cmd/goose -dir ./internal/auth/migrations/sql down

sonar:
	@echo "🔁 Starting quality check..."
	@set -e; \
	export PODZONE_TC_REUSE=1; \
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
		kubectl port-forward svc/pgbouncer 6432:6432 -n staging & pids+=($$!); \
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
	@echo "  make docker-dev                       - Run dockerized dev infra + hot reload services"
	@echo "  make docker-dev-infra                 - Run only dockerized dev infrastructure"
	@echo "  make docker-dev-down                  - Stop dockerized dev infra + services"
	@echo "  make dev-backoffice-seed TENANT_ID=t1 - Seed Consul placement + onboarding connection for one tenant"
	@echo "  make dev-backoffice-sample TENANT_ID=t1 - Seed placement plus sample POD partners/products/orders"
	@echo "  make dev-auth-bootstrap TENANT_ID=t1 - Seed user, tenant membership, session, and token bundle"
	@echo "  make dev-ui-auth-sync                - Copy the dev auth bundle into the UI public assets"
	@echo "  make dev-pod-sample TENANT_ID=t1 - Seed infra, sample business data, and auth bootstrap together"
	@echo "  make gql-backoffice                   - Generate backoffice graphql"
	@echo "  make k8s ENV=${env} SVC=${service}    - Deploy service to k8s dev EG: make k8s ENV="staging" SVC="grpcgateway catalog auth storefront backoffice""
	@echo "  make k8s-ui ENV=${env} SVC=${service} - Deploy service to k8s dev EG: make k8s-ui ENV="staging" SVC="ui-podzone""
