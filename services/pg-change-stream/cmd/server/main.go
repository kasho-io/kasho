package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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

	// Create the gRPC server
	changeStreamServer := server.NewChangeStreamServer(buffer)
	
	// Initialize state from Redis and database
	dbURL := os.Getenv("PRIMARY_DATABASE_URL")
	if dbURL == "" {
		log.Fatal("PRIMARY_DATABASE_URL environment variable is required")
	}
	
	state, err := changeStreamServer.LoadState(ctx)
	if err != nil {
		log.Printf("Failed to load state from Redis, using defaults: %v", err)
	}
	
	// Determine initial state based on slot existence
	initialState, err := server.DetermineInitialState(ctx, dbURL, state)
	if err != nil {
		log.Printf("Failed to determine initial state: %v", err)
		initialState = server.StateWaiting
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
	
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
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

	// Start WAL processing goroutine that monitors state changes
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
					log.Println("In STREAMING state, starting WAL client")
					client, err = server.NewClient(ctx, dbURL)
					if err != nil {
						log.Printf("Failed to create WAL client: %v", err)
						continue
					}
				} else if currentState != server.StateStreaming && client != nil {
					log.Println("Not in STREAMING state, closing WAL client")
					client.Close(ctx)
					client = nil
					continue
				}
				
				// If we have a client and we're streaming, process messages
				if client != nil && currentState == server.StateStreaming {
					changes, err := client.ReceiveMessage(ctx)
					if err != nil {
						log.Printf("Error receiving message: %v", err)

						if strings.Contains(err.Error(), "connection") || strings.Contains(err.Error(), "closed") {
							log.Println("Connection lost")
							if err := client.ConnectWithRetry(ctx); err != nil {
								log.Printf("Failed to reconnect: %v", err)
								client = nil
								continue
							}
						}
						continue
					}

					for _, change := range changes {
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
			}
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
}
