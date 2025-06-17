package main

import (
	"context"
	dbsql "database/sql"
	"log"
	"os"
	"time"

	"kasho/proto"
	"pg-translicator/internal/sql"
	"pg-translicator/internal/transform"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	maxBackoff = 30 * time.Second
)

func connectWithRetry[T any](ctx context.Context, connectFn func() (T, error)) (T, error) {
	var zero T
	backoff := time.Second
	for {
		result, err := connectFn()
		if err == nil {
			return result, nil
		}

		log.Printf("Connection failed: %v", err)
		log.Printf("Retrying in %v...", backoff)
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

func main() {
	// Use hardcoded config directory path - expects mounted /app/config directory
	configFile := "/app/config/transforms.yml"
	
	// Verify config directory exists and is actually a directory
	configDir := "/app/config"
	if stat, err := os.Stat(configDir); os.IsNotExist(err) {
		log.Fatal("Config directory /app/config does not exist. Please mount a config directory to /app/config")
	} else if err != nil {
		log.Fatalf("Error checking config directory: %v", err)
	} else if !stat.IsDir() {
		log.Fatal("/app/config exists but is not a directory. Please mount a config directory to /app/config")
	}
	
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Fatal("Required config file /app/config/transforms.yml not found. Please ensure transforms.yml exists in the mounted config directory")
	}

	config, err := transform.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbConnStr := os.Getenv("REPLICA_DATABASE_URL")
	if dbConnStr == "" {
		log.Fatal("REPLICA_DATABASE_URL environment variable is required")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := connectWithRetry(ctx, func() (*dbsql.DB, error) {
		log.Printf("Connecting to replica database ...")
		db, err := dbsql.Open("postgres", dbConnStr)
		if err != nil {
			return nil, err
		}
		if err := db.Ping(); err != nil {
			db.Close()
			return nil, err
		}
		return db, nil
	})
	if err != nil {
		log.Fatalf("Failed to connect to replica database after retries: %v", err)
	}
	defer db.Close()
	log.Printf("Successfully connected to replica database")

	// Start periodic sequence sync
	syncTicker := time.NewTicker(15 * time.Second)
	defer syncTicker.Stop()

	hasInserts := false

	go func() {
		for {
			select {
			case <-syncTicker.C:
				if hasInserts {
					if err := sql.SyncSequences(ctx, db); err != nil {
						log.Printf("Error during sequence sync: %v", err)
					}
					hasInserts = false
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	serverAddr := os.Getenv("CHANGE_STREAM_SERVICE")
	if serverAddr == "" {
		log.Fatal("CHANGE_STREAM_SERVICE environment variable is required")
	}
	client, err := connectWithRetry(ctx, func() (*grpc.ClientConn, error) {
		log.Printf("Connecting to change stream service ...")
		return grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	})
	if err != nil {
		log.Fatalf("Failed to connect to change stream service after retries: %v", err)
	}
	defer client.Close()
	log.Printf("Successfully connected to change stream service")

	streamClient := proto.NewChangeStreamClient(client)

	for {
		stream, err := streamClient.Stream(ctx, &proto.StreamRequest{LastLsn: ""})
		if err != nil {
			log.Printf("Failed to start stream: %v", err)
			time.Sleep(time.Second)
			continue
		}

		for {
			change, err := stream.Recv()
			if err != nil {
				log.Printf("Error receiving change: %v", err)
				break
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

			if dml := transformedChange.GetDml(); dml != nil && dml.Kind == "insert" {
				hasInserts = true
			}

			log.Printf("%s (%s): %s", change.Lsn, change.Type, stmt)
		}
	}
}
