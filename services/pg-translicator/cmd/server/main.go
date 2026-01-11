package main

import (
	"context"
	dbsql "database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kasho/pkg/version"
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
	log.Printf("pg-translicator version %s (commit: %s, built: %s)",
		version.Version, version.GitCommit, version.BuildDate)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	// Set session replication role to 'replica' to prevent triggers from firing
	// This mimics physical replication behavior where triggers exist but don't execute
	if _, err := db.Exec("SET session_replication_role = 'replica'"); err != nil {
		log.Fatalf("Failed to set session_replication_role: %v", err)
	}
	log.Printf("Set session_replication_role to 'replica' - triggers will not fire during replication")

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

	serverAddr := os.Getenv("CHANGE_STREAM_SERVICE_ADDR")
	if serverAddr == "" {
		log.Fatal("CHANGE_STREAM_SERVICE_ADDR environment variable is required")
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

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	// Main replication loop
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Check if replica database has any user tables to determine starting LSN
				lastLsn := determineStartingLSN(db)
				log.Printf("Starting stream from LSN: %s", lastLsn)

				stream, err := streamClient.Stream(ctx, &proto.StreamRequest{LastLsn: lastLsn})
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

					// Debug: Check if transform was applied
					if dml := change.GetDml(); dml != nil && dml.Table == "users" {
						transformedDml := transformedChange.GetDml()
						if transformedDml != nil && len(transformedDml.ColumnNames) > 0 {
							// Find password column index
							for i, col := range transformedDml.ColumnNames {
								if col == "password" && i < len(dml.ColumnValues) && i < len(transformedDml.ColumnValues) {
									origPwd := "nil"
									transPwd := "nil"
									if dml.ColumnValues[i] != nil {
										origPwd = fmt.Sprintf("%v", dml.ColumnValues[i].GetStringValue())[:20] + "..."
									}
									if transformedDml.ColumnValues[i] != nil {
										transPwd = fmt.Sprintf("%v", transformedDml.ColumnValues[i].GetStringValue())[:20] + "..."
									}
									log.Printf("Transform debug - users table password: original=%s, transformed=%s", origPwd, transPwd)
									break
								}
							}
						}
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
	}()

	// Wait for shutdown
	<-ctx.Done()
	log.Println("Shutting down pg-translicator")
}

// determineStartingLSN checks if the replica has any user tables
// Returns "0/0" if empty (needs bootstrap), or "" if tables exist
func determineStartingLSN(db *dbsql.DB) string {
	// Check for user tables (excluding system schemas)
	var tableCount int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
		AND table_type = 'BASE TABLE'
	`).Scan(&tableCount)

	if err != nil {
		log.Printf("Error checking replica tables: %v, assuming empty", err)
		return "0/0"
	}

	if tableCount == 0 {
		log.Printf("Replica database is empty, will request all changes from beginning")
		return "0/0"
	}

	log.Printf("Replica database has %d tables, will only request new changes", tableCount)
	return ""
}
