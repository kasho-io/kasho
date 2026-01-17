package kvbuffer

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
)

// Test types - simulating the types from pg-change-stream
type TestChange struct {
	LSN  string      `json:"lsn"`
	Data TestDMLData `json:"data"`
}

func (c TestChange) Type() string {
	return "dml"
}

func (c TestChange) GetLSN() string {
	return c.LSN
}

type TestDMLData struct {
	Table       string                  `json:"table"`
	Kind        string                  `json:"kind"`
	ColumnNames []string                `json:"columnnames"`
	ColumnValues []TestColumnValueWrapper `json:"columnvalues"`
}

func (d TestDMLData) Type() string {
	return "dml"
}

type TestColumnValueWrapper struct {
	Value interface{} `json:"value"`
}

func TestKVBuffer_AddChange(t *testing.T) {
	// Create mock Redis client
	db, mock := redismock.NewClientMock()
	defer db.Close()

	kvBuffer := &KVBuffer{client: db}

	ctx := context.Background()
	lsn := "0/100"
	change := TestChange{
		LSN: lsn,
		Data: TestDMLData{
			Table:       "users",
			Kind:        "insert",
			ColumnNames: []string{"id", "name"},
			ColumnValues: []TestColumnValueWrapper{
				{Value: 1},
				{Value: "test"},
			},
		},
	}

	// Marshal change for expectations
	data, _ := json.Marshal(change)

	// Set expectations
	mock.ExpectZAdd(changesKey, redis.Z{
		Score:  float64(256), // LSN 0/100 = 256
		Member: data,
	}).SetVal(1)
	mock.ExpectExpire(changesKey, changesTTL).SetVal(true)
	mock.ExpectPublish(changesChannel, data).SetVal(1)

	// Test AddChange
	err := kvBuffer.AddChange(ctx, change)
	if err != nil {
		t.Errorf("AddChange() error = %v", err)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectations were not met: %v", err)
	}
}

func TestKVBuffer_AddChange_BootstrapLSN(t *testing.T) {
	// Create mock Redis client
	db, mock := redismock.NewClientMock()
	defer db.Close()

	kvBuffer := &KVBuffer{client: db}

	ctx := context.Background()
	lsn := "0/BOOTSTRAP00000001"
	change := TestChange{
		LSN: lsn,
		Data: TestDMLData{
			Table: "users",
			Kind:  "insert",
		},
	}

	// Marshal change for expectations
	data, _ := json.Marshal(change)

	// Set expectations - bootstrap LSNs get negative scores
	mock.ExpectZAdd(changesKey, redis.Z{
		Score:  float64(-999999), // -1000000 + 1
		Member: data,
	}).SetVal(1)
	mock.ExpectExpire(changesKey, changesTTL).SetVal(true)
	mock.ExpectPublish(changesChannel, data).SetVal(1)

	// Test AddChange
	err := kvBuffer.AddChange(ctx, change)
	if err != nil {
		t.Errorf("AddChange() error = %v", err)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectations were not met: %v", err)
	}
}

func TestKVBuffer_AddChange_InvalidLSN(t *testing.T) {
	db, _ := redismock.NewClientMock()
	defer db.Close()

	kvBuffer := &KVBuffer{client: db}

	ctx := context.Background()
	invalidLSN := "invalid"
	change := TestChange{
		LSN: invalidLSN,
		Data: TestDMLData{
			Table: "users",
			Kind:  "insert",
		},
	}

	err := kvBuffer.AddChange(ctx, change)
	if err == nil {
		t.Errorf("Expected error for invalid LSN, got nil")
	}
}

func TestKVBuffer_GetChangesAfter(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	kvBuffer := &KVBuffer{client: db}

	ctx := context.Background()
	lsn := "0/100"

	// Create test changes
	change1 := TestChange{
		LSN: "0/200",
		Data: TestDMLData{
			Table:       "users",
			Kind:        "insert",
			ColumnNames: []string{"id"},
			ColumnValues: []TestColumnValueWrapper{
				{Value: 1},
			},
		},
	}
	change2 := TestChange{
		LSN: "0/300",
		Data: TestDMLData{
			Table:       "users",
			Kind:        "update",
			ColumnNames: []string{"name"},
			ColumnValues: []TestColumnValueWrapper{
				{Value: "updated"},
			},
		},
	}

	data1, _ := json.Marshal(change1)
	data2, _ := json.Marshal(change2)

	// Set expectations
	mock.ExpectZRangeByScore(changesKey, &redis.ZRangeBy{
		Min:    "(256", // Exclude LSN 0/100
		Max:    "+inf",
		Offset: 0,
		Count:  1000,
	}).SetVal([]string{string(data1), string(data2)})

	// Test GetChangesAfter - now returns []json.RawMessage
	rawChanges, err := kvBuffer.GetChangesAfter(ctx, lsn)
	if err != nil {
		t.Errorf("GetChangesAfter() error = %v", err)
	}

	if len(rawChanges) != 2 {
		t.Errorf("Expected 2 changes, got %d", len(rawChanges))
	}

	// Verify the raw JSON messages are correct
	if string(rawChanges[0]) != string(data1) {
		t.Errorf("First change JSON mismatch")
	}
	if string(rawChanges[1]) != string(data2) {
		t.Errorf("Second change JSON mismatch")
	}
}

func TestKVBuffer_GetChangesAfter_EmptyResult(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	kvBuffer := &KVBuffer{client: db}

	ctx := context.Background()
	lsn := "0/100"

	// Set expectations for empty result
	mock.ExpectZRangeByScore(changesKey, &redis.ZRangeBy{
		Min:    "(256",
		Max:    "+inf",
		Offset: 0,
		Count:  1000,
	}).SetVal([]string{})

	rawChanges, err := kvBuffer.GetChangesAfter(ctx, lsn)
	if err != nil {
		t.Errorf("GetChangesAfter() error = %v", err)
	}

	if len(rawChanges) != 0 {
		t.Errorf("Expected 0 changes, got %d", len(rawChanges))
	}
}

func TestKVBuffer_Close(t *testing.T) {
	db, mock := redismock.NewClientMock()
	kvBuffer := &KVBuffer{client: db}

	// Redis client's Close() is called internally
	err := kvBuffer.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Verify expectations
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectations were not met: %v", err)
	}
}

func TestNewKVBuffer_ValidURL(t *testing.T) {
	// Test with a valid Redis URL format
	validURL := "redis://localhost:6379/0"
	
	// Since NewKVBuffer tries to connect to Redis, and we don't have a real Redis instance,
	// this test will fail on connection. We're testing the URL parsing part.
	_, err := NewKVBuffer(validURL)
	
	// We expect a connection error, not a URL parsing error
	if err == nil {
		// If no error, that means Redis was actually available
		t.Log("NewKVBuffer succeeded (Redis was available)")
	} else if err.Error() == "failed to parse KV URL: invalid redis URL scheme: " {
		t.Errorf("NewKVBuffer() failed to parse valid URL: %v", err)
	} else {
		// Expected connection error
		t.Logf("NewKVBuffer() failed with expected connection error: %v", err)
	}
}

func TestNewKVBuffer_InvalidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "invalid scheme",
			url:  "http://localhost:6379",
		},
		{
			name: "malformed URL",
			url:  "not-a-url",
		},
		{
			name: "empty URL",
			url:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewKVBuffer(tt.url)
			if err == nil {
				t.Errorf("NewKVBuffer() expected error for invalid URL %s, got nil", tt.url)
			}
			
			// Check that it's a URL parsing error
			if !strings.Contains(err.Error(), "failed to parse KV URL") {
				t.Errorf("NewKVBuffer() expected URL parsing error, got: %v", err)
			}
		})
	}
}

func TestNewKVBuffer_ConnectionTimeout(t *testing.T) {
	// Test with a URL that will timeout (non-existent host)
	timeoutURL := "redis://non-existent-host:6379/0"
	
	_, err := NewKVBuffer(timeoutURL)
	if err == nil {
		t.Error("NewKVBuffer() expected connection error for non-existent host, got nil")
	}
	
	// Check that it's a connection error
	if !strings.Contains(err.Error(), "failed to connect to KV") {
		t.Errorf("NewKVBuffer() expected connection error, got: %v", err)
	}
}

func TestParsePositionToScore(t *testing.T) {
	kvBuffer := &KVBuffer{}

	tests := []struct {
		name     string
		position string
		expected float64
		wantErr  bool
	}{
		{
			name:     "valid PostgreSQL LSN",
			position: "0/100",
			expected: 256, // 0x100 = 256
			wantErr:  false,
		},
		{
			name:     "bootstrap position",
			position: "0/BOOTSTRAP00000001",
			expected: -999999, // -1000000 + 1
			wantErr:  false,
		},
		{
			name:     "bootstrap position with higher sequence",
			position: "0/BOOTSTRAP00000123",
			expected: -999877, // -1000000 + 123
			wantErr:  false,
		},
		{
			name:    "invalid position format",
			position: "invalid",
			wantErr: true,
		},
		{
			name:    "malformed bootstrap position",
			position: "0/BOOTSTRAPinvalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, err := kvBuffer.parsePositionToScore(tt.position)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePositionToScore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && score != tt.expected {
				t.Errorf("parsePositionToScore() = %v, want %v", score, tt.expected)
			}
		})
	}
}

func TestParsePositionToScore_MySQLBinlog(t *testing.T) {
	buffer := &KVBuffer{}

	tests := []struct {
		name     string
		position string
		wantErr  bool
	}{
		{"mysql binlog basic", "mysql-bin.000001:4", false},
		{"mysql binlog large file", "mysql-bin.000100:1234567", false},
		{"mysql binlog zero offset", "mysql-bin.000001:0", false},
		{"binlog variant", "binlog.000001:100", false},
		{"invalid no colon", "mysql-bin.000001", true},
		{"invalid no file number", "mysql-bin:100", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, err := buffer.parsePositionToScore(tt.position)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for position %s", tt.position)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error for position %s: %v", tt.position, err)
				return
			}
			if score <= 0 {
				t.Errorf("expected positive score for position %s, got %f", tt.position, score)
			}
		})
	}
}

func TestParsePositionToScore_MySQLOrdering(t *testing.T) {
	buffer := &KVBuffer{}

	// Test that positions are ordered correctly
	positions := []string{
		"mysql-bin.000001:4",
		"mysql-bin.000001:100",
		"mysql-bin.000001:1000",
		"mysql-bin.000002:4",
		"mysql-bin.000002:100",
	}

	var lastScore float64 = -1
	for _, pos := range positions {
		score, err := buffer.parsePositionToScore(pos)
		if err != nil {
			t.Errorf("unexpected error for position %s: %v", pos, err)
			continue
		}
		if score <= lastScore {
			t.Errorf("position %s (score %f) should be > previous score %f", pos, score, lastScore)
		}
		lastScore = score
	}
}