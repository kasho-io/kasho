package main

import (
	"context"
	dbsql "database/sql"
	"log"
	"os"
	"strings"

	"pg-change-stream/api"
	"pg-translicator/internal/sql"
	"pg-translicator/internal/transform"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	configFile := os.Getenv("TRANSFORM_CONFIG_FILE")
	if configFile == "" {
		log.Fatal("TRANSFORM_CONFIG_FILE environment variable is required")
	}

	config, err := transform.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbConnStr := os.Getenv("REPLICA_DATABASE_URL")
	if dbConnStr == "" {
		log.Fatal("REPLICA_DATABASE_URL environment variable is required")
	}
	if !strings.Contains(dbConnStr, "sslmode=") {
		if strings.Contains(dbConnStr, "?") {
			dbConnStr += "&sslmode=disable"
		} else {
			dbConnStr += "?sslmode=disable"
		}
	}
	db, err := dbsql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to replica database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping replica database: %v", err)
	}

	ctx := context.Background()
	serverAddr := os.Getenv("CHANGE_STREAM_ADDR")
	if serverAddr == "" {
		serverAddr = "pg-change-stream:8080"
	}
	client, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	streamClient := api.NewChangeStreamClient(client)

	stream, err := streamClient.Stream(ctx, &api.StreamRequest{LastLsn: ""})
	if err != nil {
		log.Fatalf("Failed to start stream: %v", err)
	}

	for {
		change, err := stream.Recv()
		if err != nil {
			log.Fatalf("Error receiving change: %v", err)
		}

		transformedChange, err := transform.TransformChange(config, change)
		if err != nil {
			log.Printf("Error transforming change: %v", err)
			continue
		}

		stmt, err := sql.ToSQL(transformedChange)
		if err != nil {
			log.Printf("Error generating SQL: %v", err)
			continue
		}

		if _, err := db.ExecContext(ctx, stmt); err != nil {
			log.Printf("Error executing SQL: %v", err)
			continue
		}

		log.Printf("%s (%s): %s", change.Lsn, change.Type, stmt)
	}
}
