FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /pg-change-stream ./cmd/server

FROM golang:1.24-alpine AS development
WORKDIR /app
RUN go install github.com/air-verse/air@latest
RUN apk add --no-cache redis
COPY . .
RUN go mod download
WORKDIR /app/services/pg-change-stream
COPY environments/development/.air.toml /app/.air.toml
ENV KV_URL=redis://127.0.0.1:6379
CMD ["sh", "-c", "redis-server --daemonize no --protected-mode no --logfile /dev/stdout & sleep 2 && air -c /app/.air.toml"]

FROM alpine:latest AS alpine
WORKDIR /app
COPY --from=builder /pg-change-stream .
RUN apk add --no-cache redis
EXPOSE 8080
ENV KV_URL=redis://127.0.0.1:6379
CMD ["sh", "-c", "redis-server --daemonize no --protected-mode no --logfile /dev/stdout & sleep 2 && ./pg-change-stream"] 