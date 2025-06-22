package server

import (
	"context"
	"database/sql"
	"fmt"
	
	_ "github.com/lib/pq"
)

// CheckReplicationSlot checks if the kasho_slot exists
func CheckReplicationSlot(ctx context.Context, dbURL string) (bool, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return false, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()
	
	var slotName string
	err = db.QueryRowContext(ctx, `
		SELECT slot_name FROM pg_replication_slots 
		WHERE slot_name = 'kasho_slot'
	`).Scan(&slotName)
	
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to query replication slots: %w", err)
	}
	
	return true, nil
}

// DetermineInitialState determines the initial state based on saved state
func DetermineInitialState(ctx context.Context, dbURL string, savedState *StateInfo) (State, error) {
	// If we have a saved state in Redis, use it
	if savedState != nil {
		return savedState.Current, nil
	}
	
	// No saved state - always start in WAITING state
	// This ensures bootstrap coordination can happen properly
	return StateWaiting, nil
}