package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"kasho/pkg/kvbuffer"
	"kasho/proto"
	"pg-change-stream/internal/server"

	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kvURL := os.Getenv("KV_URL")
	if kvURL == "" {
		log.Fatal("KV_URL environment variable is required")
	}

	buffer, err := kvbuffer.NewKVBuffer(kvURL)
	if err != nil {
		log.Fatalf("Failed to create KV buffer: %v", err)
	}
	defer buffer.Close()

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	proto.RegisterChangeStreamServer(s, server.NewChangeStreamServer(buffer))
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		s.GracefulStop()
		cancel()
	}()

	dbUser := os.Getenv("PRIMARY_DATABASE_KASHO_USER")
	dbPassword := os.Getenv("PRIMARY_DATABASE_KASHO_PASSWORD")
	dbHost := os.Getenv("PRIMARY_DATABASE_HOST")
	dbPort := os.Getenv("PRIMARY_DATABASE_PORT")
	dbName := os.Getenv("PRIMARY_DATABASE_DB")

	if dbUser == "" || dbPassword == "" || dbHost == "" || dbPort == "" || dbName == "" {
		log.Fatal("All database environment variables (PRIMARY_DATABASE_*) are required")
	}

	dbURL := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
	if sslMode := os.Getenv("PRIMARY_DATABASE_SSLMODE"); sslMode != "" {
		dbURL += fmt.Sprintf("?sslmode=%s", sslMode)
	}

	client, err := server.NewClient(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			changes, err := client.ReceiveMessage(ctx)
			if err != nil {
				log.Printf("Error receiving message: %v", err)

				if strings.Contains(err.Error(), "connection") || strings.Contains(err.Error(), "closed") {
					log.Println("Connection lost")
					if err := client.ConnectWithRetry(ctx); err != nil {
						log.Printf("Failed to reconnect: %v", err)
						return
					}
					continue
				}
				continue
			}

			for _, change := range changes {
				// Store change in KV buffer
				if err := buffer.AddChange(ctx, change); err != nil {
					log.Printf("Error storing change in KV: %v", err)
				}
			}
		}
	}
}
