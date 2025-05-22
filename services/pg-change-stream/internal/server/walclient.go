package server

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"pg-change-stream/internal/types"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5"
)

type Client struct {
	conn    *pgx.Conn
	slotLSN pglogrepl.LSN
	ticker  *time.Ticker
	done    chan struct{}
	dbURL   string
}

func (c *Client) Connect(ctx context.Context) error {
	walURL := c.dbURL
	if !strings.Contains(walURL, "replication=database") {
		if strings.Contains(walURL, "?") {
			walURL += "&replication=database"
		} else {
			walURL += "?replication=database"
		}
	}

	log.Printf("Connecting to main database...")
	conn, err := pgx.Connect(ctx, c.dbURL)
	if err != nil {
		return err
	}

	// Check replication slot status using main connection
	var slotExists bool
	var active bool
	var restartLSN string
	var confirmedFlushLSN string
	if err := conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_replication_slots WHERE slot_name = 'translicate_slot'), active, restart_lsn, confirmed_flush_lsn FROM pg_replication_slots WHERE slot_name = 'translicate_slot'").Scan(&slotExists, &active, &restartLSN, &confirmedFlushLSN); err != nil {
		conn.Close(ctx)
		return fmt.Errorf("failed to check replication slot: %w", err)
	}
	log.Printf("Replication slot exists: %v, active: %v, restart_lsn: %s, confirmed_flush_lsn: %s", slotExists, active, restartLSN, confirmedFlushLSN)

	if !slotExists {
		conn.Close(ctx)
		return fmt.Errorf("replication slot 'translicate_slot' does not exist")
	}

	log.Printf("Connecting to WAL database...")
	walConn, err := pgx.Connect(ctx, walURL)
	if err != nil {
		conn.Close(ctx)
		return err
	}

	startLSN, err := pglogrepl.ParseLSN("0/0")
	if err != nil {
		conn.Close(ctx)
		walConn.Close(ctx)
		return fmt.Errorf("failed to parse restart LSN: %w", err)
	}

	log.Printf("Starting replication from LSN: %s", startLSN)
	if err := pglogrepl.StartReplication(ctx, walConn.PgConn(), "translicate_slot", startLSN, pglogrepl.StartReplicationOptions{
		Mode:       pglogrepl.LogicalReplication,
		PluginArgs: []string{"proto_version '2'", "publication_names 'translicate_pub'"},
	}); err != nil {
		conn.Close(ctx)
		walConn.Close(ctx)
		return fmt.Errorf("failed to start replication: %w", err)
	}
	log.Printf("Replication started successfully")

	if c.conn != nil {
		c.conn.Close(ctx)
	}
	if c.ticker != nil {
		c.ticker.Stop()
	}
	if c.done != nil {
		close(c.done)
	}
	c.conn = walConn
	c.slotLSN = startLSN
	c.ticker = time.NewTicker(10 * time.Second)
	c.done = make(chan struct{})

	go c.sendStatusUpdates(ctx)
	return nil
}

func (c *Client) ConnectWithRetry(ctx context.Context) error {
	backoff := time.Second
	for {
		log.Printf("Connecting to database...")
		err := c.Connect(ctx)
		if err == nil {
			log.Printf("Successfully connected and started replication")
			return nil
		}

		log.Printf("Connection failed: %v", err)
		log.Printf("Retrying in %v...", backoff)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
		}
	}
}

func NewClient(ctx context.Context, dbURL string) (*Client, error) {
	client := &Client{dbURL: dbURL}
	if err := client.ConnectWithRetry(ctx); err != nil {
		return nil, err
	}
	return client, nil
}

func (c *Client) Close(ctx context.Context) {
	if c.ticker != nil {
		c.ticker.Stop()
	}
	close(c.done)
	if c.conn != nil {
		c.conn.Close(ctx)
	}
}

func (c *Client) sendStatusUpdates(ctx context.Context) {
	for {
		select {
		case <-c.ticker.C:
			if err := pglogrepl.SendStandbyStatusUpdate(ctx, c.conn.PgConn(), pglogrepl.StandbyStatusUpdate{
				WALWritePosition: c.slotLSN,
				WALFlushPosition: c.slotLSN,
				WALApplyPosition: c.slotLSN,
			}); err != nil {
				log.Printf("Error sending status update: %v", err)
				return
			}
		case <-c.done:
			return
		}
	}
}

func (c *Client) ReceiveMessage(ctx context.Context) ([]types.Change, error) {
	msg, err := c.conn.PgConn().ReceiveMessage(ctx)
	if err != nil {
		return nil, err
	}
	changes, lsn, err := ParseMessage(msg)
	if err != nil {
		return nil, err
	}
	if lsn != 0 {
		c.slotLSN = lsn
	}
	return changes, nil
}
