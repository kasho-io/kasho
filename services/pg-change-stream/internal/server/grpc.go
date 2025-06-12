package server

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"kasho/pkg/kvbuffer"
	"kasho/pkg/types"
	"kasho/proto"
)

type ChangeStreamServer struct {
	proto.UnimplementedChangeStreamServer
	buffer *kvbuffer.KVBuffer
}

func NewChangeStreamServer(buffer *kvbuffer.KVBuffer) *ChangeStreamServer {
	return &ChangeStreamServer{
		buffer: buffer,
	}
}

func (s *ChangeStreamServer) Stream(req *proto.StreamRequest, stream proto.ChangeStream_StreamServer) error {
	// Send buffered changes first
	if req.LastLsn != "" {
		rawChanges, err := s.buffer.GetChangesAfter(stream.Context(), req.LastLsn)
		if err != nil {
			return fmt.Errorf("failed to get buffered changes: %w", err)
		}

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
