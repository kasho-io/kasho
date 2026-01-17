package converter

import (
	"testing"
)

func TestPositionGenerator_Next(t *testing.T) {
	gen := NewPositionGenerator()

	pos1 := gen.Next()
	if pos1 != "0/BOOTSTRAP0000000000000001" {
		t.Errorf("first position = %v, want 0/BOOTSTRAP0000000000000001", pos1)
	}

	pos2 := gen.Next()
	if pos2 != "0/BOOTSTRAP0000000000000002" {
		t.Errorf("second position = %v, want 0/BOOTSTRAP0000000000000002", pos2)
	}

	if gen.GetSequence() != 2 {
		t.Errorf("sequence = %d, want 2", gen.GetSequence())
	}
}

func TestPositionGenerator_Peek(t *testing.T) {
	gen := NewPositionGenerator()

	peek1 := gen.Peek()
	if peek1 != "0/BOOTSTRAP0000000000000001" {
		t.Errorf("peek = %v, want 0/BOOTSTRAP0000000000000001", peek1)
	}

	// Peek should not advance the sequence
	if gen.GetSequence() != 0 {
		t.Errorf("sequence after peek = %d, want 0", gen.GetSequence())
	}

	// Next should return what peek returned
	next := gen.Next()
	if next != peek1 {
		t.Errorf("next = %v, want %v (same as peek)", next, peek1)
	}
}

func TestParseBootstrapPosition(t *testing.T) {
	tests := []struct {
		position string
		wantSeq  int64
		wantErr  bool
	}{
		{"0/BOOTSTRAP0000000000000001", 1, false},
		{"0/BOOTSTRAP0000000000000100", 100, false},
		{"0/BOOTSTRAP9999999999999999", 9999999999999999, false},
		{"mysql-bin.000001:4", 0, true},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.position, func(t *testing.T) {
			seq, err := ParseBootstrapPosition(tt.position)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBootstrapPosition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if seq != tt.wantSeq {
				t.Errorf("ParseBootstrapPosition() = %v, want %v", seq, tt.wantSeq)
			}
		})
	}
}

func TestIsBootstrapPosition(t *testing.T) {
	tests := []struct {
		position string
		want     bool
	}{
		{"0/BOOTSTRAP0000000000000001", true},
		{"0/BOOTSTRAP9999999999999999", true},
		{"mysql-bin.000001:4", false},
		{"binlog.000001:123", false},
		{"0/16000000", false},
	}

	for _, tt := range tests {
		t.Run(tt.position, func(t *testing.T) {
			got := IsBootstrapPosition(tt.position)
			if got != tt.want {
				t.Errorf("IsBootstrapPosition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsMySQLBinlogPosition(t *testing.T) {
	tests := []struct {
		position string
		want     bool
	}{
		{"mysql-bin.000001:4", true},
		{"binlog.000001:123", true},
		{"mysql-bin.000123:456789", true},
		{"0/BOOTSTRAP0000000000000001", false},
		{"0/16000000", false},
	}

	for _, tt := range tests {
		t.Run(tt.position, func(t *testing.T) {
			got := IsMySQLBinlogPosition(tt.position)
			if got != tt.want {
				t.Errorf("IsMySQLBinlogPosition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComparePositions(t *testing.T) {
	tests := []struct {
		name string
		pos1 string
		pos2 string
		want int
	}{
		{
			name: "bootstrap less than bootstrap",
			pos1: "0/BOOTSTRAP0000000000000001",
			pos2: "0/BOOTSTRAP0000000000000002",
			want: -1,
		},
		{
			name: "bootstrap equal",
			pos1: "0/BOOTSTRAP0000000000000001",
			pos2: "0/BOOTSTRAP0000000000000001",
			want: 0,
		},
		{
			name: "bootstrap greater than bootstrap",
			pos1: "0/BOOTSTRAP0000000000000002",
			pos2: "0/BOOTSTRAP0000000000000001",
			want: 1,
		},
		{
			name: "bootstrap always less than binlog",
			pos1: "0/BOOTSTRAP9999999999999999",
			pos2: "mysql-bin.000001:4",
			want: -1,
		},
		{
			name: "binlog always greater than bootstrap",
			pos1: "mysql-bin.000001:4",
			pos2: "0/BOOTSTRAP0000000000000001",
			want: 1,
		},
		{
			name: "binlog comparison same file",
			pos1: "mysql-bin.000001:100",
			pos2: "mysql-bin.000001:200",
			want: -1,
		},
		{
			name: "binlog comparison different files",
			pos1: "mysql-bin.000001:999999",
			pos2: "mysql-bin.000002:1",
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ComparePositions(tt.pos1, tt.pos2)
			if err != nil {
				t.Fatalf("ComparePositions() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("ComparePositions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPositionGenerator_Reset(t *testing.T) {
	gen := NewPositionGenerator()

	gen.Next()
	gen.Next()
	gen.Next()

	if gen.GetSequence() != 3 {
		t.Errorf("sequence before reset = %d, want 3", gen.GetSequence())
	}

	gen.Reset()

	if gen.GetSequence() != 0 {
		t.Errorf("sequence after reset = %d, want 0", gen.GetSequence())
	}

	pos := gen.Next()
	if pos != "0/BOOTSTRAP0000000000000001" {
		t.Errorf("position after reset = %v, want 0/BOOTSTRAP0000000000000001", pos)
	}
}

func TestPositionGenerator_SetSequence(t *testing.T) {
	gen := NewPositionGenerator()

	err := gen.SetSequence(100)
	if err != nil {
		t.Fatalf("SetSequence() error = %v", err)
	}

	if gen.GetSequence() != 100 {
		t.Errorf("sequence = %d, want 100", gen.GetSequence())
	}

	pos := gen.Next()
	if pos != "0/BOOTSTRAP0000000000000101" {
		t.Errorf("position after SetSequence = %v, want 0/BOOTSTRAP0000000000000101", pos)
	}

	// Test negative sequence error
	err = gen.SetSequence(-1)
	if err == nil {
		t.Error("SetSequence(-1) should return error")
	}
}
