package converter

import (
	"testing"
)

func TestLSNGenerator_Next(t *testing.T) {
	gen := NewLSNGenerator()

	// Test sequential LSN generation
	expected := []string{
		"0/BOOTSTRAP0000000000000001",
		"0/BOOTSTRAP0000000000000002",
		"0/BOOTSTRAP0000000000000003",
	}

	for i, expectedLSN := range expected {
		lsn := gen.Next()
		if lsn != expectedLSN {
			t.Errorf("LSN %d: expected %q, got %q", i+1, expectedLSN, lsn)
		}

		// Verify sequence number
		if gen.GetSequence() != int64(i+1) {
			t.Errorf("Sequence %d: expected %d, got %d", i+1, i+1, gen.GetSequence())
		}
	}
}

func TestLSNGenerator_Peek(t *testing.T) {
	gen := NewLSNGenerator()

	// Peek should return next LSN without incrementing
	peek1 := gen.Peek()
	peek2 := gen.Peek()

	if peek1 != peek2 {
		t.Errorf("Peek should be consistent: %q != %q", peek1, peek2)
	}

	if peek1 != "0/BOOTSTRAP0000000000000001" {
		t.Errorf("First peek should be 0/BOOTSTRAP0000000000000001, got %q", peek1)
	}

	// After calling Next(), peek should advance
	lsn := gen.Next()
	if lsn != peek1 {
		t.Errorf("Next() should return peeked value: %q != %q", lsn, peek1)
	}

	peek3 := gen.Peek()
	if peek3 != "0/BOOTSTRAP0000000000000002" {
		t.Errorf("Peek after Next() should be 0/BOOTSTRAP0000000000000002, got %q", peek3)
	}
}

func TestLSNGenerator_SetSequence(t *testing.T) {
	gen := NewLSNGenerator()

	// Set sequence to specific value
	err := gen.SetSequence(100)
	if err != nil {
		t.Errorf("SetSequence failed: %v", err)
	}

	if gen.GetSequence() != 100 {
		t.Errorf("Expected sequence 100, got %d", gen.GetSequence())
	}

	// Next LSN should be sequence + 1
	lsn := gen.Next()
	if lsn != "0/BOOTSTRAP0000000000000101" {
		t.Errorf("Expected 0/BOOTSTRAP0000000000000101, got %q", lsn)
	}

	// Test invalid sequence
	err2 := gen.SetSequence(-1)
	if err2 == nil {
		t.Error("SetSequence should fail for negative values")
	}
}

func TestLSNGenerator_Reset(t *testing.T) {
	gen := NewLSNGenerator()

	// Generate some LSNs
	gen.Next()
	gen.Next()

	if gen.GetSequence() != 2 {
		t.Errorf("Expected sequence 2, got %d", gen.GetSequence())
	}

	// Reset should go back to 0
	gen.Reset()

	if gen.GetSequence() != 0 {
		t.Errorf("After reset, expected sequence 0, got %d", gen.GetSequence())
	}

	if gen.GetCurrent() != "" {
		t.Errorf("After reset, expected empty current LSN, got %q", gen.GetCurrent())
	}

	// Next LSN should start from 1 again
	lsn := gen.Next()
	if lsn != "0/BOOTSTRAP0000000000000001" {
		t.Errorf("After reset, expected 0/BOOTSTRAP0000000000000001, got %q", lsn)
	}
}

func TestParseBootstrapLSN(t *testing.T) {
	tests := []struct {
		lsn         string
		expectedSeq int64
		expectError bool
	}{
		{"0/BOOTSTRAP0000000000000001", 1, false},
		{"0/BOOTSTRAP0000000000000123", 123, false},
		{"0/BOOTSTRAP9999999999999999", 9999999999999999, false},
		{"0/100", 0, true},              // Not a bootstrap LSN
		{"0/BOOTSTRAPinvalid", 0, true}, // Invalid format
		{"invalid", 0, true},            // Invalid format
	}

	for _, tt := range tests {
		seq, err := ParseBootstrapLSN(tt.lsn)

		if tt.expectError {
			if err == nil {
				t.Errorf("ParseBootstrapLSN(%q) should have failed", tt.lsn)
			}
			continue
		}

		if err != nil {
			t.Errorf("ParseBootstrapLSN(%q) failed: %v", tt.lsn, err)
			continue
		}

		if seq != tt.expectedSeq {
			t.Errorf("ParseBootstrapLSN(%q) = %d, expected %d", tt.lsn, seq, tt.expectedSeq)
		}
	}
}

func TestIsBootstrapLSN(t *testing.T) {
	tests := []struct {
		lsn      string
		expected bool
	}{
		{"0/BOOTSTRAP0000000000000001", true},
		{"0/BOOTSTRAP9999999999999999", true},
		{"0/100", false},
		{"0/1A000000", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		result := IsBootstrapLSN(tt.lsn)
		if result != tt.expected {
			t.Errorf("IsBootstrapLSN(%q) = %v, expected %v", tt.lsn, result, tt.expected)
		}
	}
}

func TestCompareLSNs(t *testing.T) {
	tests := []struct {
		lsn1     string
		lsn2     string
		expected int
	}{
		// Bootstrap LSN comparisons
		{"0/BOOTSTRAP0000000000000001", "0/BOOTSTRAP0000000000000002", -1},
		{"0/BOOTSTRAP0000000000000002", "0/BOOTSTRAP0000000000000001", 1},
		{"0/BOOTSTRAP0000000000000001", "0/BOOTSTRAP0000000000000001", 0},

		// Bootstrap vs PostgreSQL LSN
		{"0/BOOTSTRAP0000000000000001", "0/100", -1},
		{"0/100", "0/BOOTSTRAP0000000000000001", 1},

		// PostgreSQL LSN comparisons
		{"0/100", "0/200", -1},
		{"0/200", "0/100", 1},
		{"0/100", "0/100", 0},
	}

	for _, tt := range tests {
		result, err := CompareLSNs(tt.lsn1, tt.lsn2)
		if err != nil {
			t.Errorf("CompareLSNs(%q, %q) failed: %v", tt.lsn1, tt.lsn2, err)
			continue
		}

		if result != tt.expected {
			t.Errorf("CompareLSNs(%q, %q) = %d, expected %d", tt.lsn1, tt.lsn2, result, tt.expected)
		}
	}
}
