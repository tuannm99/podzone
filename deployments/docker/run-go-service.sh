#!/bin/sh
set -eu

service="${GO_SERVICE:-}"

if [ -z "$service" ]; then
  echo "GO_SERVICE is required" >&2
  exit 1
fi

case "$service" in
  auth)
    build_cmd='go build -o ./bin/auth ./cmd/auth/main.go'
    build_bin='./bin/auth'
    ;;
  auth-worker)
    build_cmd='go build -o ./bin/auth-worker ./cmd/auth-worker/main.go'
    build_bin='./bin/auth-worker'
    ;;
  iam)
    build_cmd='go build -o ./bin/iam ./cmd/iam/main.go'
    build_bin='./bin/iam'
    ;;
  iam-worker)
    build_cmd='go build -o ./bin/iam-worker ./cmd/iam-worker/main.go'
    build_bin='./bin/iam-worker'
    ;;
  catalog)
    build_cmd='go build -o ./bin/catalog ./cmd/catalog/main.go'
    build_bin='./bin/catalog'
    ;;
  partner)
    build_cmd='go build -o ./bin/partner ./cmd/partner/main.go'
    build_bin='./bin/partner'
    ;;
  onboarding)
    build_cmd='go build -o ./bin/onboarding ./cmd/onboarding/main.go'
    build_bin='./bin/onboarding'
    ;;
  backoffice)
    build_cmd='go build -o ./bin/backoffice ./cmd/backoffice/main.go'
    build_bin='./bin/backoffice'
    ;;
  grpcgateway)
    build_cmd='go build -o ./bin/grpcgateway ./cmd/grpcgateway/main.go'
    build_bin='./bin/grpcgateway'
    ;;
  *)
    echo "Unsupported GO_SERVICE: $service" >&2
    exit 1
    ;;
esac

exec air \
  --build.cmd "$build_cmd" \
  --build.bin "$build_bin" \
  --tmp_dir "tmp/air-$service"
