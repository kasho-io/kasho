FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
WORKDIR /app/services/pg-translicator
RUN go mod edit -replace pg-change-stream=../../services/pg-change-stream
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /pg-translicator ./cmd/server

FROM golang:1.24-alpine AS development
WORKDIR /app
RUN go install github.com/air-verse/air@latest
COPY . .
RUN go mod download
WORKDIR /app/services/pg-translicator
RUN go mod edit -replace pg-change-stream=../../services/pg-change-stream
COPY environments/development/.air.toml /app/.air.toml
CMD ["air", "-c", "/app/.air.toml"]

FROM alpine:latest AS production
WORKDIR /app
COPY --from=builder /pg-translicator .
CMD ["./pg-translicator"] 