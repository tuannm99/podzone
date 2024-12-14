FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git make protoc

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

ARG SERVICE_NAME
RUN make build SERVICE=${SERVICE_NAME}

FROM alpine:3.18

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

ARG SERVICE_NAME
COPY --from=builder /app/services/${SERVICE_NAME}/bin/${SERVICE_NAME} /app/
COPY --from=builder /app/config /app/config

ENV SERVICE_NAME=${SERVICE_NAME}
ENV ENV=production

EXPOSE 8000

CMD ["/app/${SERVICE_NAME}"]
