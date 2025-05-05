FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum* ./

RUN go mod download

COPY cmd/grpcgateway/ ./cmd/grpcgateway/
COPY pkg/ ./pkg/

RUN CGO_ENABLED=0 GOOS=linux go build -o grpcgateway ./cmd/grpcgateway/main.go

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/grpcgateway /app/

EXPOSE 8080

CMD ["/app/grpcgateway"]
