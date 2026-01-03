# cool-kit Dockerfile
# Multi-stage build with security best practices
#
# Build: docker build -t cool-kit .
# Run:   docker run --rm cool-kit --help

# =============================================================================
# Build arguments
# =============================================================================
ARG GO_VERSION=1.25
ARG UBUNTU_VERSION=3.23
ARG BUILD_VERSION=dev
ARG BUILD_DATE=unknown
ARG GIT_COMMIT=unknown

# =============================================================================
# Builder stage
# =============================================================================
FROM golang:${GO_VERSION}-ubuntu${UBUNTU_VERSION} AS builder

# Install build dependencies
RUN apk add --no-cache \
    ca-certificates \
    git \
    make \
    tzdata

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build arguments
ARG BUILD_VERSION
ARG BUILD_DATE
ARG GIT_COMMIT

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w \
        -X main.version=${BUILD_VERSION} \
        -X main.date=${BUILD_DATE} \
        -X main.commit=${GIT_COMMIT}" \
    -o /cool-kit \
    .

# =============================================================================
# Runtime stage (minimal)
# =============================================================================
FROM ubuntu:${UBUNTU_VERSION} AS runtime

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    git \
    tzdata \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1000 app \
    && adduser -u 1000 -G app -s /bin/sh -D app

# Copy binary from builder
COPY --from=builder /cool-kit /usr/local/bin/cool-kit

# Ensure binary is executable
RUN chmod +x /usr/local/bin/cool-kit

# Switch to non-root user
USER app

# Set working directory
WORKDIR /workspace

# Build arguments for labels
ARG BUILD_VERSION
ARG BUILD_DATE
ARG GIT_COMMIT

# OCI labels
LABEL org.opencontainers.image.title="cool-kit" \
      org.opencontainers.image.description="CLI toolkit for Coolify deployments" \
      org.opencontainers.image.version="${BUILD_VERSION}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.revision="${GIT_COMMIT}" \
      org.opencontainers.image.source="https://github.com/entro314-labs/cool-kit" \
      org.opencontainers.image.url="https://github.com/entro314-labs/cool-kit" \
      org.opencontainers.image.vendor="entro314-labs" \
      org.opencontainers.image.licenses="MIT"

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/usr/local/bin/cool-kit", "--version"]

# Default entrypoint and command
ENTRYPOINT ["/usr/local/bin/cool-kit"]
CMD ["--help"]
