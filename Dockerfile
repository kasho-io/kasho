# Kasho Consolidated Container
# Includes all services and tools in a single image for simplified deployment

# Build argument for base image (defaults to local for development)
ARG BASE_IMAGE=kasho-base:latest

# Build the base image first (or use pre-built from registry)
FROM ${BASE_IMAGE} AS builder

# Accept LDFLAGS as build argument
ARG LDFLAGS=""

# Set working directory
WORKDIR /app

# Copy the entire workspace
COPY . .

# Download all dependencies with cache mount
RUN --mount=type=cache,target=/go/pkg/mod go work sync

# Build all services and tools with version information
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "${LDFLAGS}" -o /bin/pg-change-stream ./services/pg-change-stream/cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "${LDFLAGS}" -o /bin/pg-translicator ./services/pg-translicator/cmd/server
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags "${LDFLAGS}" -o /bin/pg-bootstrap-sync ./tools/runtime/pg-bootstrap-sync
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "${LDFLAGS}" -o /bin/env-template ./tools/runtime/env-template

# Development stage with hot reload
FROM ${BASE_IMAGE} AS development

# Accept LDFLAGS as build argument
ARG LDFLAGS=""

WORKDIR /app

# Copy source code and workspace files
COPY . .
COPY environments/development/.air.toml /app/.air.toml

# Download dependencies (using base image's module cache)
RUN go work sync

# Create necessary directories
RUN mkdir -p /app/bin /app/scripts /data/redis

# Build essential tools that are needed immediately with version information
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "${LDFLAGS}" -o /app/bin/env-template ./tools/runtime/env-template
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags "${LDFLAGS}" -o /app/bin/pg-bootstrap-sync ./tools/runtime/pg-bootstrap-sync

# Copy only runtime scripts to scripts directory (same as production)
COPY scripts/runtime/ /app/scripts/
RUN chmod +x /app/scripts/*.sh

# Set default environment variables
ENV KV_URL=redis://127.0.0.1:6379
ENV GO_ENV=development

# Expose common ports
EXPOSE 50051 50052 6379

# Default command starts Redis and can be overridden
CMD ["sh", "-c", "redis-server --daemonize no --protected-mode no --logfile /dev/stdout --dir /data/redis --dbfilename redis.rdb & sleep 2 && tail -f /dev/null"]

# Production stage
FROM alpine:latest AS production

# Install runtime dependencies including trurl for URL parsing
RUN apk add --no-cache ca-certificates tzdata redis postgresql-client bash
# Add trurl from edge/community repository
RUN apk add --no-cache --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community trurl
# Install grpcurl for bootstrap coordination
RUN apk add --no-cache curl && \
    curl -sSL https://github.com/fullstorydev/grpcurl/releases/download/v1.8.9/grpcurl_1.8.9_linux_x86_64.tar.gz | tar -xz -C /usr/local/bin grpcurl && \
    chmod +x /usr/local/bin/grpcurl && \
    apk del curl

# Create app user for security
RUN addgroup -g 1001 -S kasho && \
    adduser -S kasho -u 1001 -G kasho

# Create necessary directories
RUN mkdir -p /app/bin /app/scripts /data/redis /app/config && \
    chown -R kasho:kasho /app /data/redis

WORKDIR /app

# Copy built binaries from builder stage
COPY --from=builder /bin/pg-change-stream /app/bin/
COPY --from=builder /bin/pg-translicator /app/bin/
COPY --from=builder /bin/pg-bootstrap-sync /app/bin/
COPY --from=builder /bin/env-template /app/bin/

# Copy only runtime scripts to scripts directory
COPY scripts/runtime/ /app/scripts/

# Make scripts executable
RUN chmod +x /app/scripts/*.sh

# Copy only necessary SQL files
COPY sql/setup/ /app/sql/setup/

# Copy proto files for grpcurl
COPY proto/ /app/proto/

# Set ownership
RUN chown -R kasho:kasho /app

# Switch to non-root user
USER kasho

# Set default environment variables
ENV KV_URL=redis://127.0.0.1:6379
ENV GO_ENV=production
ENV LOG_LEVEL=info

# Expose common ports
EXPOSE 50051 50052 6379

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD pgrep redis-server || exit 1

# Default command shows help text
# Override with docker run commands to start specific services
CMD ["/app/scripts/kasho-help.sh"]

# Alternative entry points (examples):
# docker run kasho /app/bin/pg-translicator
# docker run kasho /app/bin/pg-bootstrap-sync --help
# docker run kasho /app/bin/env-template