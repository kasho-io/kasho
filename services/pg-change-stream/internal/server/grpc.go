package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"kasho/pkg/kvbuffer"
	"kasho/pkg/types"
	"kasho/proto"
)

type ChangeStreamServer struct {
	proto.UnimplementedChangeStreamServer
	buffer           *kvbuffer.KVBuffer
	state            *StateInfo
	stateMu          sync.RWMutex
	connectedClients int32
	clientsMu        sync.Mutex
	startTime        time.Time
}

func NewChangeStreamServer(buffer *kvbuffer.KVBuffer) *ChangeStreamServer {
	return &ChangeStreamServer{
		buffer:    buffer,
		startTime: time.Now(),
		state: &StateInfo{
			Current:        StateWaiting,
			TransitionTime: time.Now(),
		},
	}
}

// SetState sets the state
func (s *ChangeStreamServer) SetState(state *StateInfo) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	s.state = state
}

// GetState returns the current state
func (s *ChangeStreamServer) GetState() State {
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()
	return s.state.Current
}

// IncrementAccumulated increments the accumulated change count
func (s *ChangeStreamServer) IncrementAccumulated() {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	s.state.AccumulatedChanges++
}

func (s *ChangeStreamServer) Stream(req *proto.StreamRequest, stream proto.ChangeStream_StreamServer) error {
	// Check if we're in streaming state
	s.stateMu.RLock()
	currentState := s.state.Current
	s.stateMu.RUnlock()
	
	if currentState != StateStreaming {
		// Block until we're in streaming state
		for {
			s.stateMu.RLock()
			if s.state.Current == StateStreaming {
				s.stateMu.RUnlock()
				break
			}
			s.stateMu.RUnlock()
			
			select {
			case <-stream.Context().Done():
				return stream.Context().Err()
			case <-time.After(100 * time.Millisecond):
				// Check again
			}
		}
	}
	
	// Track connected clients
	s.clientsMu.Lock()
	s.connectedClients++
	s.clientsMu.Unlock()
	defer func() {
		s.clientsMu.Lock()
		s.connectedClients--
		s.clientsMu.Unlock()
	}()
	
	// Send buffered changes first in batches
	if req.LastLsn != "" {
		const batchSize = 1000
		offset := int64(0)
		
		for {
			rawChanges, err := s.buffer.GetChangesAfterBatch(stream.Context(), req.LastLsn, offset, batchSize)
			if err != nil {
				return fmt.Errorf("failed to get buffered changes: %w", err)
			}

			// Send this batch
			for _, rawChange := range rawChanges {
				var change types.Change
				if err := json.Unmarshal(rawChange, &change); err != nil {
					log.Printf("Error unmarshaling buffered change: %v", err)
					continue
				}
				
				protoChange := convertToProtoChange(change)
				if err := stream.Send(protoChange); err != nil {
					return err
				}
			}
			
			// If we got fewer than batchSize results, we're done
			if len(rawChanges) < batchSize {
				break
			}
			
			offset += batchSize
		}
	}

	// Subscribe to new changes
	pubsub := s.buffer.Subscribe(stream.Context(), "pg:changes")
	defer pubsub.Close()

	// Keep the connection open and wait for new changes
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case msg := <-pubsub.Channel():
			var change types.Change
			if err := json.Unmarshal([]byte(msg.Payload), &change); err != nil {
				log.Printf("Error unmarshaling change: %v", err)
				continue
			}
			
			protoChange := convertToProtoChange(change)
			if err := stream.Send(protoChange); err != nil {
				return err
			}
		}
	}
}

func convertToProtoChange(change types.Change) *proto.Change {
	protoChange := &proto.Change{
		Lsn:  change.GetLSN(),
		Type: change.Type(),
	}

	switch data := change.Data.(type) {
	case *types.DMLData:
		dml := &proto.DMLData{
			Table:        data.Table,
			ColumnNames:  data.ColumnNames,
			ColumnValues: make([]*proto.ColumnValue, len(data.ColumnValues)),
			Kind:         data.Kind,
		}
		for i, cv := range data.ColumnValues {
			dml.ColumnValues[i] = cv.ColumnValue
		}
		if data.OldKeys != nil {
			dml.OldKeys = &proto.OldKeys{
				KeyNames:  data.OldKeys.KeyNames,
				KeyValues: make([]*proto.ColumnValue, len(data.OldKeys.KeyValues)),
			}
			for i, cv := range data.OldKeys.KeyValues {
				dml.OldKeys.KeyValues[i] = cv.ColumnValue
			}
		}
		protoChange.Data = &proto.Change_Dml{Dml: dml}
	case *types.DDLData:
		protoChange.Data = &proto.Change_Ddl{
			Ddl: &proto.DDLData{
				Id:       int32(data.ID),
				Time:     data.Time.Format(time.RFC3339),
				Username: data.Username,
				Database: data.Database,
				Ddl:      data.DDL,
			},
		}
	}

	return protoChange
}

// StartBootstrap begins the accumulation phase for bootstrap
func (s *ChangeStreamServer) StartBootstrap(ctx context.Context, req *proto.StartBootstrapRequest) (*proto.BootstrapResponse, error) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	
	// Validate we're in WAITING state
	if s.state.Current != StateWaiting {
		return &proto.BootstrapResponse{
			Status:        "error",
			PreviousState: s.state.Current.String(),
			CurrentState:  s.state.Current.String(),
		}, fmt.Errorf("can only start bootstrap from WAITING state, current state: %s", s.state.Current)
	}
	
	// Transition to ACCUMULATING
	s.state.Current = StateAccumulating
	s.state.StartLSN = req.StartLsn
	s.state.TransitionTime = time.Now()
	s.state.AccumulatedChanges = 0
	
	if err := s.SaveState(ctx, s.state); err != nil {
		// Rollback
		s.state.Current = StateWaiting
		return nil, fmt.Errorf("failed to save state: %w", err)
	}
	
	return &proto.BootstrapResponse{
		Status:             "started",
		PreviousState:      "WAITING",
		CurrentState:       "ACCUMULATING",
		AccumulatedChanges: 0,
		ReadyToStream:      false,
	}, nil
}

// CompleteBootstrap transitions from accumulating to streaming
func (s *ChangeStreamServer) CompleteBootstrap(ctx context.Context, req *proto.CompleteBootstrapRequest) (*proto.BootstrapResponse, error) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	
	// Validate we're in ACCUMULATING state
	if s.state.Current != StateAccumulating {
		return &proto.BootstrapResponse{
			Status:        "error",
			PreviousState: s.state.Current.String(),
			CurrentState:  s.state.Current.String(),
		}, fmt.Errorf("can only complete bootstrap from ACCUMULATING state, current state: %s", s.state.Current)
	}
	
	// Transition to STREAMING
	previousState := s.state.Current.String()
	s.state.Current = StateStreaming
	s.state.TransitionTime = time.Now()
	
	if err := s.SaveState(ctx, s.state); err != nil {
		// Rollback
		s.state.Current = StateAccumulating
		return nil, fmt.Errorf("failed to save state: %w", err)
	}
	
	return &proto.BootstrapResponse{
		Status:             "completed",
		PreviousState:      previousState,
		CurrentState:       "STREAMING",
		AccumulatedChanges: s.state.AccumulatedChanges,
		ReadyToStream:      true,
	}, nil
}

// GetStatus returns the current status of the change stream
func (s *ChangeStreamServer) GetStatus(ctx context.Context, req *proto.GetStatusRequest) (*proto.StatusResponse, error) {
	s.stateMu.RLock()
	currentState := s.state.Current.String()
	startLSN := s.state.StartLSN
	accumulated := s.state.AccumulatedChanges
	s.stateMu.RUnlock()
	
	s.clientsMu.Lock()
	clients := s.connectedClients
	s.clientsMu.Unlock()
	
	uptime := int64(time.Since(s.startTime).Seconds())
	
	// Get current LSN from WAL client if available
	currentLSN := ""
	// TODO: Get from WAL client when integrated
	
	return &proto.StatusResponse{
		State:              currentState,
		StartLsn:           startLSN,
		CurrentLsn:         currentLSN,
		AccumulatedChanges: accumulated,
		ConnectedClients:   clients,
		UptimeSeconds:      uptime,
	}, nil
}
