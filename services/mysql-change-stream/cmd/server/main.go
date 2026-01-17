package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kasho/pkg/kvbuffer"
	"kasho/pkg/version"
	"kasho/proto"
	"mysql-change-stream/internal/server"

	"google.golang.org/grpc"
)

func main() {
	log.Printf("mysql-change-stream version %s (commit: %s, built: %s)",
		version.Version, version.GitCommit, version.BuildDate)

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

	// Create the gRPC server
	changeStreamServer := server.NewChangeStreamServer(buffer)

	// Initialize state from Redis
	dbURL := os.Getenv("PRIMARY_DATABASE_URL")
	if dbURL == "" {
		log.Fatal("PRIMARY_DATABASE_URL environment variable is required")
	}

	state, err := changeStreamServer.LoadState(ctx)
	if err != nil {
		log.Printf("Failed to load state from Redis, using defaults: %v", err)
	}

	// Determine initial state
	// For MySQL, we don't have replication slots to check, so we rely on saved state
	initialState := server.StateWaiting
	if state != nil && state.Current == server.StateStreaming {
		initialState = server.StateStreaming
	}

	if state == nil {
		state = &server.StateInfo{
			Current:        initialState,
			TransitionTime: time.Now(),
		}
	} else {
		state.Current = initialState
	}

	changeStreamServer.SetState(state)
	log.Printf("Starting in %s state", state.Current)

	// Get gRPC port from environment or use default
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", port, err)
	}
	log.Printf("gRPC server listening on port %s", port)
	s := grpc.NewServer()
	proto.RegisterChangeStreamServer(s, changeStreamServer)
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

	// Start binlog processing goroutine that monitors state changes
	go func() {
		var client *server.Client
		var err error

		for {
			select {
			case <-ctx.Done():
				if client != nil {
					client.Close(ctx)
				}
				return
			case <-time.After(1 * time.Second):
				currentState := changeStreamServer.GetState()

				// If we're in STREAMING state but don't have a client, create one
				if currentState == server.StateStreaming && client == nil {
					log.Println("In STREAMING state, starting binlog client")
					client, err = server.NewClient(ctx, dbURL, buffer, changeStreamServer)
					if err != nil {
						log.Printf("Failed to create binlog client: %v", err)
						continue
					}

					// Start goroutine to process changes from the client
					go func(c *server.Client) {
						for {
							select {
							case <-ctx.Done():
								return
							case change, ok := <-c.Changes():
								if !ok {
									return
								}
								// Store change in KV buffer
								if err := buffer.AddChange(ctx, change); err != nil {
									log.Printf("Error storing change in KV: %v", err)
								}

								// Update accumulated count if in ACCUMULATING state
								if changeStreamServer.GetState() == server.StateAccumulating {
									changeStreamServer.IncrementAccumulated()
								}
							}
						}
					}(client)
				} else if currentState != server.StateStreaming && client != nil {
					log.Println("Not in STREAMING state, closing binlog client")
					client.Close(ctx)
					client = nil
				}
			}
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("Shutting down mysql-change-stream")
}
