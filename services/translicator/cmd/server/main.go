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
	"translicator/internal/dialect"
	"translicator/internal/sql"
	"translicator/internal/transform"

	_ "github.com/go-sql-driver/mysql"
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
	log.Printf("translicator version %s (commit: %s, built: %s)",
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

	// Determine the dialect from the connection string
	dbDialect, err := dialect.FromConnectionString(dbConnStr)
	if err != nil {
		log.Fatalf("Failed to determine database dialect: %v", err)
	}
	log.Printf("Using %s dialect", dbDialect.Name())

	// Create SQL generator with the detected dialect
	sqlGenerator := sql.NewSQLGenerator(dbDialect)

	// Convert connection string to driver-specific DSN format
	dsn := dbDialect.FormatDSN(dbConnStr)

	db, err := connectWithRetry(ctx, func() (*dbsql.DB, error) {
		log.Printf("Connecting to replica database ...")
		db, err := dbsql.Open(dbDialect.GetDriverName(), dsn)
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

	// Set up connection for replication (dialect-specific)
	if err := dbDialect.SetupConnection(db); err != nil {
		log.Fatalf("Failed to set up connection: %v", err)
	}
	log.Printf("Connection setup complete for %s dialect", dbDialect.Name())

	// Start periodic sequence/auto-increment sync
	syncTicker := time.NewTicker(15 * time.Second)
	defer syncTicker.Stop()

	hasInserts := false

	go func() {
		for {
			select {
			case <-syncTicker.C:
				if hasInserts {
					if err := dbDialect.SyncSequences(ctx, db); err != nil {
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
				// Check if replica database has any user tables to determine starting position
				lastPosition := determineStartingPosition(db, dbDialect)
				log.Printf("Starting stream from position: %s", lastPosition)

				stream, err := streamClient.Stream(ctx, &proto.StreamRequest{LastPosition: lastPosition})
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

					stmt, err := sqlGenerator.ToSQL(transformedChange)
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

					log.Printf("%s (%s): %s", change.Position, change.Type, stmt)
				}
			}
		}
	}()

	// Wait for shutdown
	<-ctx.Done()
	log.Println("Shutting down translicator")
}

// determineStartingPosition checks if the replica has any user tables
// Returns "bootstrap" if empty (needs bootstrap), or "" if tables exist
func determineStartingPosition(db *dbsql.DB, dbDialect dialect.Dialect) string {
	// Check for user tables (using dialect-specific query)
	var tableCount int
	err := db.QueryRow(dbDialect.GetUserTablesQuery()).Scan(&tableCount)

	if err != nil {
		log.Printf("Error checking replica tables: %v, assuming empty", err)
		return "bootstrap"
	}

	if tableCount == 0 {
		log.Printf("Replica database is empty, will request all changes from beginning")
		return "bootstrap"
	}

	log.Printf("Replica database has %d tables, will only request new changes", tableCount)
	return ""
}
