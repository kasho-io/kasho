package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"translicate/internal/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatal("REDIS_URL environment variable is required")
	}

	broker, err := server.NewMessageBroker(redisURL)
	if err != nil {
		log.Fatalf("Failed to create message broker: %v", err)
	}
	defer broker.Close()

	go broker.Run()
	go server.StartServer(broker)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	dbURL := os.Getenv("PRIMARY_DATABASE_URL")
	if dbURL == "" {
		log.Fatal("PRIMARY_DATABASE_URL environment variable is required")
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
				jsonData, err := json.Marshal(change)
				if err != nil {
					log.Printf("Error marshaling change: %v", err)
					continue
				}

				if err := broker.AddChange(ctx, change.LSN, change); err != nil {
					log.Printf("Error storing change in Redis: %v", err)
				}
				broker.Broadcast(jsonData, change.LSN)
			}
		}
	}
}
