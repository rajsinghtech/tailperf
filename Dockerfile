# Multi-stage build for tailperf
FROM golang:1.25.4-alpine AS builder

# Allow Go to download and use newer toolchain versions
ENV GOTOOLCHAIN=auto

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies with cache mount (includes gvisor override)
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY *.go ./

# Build the binary with cache mounts for faster rebuilds
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -trimpath -o tailperf .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates iptables ip6tables

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/tailperf .

# Create state directory
RUN mkdir -p /var/lib/tailperf

# Run as non-root when possible (but Tailscale needs some permissions)
# USER nobody

ENTRYPOINT ["/app/tailperf"]
