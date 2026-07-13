# Build stage
FROM golang:1.26-alpine AS builder

ARG VERSION=dev
ARG GIT_COMMIT=unknown

WORKDIR /build

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
COPY . .

# Build da aplicação
RUN BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s \
      -X github.com/lasfh/goapi/examples/api/internal/version.Version=${VERSION} \
      -X github.com/lasfh/goapi/examples/api/internal/version.GitCommit=${GIT_COMMIT} \
      -X github.com/lasfh/goapi/examples/api/internal/version.BuildTime=${BUILD_TIME}" \
    -mod=vendor \
    -o api \
    ./main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /build/api /app/api

# Criar usuário não-root
RUN addgroup -g 1000 nonroot && \
    adduser -D -u 1000 -G nonroot nonroot && \
    chown -R nonroot:nonroot /app

USER nonroot
