package server

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"pg-change-stream/api"
	"pg-change-stream/internal/types"
)

type ChangeStreamServer struct {
	api.UnimplementedChangeStreamServer
	buffer *RedisBuffer
}

func NewChangeStreamServer(buffer *RedisBuffer) *ChangeStreamServer {
	return &ChangeStreamServer{
		buffer: buffer,
	}
}

func (s *ChangeStreamServer) Stream(req *api.StreamRequest, stream api.ChangeStream_StreamServer) error {
	// Send buffered changes first
	if req.LastLsn != "" {
		changes, err := s.buffer.GetChangesAfter(stream.Context(), req.LastLsn)
		if err != nil {
			return fmt.Errorf("failed to get buffered changes: %w", err)
		}

		for _, change := range changes {
			protoChange := convertToProtoChange(change)
			if err := stream.Send(protoChange); err != nil {
				return err
			}
		}
	}

	// Subscribe to new changes
	pubsub := s.buffer.client.Subscribe(stream.Context(), "pg:changes")
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

func convertToProtoChange(change types.Change) *api.Change {
	protoChange := &api.Change{
		Lsn:  change.LSN,
		Type: change.Type(),
	}

	switch data := change.Data.(type) {
	case *types.DMLData:
		dml := &api.DMLData{
			Table:        data.Table,
			ColumnNames:  data.ColumnNames,
			ColumnValues: convertToStrings(data.ColumnValues),
			Kind:         data.Kind,
		}
		if data.OldKeys != nil {
			dml.OldKeys = &api.OldKeys{
				KeyNames:  data.OldKeys.KeyNames,
				KeyValues: convertToStrings(data.OldKeys.KeyValues),
			}
		}
		protoChange.Data = &api.Change_Dml{Dml: dml}
	case *types.DDLData:
		protoChange.Data = &api.Change_Ddl{
			Ddl: &api.DDLData{
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

func convertToStrings(values []any) []string {
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = fmt.Sprint(v)
	}
	return result
}
