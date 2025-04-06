FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum* ./

RUN go mod download

COPY cmd/auth/ ./cmd/auth/
COPY pkg/ ./pkg/
COPY services/auth/ ./services/auth/

RUN CGO_ENABLED=0 GOOS=linux go build -o auth ./cmd/auth/main.go

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/auth /app/

EXPOSE 50051
EXPOSE 8080

CMD ["/app/auth"]
