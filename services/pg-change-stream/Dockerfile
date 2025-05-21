FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /pg-change-stream ./cmd/server

FROM golang:1.24-alpine AS development
WORKDIR /app
RUN go install github.com/air-verse/air@latest
COPY . .
RUN go mod download
WORKDIR /app/services/pg-change-stream
COPY environments/development/.air.toml /app/.air.toml
CMD ["air", "-c", "/app/.air.toml"]

FROM alpine:latest
WORKDIR /app
COPY --from=builder /pg-change-stream .
EXPOSE 8080
CMD ["./pg-change-stream"] 