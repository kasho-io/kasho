package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"pg-change-stream/api"
	"pg-translicator/internal/sql"
	"pg-translicator/internal/transform"

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

	// Get last LSN from command line if provided
	lastLSN := ""
	if len(os.Args) > 2 {
		lastLSN = os.Args[2]
	}

	stream, err := streamClient.Stream(ctx, &api.StreamRequest{LastLsn: lastLSN})
	if err != nil {
		log.Fatalf("Failed to start stream: %v", err)
	}

	for {
		change, err := stream.Recv()
		if err != nil {
			log.Fatalf("Error receiving change: %v", err)
		}

		switch data := change.Data.(type) {
		case *api.Change_Dml:
			dml := data.Dml
			// Apply transformations to column values if configured
			for i, col := range dml.ColumnNames {
				transformed, err := transform.GetFakeValue(config, dml.Table, col, dml.ColumnValues[i])
				if err == nil && transformed != nil {
					// Only update if transformation was successful and returned a value
					dml.ColumnValues[i] = fmt.Sprintf("%v", transformed)
				} else if err != nil {
					log.Printf("Error transforming %s.%s: %v", dml.Table, col, err)
				}
			}
			stmt, err := sql.ToSQL(change)
			if err != nil {
				log.Printf("Error generating SQL: %v", err)
				continue
			}
			log.Printf("%s (%s): %s", change.Lsn, change.Type, stmt)
		case *api.Change_Ddl:
			stmt, err := sql.ToSQL(change)
			if err != nil {
				log.Printf("Error generating SQL: %v", err)
				continue
			}
			log.Printf("%s (%s): %s", change.Lsn, change.Type, stmt)
		default:
			log.Printf("Unexpected data type: %T", data)
		}
	}
}
