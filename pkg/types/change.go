package types

import (
	"encoding/json"
	"fmt"
	"time"

	"kasho/proto"
)

type ColumnValueWrapper struct {
	*proto.ColumnValue
}

func (cv ColumnValueWrapper) MarshalJSON() ([]byte, error) {
	if cv.ColumnValue == nil {
		return json.Marshal(nil)
	}
	switch v := cv.Value.(type) {
	case *proto.ColumnValue_StringValue:
		return json.Marshal(v.StringValue)
	case *proto.ColumnValue_IntValue:
		return json.Marshal(v.IntValue)
	case *proto.ColumnValue_FloatValue:
		return json.Marshal(v.FloatValue)
	case *proto.ColumnValue_BoolValue:
		return json.Marshal(v.BoolValue)
	case *proto.ColumnValue_TimestampValue:
		return json.Marshal(v.TimestampValue)
	default:
		return nil, fmt.Errorf("unknown column value type: %T", v)
	}
}

func (cv *ColumnValueWrapper) UnmarshalJSON(data []byte) error {
	if cv.ColumnValue == nil {
		cv.ColumnValue = &proto.ColumnValue{}
	}

	// Try string first
	var strVal string
	if err := json.Unmarshal(data, &strVal); err == nil {
		cv.Value = &proto.ColumnValue_StringValue{StringValue: strVal}
		return nil
	}

	// Try int
	var intVal int64
	if err := json.Unmarshal(data, &intVal); err == nil {
		cv.Value = &proto.ColumnValue_IntValue{IntValue: intVal}
		return nil
	}

	// Try float
	var floatVal float64
	if err := json.Unmarshal(data, &floatVal); err == nil {
		cv.Value = &proto.ColumnValue_FloatValue{FloatValue: floatVal}
		return nil
	}

	// Try bool
	var boolVal bool
	if err := json.Unmarshal(data, &boolVal); err == nil {
		cv.Value = &proto.ColumnValue_BoolValue{BoolValue: boolVal}
		return nil
	}

	// Try timestamp
	var timeVal time.Time
	if err := json.Unmarshal(data, &timeVal); err == nil {
		cv.Value = &proto.ColumnValue_TimestampValue{TimestampValue: timeVal.Format(time.RFC3339)}
		return nil
	}

	return fmt.Errorf("failed to unmarshal column value: %s", string(data))
}

type Change struct {
	Position string
	Data     interface {
		Type() string
	}
}

func (c Change) Type() string {
	return c.Data.Type()
}

func (c Change) GetPosition() string {
	return c.Position
}

func (c Change) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(c.Data)
	if err != nil {
		return nil, err
	}

	return json.Marshal(struct {
		Type     string          `json:"type"`
		Position string          `json:"position"`
		Data     json.RawMessage `json:"data"`
	}{
		Type:     c.Type(),
		Position: c.Position,
		Data:     data,
	})
}

func (c *Change) UnmarshalJSON(data []byte) error {
	var aux struct {
		Type     string          `json:"type"`
		Position string          `json:"position"`
		Data     json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	c.Position = aux.Position

	switch aux.Type {
	case "dml":
		c.Data = &DMLData{}
	case "ddl":
		c.Data = &DDLData{}
	default:
		return fmt.Errorf("unknown change type: %s", aux.Type)
	}

	return json.Unmarshal(aux.Data, c.Data)
}

type DMLData struct {
	Table        string               `json:"table"`
	ColumnNames  []string             `json:"columnnames"`
	ColumnValues []ColumnValueWrapper `json:"columnvalues"`
	Kind         string               `json:"kind"`
	OldKeys      *struct {
		KeyNames  []string             `json:"keynames"`
		KeyValues []ColumnValueWrapper `json:"keyvalues"`
	} `json:"oldkeys,omitempty"`
}

func (c DMLData) Type() string {
	return "dml"
}

type DDLData struct {
	ID       int       `json:"id"`
	Time     time.Time `json:"time"`
	Username string    `json:"username"`
	Database string    `json:"database"`
	DDL      string    `json:"ddl"`
}

func (c DDLData) Type() string {
	return "ddl"
}
