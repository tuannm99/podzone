FROM golang:1.23-alpine AS builder

ARG SERVICE_NAME

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/${SERVICE_NAME}/ ./cmd/${SERVICE_NAME}/
COPY pkg/ ./pkg/
COPY services/${SERVICE_NAME}/ ./services/${SERVICE_NAME}/

RUN CGO_ENABLED=0 GOOS=linux go build -o auth ./cmd/${SERVICE_NAME}/main.go

FROM alpine:3.18

RUN apk add --no-cache ca-certificates tzdata telnet curl

WORKDIR /app

COPY --from=builder /app/${SERVICE_NAME} /app/

ENV SERVICE_NAME=${SERVICE_NAME}

EXPOSE 50051
EXPOSE 8080

CMD ["/app/${SERVICE_NAME}"]
