# Multi-stage build for tailperf
FROM golang:1.25.4-alpine AS builder

# Allow Go to download and use newer toolchain versions
ENV GOTOOLCHAIN=auto

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies (includes gvisor override)
RUN go mod download

# Copy source code
COPY *.go ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o tailperf .

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
