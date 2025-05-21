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

	client, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	streamClient := api.NewChangeStreamClient(client)

	stream, err := streamClient.Stream(ctx, &api.StreamRequest{})
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
			stmt, err := sql.ToSQL(dml)
			if err != nil {
				log.Printf("Error generating SQL: %v", err)
				continue
			}
			log.Printf("%s (%s): %s", change.Lsn, change.Type, stmt)
		case *api.Change_Ddl:
			ddl := data.Ddl
			log.Printf("%s (%s): %s", change.Lsn, change.Type, ddl.Ddl)
		default:
			log.Printf("Unexpected data type: %T", data)
		}
	}
}
