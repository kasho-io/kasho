package server

import (
	"context"
	"encoding/json"
	"testing"

	"kasho/proto"
	"pg-change-stream/internal/types"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
)

func TestKVBuffer_AddChange(t *testing.T) {
	// Create mock Redis client
	db, mock := redismock.NewClientMock()
	defer db.Close()

	kvBuffer := &KVBuffer{client: db}

	ctx := context.Background()
	lsn := "0/100"
	change := types.Change{
		LSN: lsn,
		Data: types.DMLData{
			Table:       "users",
			Kind:        "insert",
			ColumnNames: []string{"id", "name"},
			ColumnValues: []types.ColumnValueWrapper{
				{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: 1}}},
				{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: "test"}}},
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
	mock.ExpectPublish("pg:changes", data).SetVal(1)

	// Test AddChange
	err := kvBuffer.AddChange(ctx, lsn, change)
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
	change := types.Change{
		LSN: invalidLSN,
		Data: types.DMLData{
			Table: "users",
			Kind:  "insert",
		},
	}

	err := kvBuffer.AddChange(ctx, invalidLSN, change)
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
	change1 := types.Change{
		LSN: "0/200",
		Data: types.DMLData{
			Table:       "users",
			Kind:        "insert",
			ColumnNames: []string{"id"},
			ColumnValues: []types.ColumnValueWrapper{
				{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: 1}}},
			},
		},
	}
	change2 := types.Change{
		LSN: "0/300",
		Data: types.DMLData{
			Table:       "users",
			Kind:        "update",
			ColumnNames: []string{"name"},
			ColumnValues: []types.ColumnValueWrapper{
				{ColumnValue: &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: "updated"}}},
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

	// Test GetChangesAfter
	changes, err := kvBuffer.GetChangesAfter(ctx, lsn)
	if err != nil {
		t.Errorf("GetChangesAfter() error = %v", err)
	}

	if len(changes) != 2 {
		t.Errorf("Expected 2 changes, got %d", len(changes))
	}

	// Verify first change
	if changes[0].LSN != "0/200" {
		t.Errorf("Expected LSN 0/200, got %s", changes[0].LSN)
	}
	dml1, ok := changes[0].Data.(*types.DMLData)
	if !ok {
		t.Errorf("Expected DMLData, got %T", changes[0].Data)
	} else if dml1.Kind != "insert" {
		t.Errorf("Expected insert, got %s", dml1.Kind)
	}

	// Verify second change
	if changes[1].LSN != "0/300" {
		t.Errorf("Expected LSN 0/300, got %s", changes[1].LSN)
	}
	dml2, ok := changes[1].Data.(*types.DMLData)
	if !ok {
		t.Errorf("Expected DMLData, got %T", changes[1].Data)
	} else if dml2.Kind != "update" {
		t.Errorf("Expected update, got %s", dml2.Kind)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectations were not met: %v", err)
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

	changes, err := kvBuffer.GetChangesAfter(ctx, lsn)
	if err != nil {
		t.Errorf("GetChangesAfter() error = %v", err)
	}

	if len(changes) != 0 {
		t.Errorf("Expected 0 changes, got %d", len(changes))
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