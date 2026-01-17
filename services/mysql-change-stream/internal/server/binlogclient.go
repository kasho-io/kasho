package server

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"kasho/pkg/kvbuffer"
	"kasho/pkg/types"

	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
)

const (
	maxBackoff = 30 * time.Second
)

// Client manages the MySQL binlog replication connection
type Client struct {
	canal         *canal.Canal
	buffer        *kvbuffer.KVBuffer
	changeServer  *ChangeStreamServer
	dbURL         string
	done          chan struct{}
	mu            sync.Mutex
	currentPos    mysql.Position
	changeChan    chan types.Change
	ready         chan struct{} // signals when canal is ready to receive events
	wg            sync.WaitGroup // tracks the canal goroutine
}

// EventHandler implements the canal.EventHandler interface
type EventHandler struct {
	client *Client
}

func (h *EventHandler) OnRow(e *canal.RowsEvent) error {
	pos := h.client.GetPosition()
	changes := RowsEventToChanges(e, pos)
	for _, change := range changes {
		select {
		case h.client.changeChan <- change:
		case <-h.client.done:
			return fmt.Errorf("client closed")
		}
	}
	return nil
}

func (h *EventHandler) OnDDL(header *replication.EventHeader, nextPos mysql.Position, queryEvent *replication.QueryEvent) error {
	h.client.SetPosition(nextPos)
	change := QueryEventToChange(header, queryEvent, nextPos)
	if change != nil {
		select {
		case h.client.changeChan <- *change:
		case <-h.client.done:
			return fmt.Errorf("client closed")
		}
	}
	return nil
}

func (h *EventHandler) OnRotate(header *replication.EventHeader, e *replication.RotateEvent) error {
	pos := mysql.Position{
		Name: string(e.NextLogName),
		Pos:  uint32(e.Position),
	}
	h.client.SetPosition(pos)
	log.Printf("Binlog rotated to %s:%d", pos.Name, pos.Pos)
	return nil
}

func (h *EventHandler) OnTableChanged(header *replication.EventHeader, schema string, table string) error {
	return nil
}

func (h *EventHandler) OnGTID(header *replication.EventHeader, gtidEvent mysql.BinlogGTIDEvent) error {
	return nil
}

func (h *EventHandler) OnPosSynced(header *replication.EventHeader, pos mysql.Position, set mysql.GTIDSet, force bool) error {
	h.client.SetPosition(pos)
	return nil
}

func (h *EventHandler) OnXID(header *replication.EventHeader, nextPos mysql.Position) error {
	return nil
}

func (h *EventHandler) OnRowsQueryEvent(e *replication.RowsQueryEvent) error {
	return nil
}

func (h *EventHandler) String() string {
	return "KashoEventHandler"
}

// parseMySQLURL parses a MySQL connection URL into canal.Config fields
// Supports: mysql://user:pass@host:port/database
func parseMySQLURL(dbURL string) (host string, port uint16, user, password, database string, err error) {
	u, err := url.Parse(dbURL)
	if err != nil {
		return "", 0, "", "", "", fmt.Errorf("failed to parse URL: %w", err)
	}

	if u.Scheme != "mysql" {
		return "", 0, "", "", "", fmt.Errorf("expected mysql:// scheme, got %s://", u.Scheme)
	}

	host = u.Hostname()
	portStr := u.Port()
	if portStr == "" {
		port = 3306
	} else {
		p, err := strconv.Atoi(portStr)
		if err != nil {
			return "", 0, "", "", "", fmt.Errorf("invalid port: %s", portStr)
		}
		port = uint16(p)
	}

	user = u.User.Username()
	password, _ = u.User.Password()
	database = strings.TrimPrefix(u.Path, "/")

	return host, port, user, password, database, nil
}

// NewClient creates a new MySQL binlog replication client
func NewClient(ctx context.Context, dbURL string, buffer *kvbuffer.KVBuffer, changeServer *ChangeStreamServer) (*Client, error) {
	client := &Client{
		dbURL:        dbURL,
		buffer:       buffer,
		changeServer: changeServer,
		done:         make(chan struct{}),
		changeChan:   make(chan types.Change, 1000),
		ready:        make(chan struct{}),
	}

	if err := client.ConnectWithRetry(ctx); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) Connect(ctx context.Context) error {
	host, port, user, password, database, err := parseMySQLURL(c.dbURL)
	if err != nil {
		return fmt.Errorf("failed to parse database URL: %w", err)
	}

	cfg := canal.NewDefaultConfig()
	cfg.Addr = fmt.Sprintf("%s:%d", host, port)
	cfg.User = user
	cfg.Password = password
	cfg.Flavor = "mysql"
	cfg.ServerID = 1001 // Unique server ID for this replica
	cfg.Dump.ExecutionPath = "" // Disable mysqldump (we use bootstrap-sync instead)
	cfg.Dump.DiscardErr = true

	// Include only the specific database
	if database != "" {
		cfg.IncludeTableRegex = []string{fmt.Sprintf("%s\\..*", database)}
	}

	log.Printf("Connecting to MySQL at %s:%d as %s", host, port, user)

	canalInstance, err := canal.NewCanal(cfg)
	if err != nil {
		return fmt.Errorf("failed to create canal: %w", err)
	}

	// Set the event handler
	handler := &EventHandler{client: c}
	canalInstance.SetEventHandler(handler)

	// Start from the beginning or from saved position
	c.mu.Lock()
	startPos := c.currentPos
	if c.canal != nil {
		c.canal.Close()
	}
	c.canal = canalInstance
	c.mu.Unlock()

	if startPos.Name == "" {
		// Get current binlog position to start from
		pos, err := canalInstance.GetMasterPos()
		if err != nil {
			return fmt.Errorf("failed to get master position: %w", err)
		}
		startPos = pos
		log.Printf("Starting from current position: %s:%d", startPos.Name, startPos.Pos)
	} else {
		log.Printf("Resuming from position: %s:%d", startPos.Name, startPos.Pos)
	}

	// Run canal in a goroutine with proper synchronization
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		// Signal that the canal goroutine has started and is ready to process events
		close(c.ready)
		if err := canalInstance.RunFrom(startPos); err != nil {
			// Only log if not a clean shutdown
			select {
			case <-c.done:
				// Clean shutdown, don't log error
			default:
				log.Printf("Canal error: %v", err)
			}
		}
	}()

	// Wait for the canal goroutine to signal it's ready
	select {
	case <-c.ready:
		log.Printf("MySQL binlog replication started")
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (c *Client) ConnectWithRetry(ctx context.Context) error {
	backoff := time.Second
	for {
		log.Printf("Connecting to MySQL database...")
		err := c.Connect(ctx)
		if err == nil {
			log.Printf("Successfully connected and started binlog replication")
			return nil
		}

		log.Printf("Connection failed: %v", err)
		log.Printf("Retrying in %v...", backoff)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

func (c *Client) Close(ctx context.Context) {
	close(c.done)
	c.mu.Lock()
	if c.canal != nil {
		c.canal.Close()
		c.canal = nil
	}
	c.mu.Unlock()

	// Wait for canal goroutine to finish
	c.wg.Wait()
}

func (c *Client) GetPosition() mysql.Position {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.currentPos
}

func (c *Client) SetPosition(pos mysql.Position) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentPos = pos
}

// Changes returns the channel of changes
func (c *Client) Changes() <-chan types.Change {
	return c.changeChan
}

// Done returns a channel that is closed when the client is closed
func (c *Client) Done() <-chan struct{} {
	return c.done
}
