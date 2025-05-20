package types

import (
	"encoding/json"
	"fmt"
	"time"
)

type Change struct {
	LSN  string
	Data interface {
		Type() string
	}
}

func (c Change) Type() string {
	return c.Data.Type()
}

func (c Change) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(c.Data)
	if err != nil {
		return nil, err
	}

	return json.Marshal(struct {
		Type string          `json:"type"`
		LSN  string          `json:"lsn"`
		Data json.RawMessage `json:"data"`
	}{
		Type: c.Type(),
		LSN:  c.LSN,
		Data: data,
	})
}

func (c *Change) UnmarshalJSON(data []byte) error {
	var aux struct {
		Type string          `json:"type"`
		LSN  string          `json:"lsn"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	c.LSN = aux.LSN

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
	Table        string   `json:"table"`
	ColumnNames  []string `json:"columnnames"`
	ColumnValues []any    `json:"columnvalues"`
	Kind         string   `json:"kind"`
	OldKeys      *struct {
		KeyNames  []string `json:"keynames"`
		KeyValues []any    `json:"keyvalues"`
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
