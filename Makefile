.PHONY: all proto swagger build test coverage lint fmt vulncheck dev down clean help
.PHONY: docker-dev docker-dev-infra docker-dev-down mocks mocks-gen
.PHONY: dev-backoffice-seed dev-backoffice-sample dev-kv-store-refresh dev-onboarding-reconcile-tenant
.PHONY: dev-auth-bootstrap
.PHONY: dev-ui-auth-sync dev-pod-sample dev-pod-up

GO := go
GO_CACHE ?= /tmp/podzone-gocache
BUF_CACHE_DIR ?= /tmp/podzone-buf-cache
COVERAGE_MIN ?= 90
DEV_OWNER_ID_OUTPUT ?= /tmp/podzone-dev-owner-id

LINT_PKGS = $(shell GOCACHE=$(GO_CACHE) $(GO) list -f '{{.Dir}}' ./... | \
	sed '\#/generated$$#d; \#/mocks$$#d; \#/node_modules/#d; s#$(CURDIR)#.#')
COVER_PKGS = $(shell GOCACHE=$(GO_CACHE) $(GO) list -f '{{.Dir}}' ./internal/... ./pkg/... | \
	sed '\#/generated$$#d; \#/generated/#d; \#/mocks$$#d; \#/node_modules/#d; \
		\#/internal/bootstrap$$#d; \#/pkg/api/proto/#d; s#$(CURDIR)#.#')

# Colors for output
COLOR_RESET = \033[0m
COLOR_GREEN = \033[32m
COLOR_YELLOW = \033[33m

SVC ?= none
.PHONY: svc help

proto:
	@echo "Generating protobuf via buf..."
	@BUF_CACHE_DIR=$(BUF_CACHE_DIR) GOCACHE=$(GO_CACHE) $(GO) tool buf generate

proto_lint:
	@BUF_CACHE_DIR=$(BUF_CACHE_DIR) GOCACHE=$(GO_CACHE) $(GO) tool buf lint

proto_svc:
	@echo "$(COLOR_GREEN)Generating protobuf for $(SVC)...$(COLOR_RESET)"
	@BUF_CACHE_DIR=$(BUF_CACHE_DIR) GOCACHE=$(GO_CACHE) $(GO) tool buf generate --path api/proto/$(SVC)


# Run tests for all packages
test:
	@echo "$(COLOR_GREEN)Running tests...$(COLOR_RESET)"
	@set -e; \
	export PODZONE_TC_REUSE=1; \
	GOCACHE=$(GO_CACHE) $(GO) test ./... -v -cover

coverage:
	@echo "$(COLOR_GREEN)Checking test coverage...$(COLOR_RESET)"
	@GOCACHE=$(GO_CACHE) $(GO) test $(COVER_PKGS) \
		-covermode=atomic -coverprofile=/tmp/podzone-coverage.out
	@coverage=$$(GOCACHE=$(GO_CACHE) $(GO) tool cover \
		-func=/tmp/podzone-coverage.out | awk '/^total:/ {gsub("%", "", $$3); print $$3}'); \
	echo "coverage: $${coverage}% (minimum: $(COVERAGE_MIN)%)"; \
	awk -v coverage="$$coverage" -v minimum="$(COVERAGE_MIN)" \
		'BEGIN { if (coverage + 0 < minimum + 0) exit 1 }'

# Run linter
lint:
	@echo "$(COLOR_GREEN)Running linter...$(COLOR_RESET)"
	GOCACHE=$(GO_CACHE) $(GO) tool golangci-lint run --timeout=5m $(LINT_PKGS)

vulncheck:
	@echo "$(COLOR_GREEN)Running govulncheck...$(COLOR_RESET)"
	GOCACHE=$(GO_CACHE) $(GO) tool govulncheck ./internal/... ./pkg/...

fmt:
	@echo "$(COLOR_GREEN)Formatting Go code...$(COLOR_RESET)"
	@$(GO) tool golines -w --max-len=120 .
	@$(GO) tool gofumpt -w .

docker-dev-infra:
	docker compose -f deployments/docker/infras.yml up -d

comma := ,
PROFILE ?= full

docker-dev:
	PROFILE=$(PROFILE) docker compose $(foreach p,$(subst $(comma), ,$(PROFILE)),--profile $(p)) \
		-f deployments/docker/infras.yml -f deployments/docker/services.yml up --build

dev-pod-up:
	@sh scripts/dev/run_local_pod_dev.sh "$(TENANT_ID)" "$(STORE_NAME)" "$(STORE_SUBDOMAIN)" "$(DEV_USERNAME)" "$(DEV_EMAIL)" "$(DEV_PASSWORD)"

docker-dev-down:
	docker compose --profile full \
		-f deployments/docker/infras.yml -f deployments/docker/services.yml down

dev-backoffice-seed:
	@DB_NAME=$(DB_NAME) \
	SCHEMA_NAME=$(SCHEMA_NAME) \
	PG_HOST=$(PG_HOST) \
	PG_PORT=$(PG_PORT) \
	PG_USER=$(PG_USER) \
	PG_PASSWORD=$(PG_PASSWORD) \
	PG_SSL_MODE=$(PG_SSL_MODE) \
	CREATE_STORE=$(CREATE_STORE) \
	STORE_OWNER_ID=$(STORE_OWNER_ID) \
	sh scripts/dev/seed_backoffice_tenant.sh "$(TENANT_ID)" "$(STORE_NAME)" "$(STORE_SUBDOMAIN)" "$(ONBOARDING_URL)"

dev-kv-store-refresh:
	@MONGO_URI=$(MONGO_URI) \
	PG_HOST=$(PG_HOST) \
	PG_PORT=$(PG_PORT) \
	PG_USER=$(PG_USER) \
	PG_PASSWORD=$(PG_PASSWORD) \
	PG_SSL_MODE=$(PG_SSL_MODE) \
	sh scripts/dev/refresh_kv_store_from_onboarding.sh

dev-onboarding-reconcile-tenant:
	@TENANT_ID=$(TENANT_ID) \
	OWNER_ID=$(OWNER_ID) \
	STORE_NAME="$(STORE_NAME)" \
	STORE_SUBDOMAIN=$(STORE_SUBDOMAIN) \
	ONBOARDING_URL=$(ONBOARDING_URL) \
	ONBOARDING_SERVICE_TOKEN=$(ONBOARDING_SERVICE_TOKEN) \
	GOCACHE=$(GO_CACHE) \
	$(GO) run scripts/dev/reconcile_legacy_tenant.go

dev-backoffice-sample:
	@DB_NAME=$(DB_NAME) \
	SCHEMA_NAME=$(SCHEMA_NAME) \
	PG_HOST=$(PG_HOST) \
	PG_PORT=$(PG_PORT) \
	PG_USER=$(PG_USER) \
	PG_PASSWORD=$(PG_PASSWORD) \
	PG_SSL_MODE=$(PG_SSL_MODE) \
	CREATE_STORE=$(CREATE_STORE) \
	STORE_OWNER_ID=$(STORE_OWNER_ID) \
	sh scripts/dev/seed_backoffice_tenant.sh "$(TENANT_ID)" "$(STORE_NAME)" "$(STORE_SUBDOMAIN)" "$(ONBOARDING_URL)"
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
	DEV_OWNER_ID_OUTPUT=$(DEV_OWNER_ID_OUTPUT) \
	$(GO) run ./scripts/dev/seed_auth_bootstrap.go

dev-ui-auth-sync:
	@AUTH_BOOTSTRAP_OUTPUT=$(AUTH_BOOTSTRAP_OUTPUT) \
	UI_AUTH_BOOTSTRAP_TARGET=$(UI_AUTH_BOOTSTRAP_TARGET) \
	sh scripts/dev/sync_ui_auth_bootstrap.sh

dev-pod-sample:
	@$(MAKE) dev-auth-bootstrap \
		"TENANT_ID=$(TENANT_ID)" \
		"TENANT_NAME=$(TENANT_NAME)" \
		"TENANT_SLUG=$(TENANT_SLUG)" \
		"DEV_USERNAME=$(DEV_USERNAME)" \
		"DEV_EMAIL=$(DEV_EMAIL)" \
		"DEV_PASSWORD=$(DEV_PASSWORD)" \
		"DEV_FULL_NAME=$(DEV_FULL_NAME)" \
		"PG_HOST=$(PG_HOST)" \
		"PG_PORT=$(PG_PORT)" \
		"PG_USER=$(PG_USER)" \
		"PG_PASSWORD=$(PG_PASSWORD)" \
		"PG_SSL_MODE=$(PG_SSL_MODE)" \
		"JWT_SECRET=$(JWT_SECRET)" \
		"JWT_KEY=$(JWT_KEY)" \
		"AUTH_BOOTSTRAP_OUTPUT=$(AUTH_BOOTSTRAP_OUTPUT)" \
		"DEV_OWNER_ID_OUTPUT=$(DEV_OWNER_ID_OUTPUT)"
	@store_owner_id=$$(cat "$(DEV_OWNER_ID_OUTPUT)"); \
	$(MAKE) dev-backoffice-sample \
		"TENANT_ID=$(TENANT_ID)" \
		"STORE_NAME=$(STORE_NAME)" \
		"STORE_SUBDOMAIN=$(STORE_SUBDOMAIN)" \
		"TENANT_NAME=$(TENANT_NAME)" \
		"TENANT_SLUG=$(TENANT_SLUG)" \
		"CLUSTER_NAME=$(CLUSTER_NAME)" \
		"DB_NAME=$(DB_NAME)" \
		"SCHEMA_NAME=$(SCHEMA_NAME)" \
		"PG_HOST=$(PG_HOST)" \
		"PG_PORT=$(PG_PORT)" \
		"PG_USER=$(PG_USER)" \
		"PG_PASSWORD=$(PG_PASSWORD)" \
		"PG_SSL_MODE=$(PG_SSL_MODE)" \
		"ONBOARDING_URL=$(ONBOARDING_URL)" \
		"CREATE_STORE=$(CREATE_STORE)" \
		"STORE_OWNER_ID=$$store_owner_id"
	@$(MAKE) dev-ui-auth-sync \
		"AUTH_BOOTSTRAP_OUTPUT=$(AUTH_BOOTSTRAP_OUTPUT)" \
		"UI_AUTH_BOOTSTRAP_TARGET=$(UI_AUTH_BOOTSTRAP_TARGET)"

gql-backoffice:
	$(GO) tool gqlgen generate


mocks-gen: mocks
	@echo "$(COLOR_GREEN)Generating mocks via mockery...$(COLOR_RESET)"
	@GOCACHE=$(GO_CACHE) $(GO) tool mockery

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
	export GOOSE_DBSTRING="postgres://postgres:postgres@localhost:5432/auth?sslmode=disable"; \
	$(GO) tool goose -dir ./internal/auth/migrations/authsql status; \
	$(GO) tool goose -dir ./internal/auth/migrations/authsql up
# 	$(GO) tool goose -dir ./internal/auth/migrations/authsql down

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
	@echo "📦 Building and deploying frontend..."
	set -a; source .env; set +a; \
	docker build -t tuannm99/podzone-frontend:$(ENV) \
		--build-arg VITE_ADMIN_API_URL=$$VITE_ADMIN_API_URL \
		-f Dockerfile-ui .; \
	docker push tuannm99/podzone-frontend:$(ENV); \
	kubectl delete -f deployments/kubernetes/$(ENV)/services/frontend.yml --ignore-not-found; \
	kubectl apply -f deployments/kubernetes/$(ENV)/services/frontend.yml

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
	@echo "  make coverage                         - Enforce aggregate coverage (default: $(COVERAGE_MIN)%)"
	@echo "  make lint                             - Run linter"
	@echo "  make fmt                              - Format Go code with golines and gofumpt"
	@echo "  make mocks-gen                        - Generate mocks with the pinned mockery tool"
	@echo "  make portfw                           - Portfowrding"
	@echo "  make dev SVC=${service}               - Run service"
	@echo "  make docker-dev                           - Run full stack (PROFILE=full)"
	@echo "  make docker-dev PROFILE=iam               - Run iam profile only"
	@echo "  make docker-dev PROFILE=backoffice,iam    - Run multiple profiles"
	@echo "  make docker-dev PROFILE=frontend-v2       - Run frontend-v2 (Angular) standalone dev server only"
	@echo "  make docker-dev-infra                     - Run only dockerized dev infrastructure"
	@echo "  make docker-dev-down                      - Stop dockerized dev infra + services"
	@echo "  make dev-pod-up TENANT_ID=t1          - Start local docker stack and auto-bootstrap tenant/sample/auth"
	@echo "  make dev-backoffice-seed TENANT_ID=t1 - Create and provision one onboarding store"
	@echo "  make dev-kv-store-refresh             - Rebuild Mongo runtime KV from onboarding allocations"
	@echo "  make dev-onboarding-reconcile-tenant  - Enroll one legacy IAM tenant through onboarding"
	@echo "  make dev-backoffice-sample TENANT_ID=t1 - Seed placement plus sample POD partners/products/orders"
	@echo "  make dev-auth-bootstrap TENANT_ID=t1 - Seed user, tenant membership, session, and token bundle"
	@echo "  make dev-ui-auth-sync                - Copy the dev auth bundle into the UI public assets"
	@echo "  make dev-pod-sample TENANT_ID=t1 - Seed infra, sample business data, and auth bootstrap together"
	@echo "  make gql-backoffice                   - Generate backoffice graphql"
	@echo "  make k8s ENV=${env} SVC=${service}    - Deploy service to k8s dev EG: make k8s ENV="staging" SVC="grpcgateway catalog auth storefront backoffice""
	@echo "  make k8s-ui ENV=${env}                - Deploy frontend to k8s EG: make k8s-ui ENV=staging"
