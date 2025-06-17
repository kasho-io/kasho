# Kasho Consolidated Container
# Includes all services and tools in a single image for simplified deployment

FROM golang:1.24-alpine AS builder

# Install build dependencies including C compiler for CGO
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

# Set working directory
WORKDIR /app

# Copy the entire workspace
COPY . .

# Download all dependencies with cache mount
RUN --mount=type=cache,target=/go/pkg/mod go work sync

# Build all services and tools
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/pg-change-stream ./services/pg-change-stream/cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/pg-translicator ./services/pg-translicator/cmd/server
RUN CGO_ENABLED=1 GOOS=linux go build -o /bin/pg-bootstrap-sync ./tools/pg-bootstrap-sync
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/env-template ./tools/env-template

# Development stage with hot reload
FROM golang:1.24-alpine AS development

# Install development dependencies including C compiler for CGO and trurl for URL parsing
RUN apk add --no-cache git ca-certificates tzdata redis gcc musl-dev
# Add trurl from edge/community repository
RUN apk add --no-cache --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community trurl
RUN go install github.com/air-verse/air@latest

WORKDIR /app

# Copy source code and workspace files
COPY . .
COPY environments/development/.air.toml /app/.air.toml

# Download dependencies with cache mount
RUN --mount=type=cache,target=/go/pkg/mod go work sync

# Build essential tools that are needed immediately
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/env-template ./tools/env-template
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/pg-bootstrap-sync ./tools/pg-bootstrap-sync

# Copy utility scripts to /app/ for consistency with production stage
COPY scripts/parse-db-url.sh /app/
COPY scripts/env-template-wrapper.sh /app/
COPY scripts/kasho-help.sh /app/

# Create data directory for Redis
RUN mkdir -p /data/redis

# Set default environment variables
ENV KV_URL=redis://127.0.0.1:6379
ENV GO_ENV=development

# Expose common ports
EXPOSE 8080 6379

# Default command starts Redis and can be overridden
CMD ["sh", "-c", "redis-server --daemonize no --protected-mode no --logfile /dev/stdout --dir /data/redis --dbfilename redis.rdb & sleep 2 && tail -f /dev/null"]

# Production stage
FROM alpine:latest AS production

# Install runtime dependencies including trurl for URL parsing
RUN apk add --no-cache ca-certificates tzdata redis
# Add trurl from edge/community repository
RUN apk add --no-cache --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community trurl

# Create app user for security
RUN addgroup -g 1001 -S kasho && \
    adduser -S kasho -u 1001 -G kasho

# Create necessary directories
RUN mkdir -p /app /data/redis /app/config && \
    chown -R kasho:kasho /app /data/redis

WORKDIR /app

# Copy built binaries from builder stage
COPY --from=builder /bin/pg-change-stream /app/
COPY --from=builder /bin/pg-translicator /app/
COPY --from=builder /bin/pg-bootstrap-sync /app/
COPY --from=builder /bin/env-template /app/

# Copy utility scripts
COPY scripts/parse-db-url.sh /app/
COPY scripts/env-template-wrapper.sh /app/
COPY scripts/kasho-help.sh /app/

# Copy configuration files and documentation
COPY environments/demo/config /app/config/demo/
COPY environments/development/config /app/config/development/
COPY README.md /app/
COPY docs/ /app/docs/

# Set ownership
RUN chown -R kasho:kasho /app

# Switch to non-root user
USER kasho

# Set default environment variables
ENV KV_URL=redis://127.0.0.1:6379
ENV GO_ENV=production
ENV LOG_LEVEL=info

# Expose common ports
EXPOSE 8080 6379

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD pgrep redis-server || exit 1

# Default command shows help text
# Override with docker run commands to start specific services
CMD ["./kasho-help.sh"]

# Alternative entry points (examples):
# docker run kasho ./pg-translicator
# docker run kasho ./pg-bootstrap-sync --help
# docker run kasho ./env-template