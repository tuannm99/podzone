FROM golang:1.23-alpine AS builder

ARG SERVICE_NAME
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/${SERVICE_NAME}/ cmd/${SERVICE_NAME}/
COPY pkg/ pkg/
COPY services/${SERVICE_NAME}/ services/${SERVICE_NAME}/

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/${SERVICE_NAME} ./cmd/${SERVICE_NAME}/main.go

# -----------------------------------------

FROM alpine:3.18

ARG SERVICE_NAME

RUN apk add --no-cache ca-certificates tzdata curl

WORKDIR /app

COPY --from=builder /app/${SERVICE_NAME} /app/${SERVICE_NAME}

ENV SERVICE_NAME=${SERVICE_NAME}

RUN echo '#!/bin/sh' > /entrypoint.sh && \
    echo 'echo "🚀 Starting service: $SERVICE_NAME"' >> /entrypoint.sh && \
    echo 'exec /app/$SERVICE_NAME' >> /entrypoint.sh && \
    chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

EXPOSE 50051
EXPOSE 8080
