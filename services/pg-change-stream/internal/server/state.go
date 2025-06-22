package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// State represents the operational state of the change stream
type State int

const (
	StateWaiting State = iota
	StateAccumulating
	StateStreaming
)

func (s State) String() string {
	switch s {
	case StateWaiting:
		return "WAITING"
	case StateAccumulating:
		return "ACCUMULATING"
	case StateStreaming:
		return "STREAMING"
	default:
		return "UNKNOWN"
	}
}

// StateInfo holds the persistent state information
type StateInfo struct {
	Current            State     `json:"current"`
	StartLSN           string    `json:"start_lsn,omitempty"`
	TransitionTime     time.Time `json:"transition_time"`
	AccumulatedChanges int64     `json:"accumulated_changes"`
}

const stateKey = "kasho:change-stream:state"

// LoadState loads the state from Redis
func (s *ChangeStreamServer) LoadState(ctx context.Context) (*StateInfo, error) {
	data, err := s.buffer.Get(ctx, stateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}
	
	if data == "" {
		// No state found, default to waiting
		return &StateInfo{
			Current:        StateWaiting,
			TransitionTime: time.Now(),
		}, nil
	}
	
	var state StateInfo
	if err := json.Unmarshal([]byte(data), &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}
	
	return &state, nil
}

// SaveState saves the state to Redis
func (s *ChangeStreamServer) SaveState(ctx context.Context, state *StateInfo) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}
	
	if err := s.buffer.Set(ctx, stateKey, string(data)); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}
	
	return nil
}

// TransitionState updates the state and saves it
func (s *ChangeStreamServer) TransitionState(ctx context.Context, newState State, startLSN string) error {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	
	oldState := s.state.Current
	s.state.Current = newState
	s.state.TransitionTime = time.Now()
	
	if startLSN != "" {
		s.state.StartLSN = startLSN
	}
	
	if err := s.SaveState(ctx, s.state); err != nil {
		// Rollback on error
		s.state.Current = oldState
		return err
	}
	
	return nil
}