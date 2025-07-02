package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"kasho/pkg/version"
	"kasho/proto"
	"kasho/services/licensing/internal/server"
)

func main() {
	log.Printf("Starting Kasho Licensing Service v%s (commit: %s, built: %s)",
		version.Version, version.GitCommit, version.BuildDate)

	// Create gRPC server
	licenseServer, err := server.New()
	if err != nil {
		log.Fatalf("Failed to create license server: %v", err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterLicenseServer(grpcServer, licenseServer)

	// Start listening
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50053"
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down server...")
		grpcServer.GracefulStop()
	}()

	log.Printf("Listening on port %s", port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}