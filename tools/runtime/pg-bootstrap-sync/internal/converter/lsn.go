package converter

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/jackc/pglogrepl"
)

// LSNGenerator generates synthetic LSNs for bootstrap data.
//
// Bootstrap LSN Ordering Guarantee:
// Bootstrap LSNs are guaranteed to be processed before real WAL data through the
// scoring mechanism in pkg/kvbuffer, NOT through LSN string comparison.
//
// How it works:
// 1. Bootstrap LSNs (e.g., "0/BOOTSTRAP0000000000000001") are converted to negative scores:
//    score = -1000000 + sequence (e.g., -999999, -999998, ...)
// 2. PostgreSQL LSNs (e.g., "0/16000000") are converted to positive scores:
//    score = pglogrepl.ParseLSN(lsn) (e.g., 16000000.0)
// 3. Redis sorted sets order by score, so negative < positive always
//
// This means bootstrap LSNs are always processed first, regardless of their
// string representation compared to the snapshot LSN. For example:
// - "0/BOOTSTRAP9999999999999999" → score -1 (processed first)
// - "0/16000000" → score 16000000.0 (processed after bootstrap)
type LSNGenerator struct {
	mu           sync.Mutex
	sequence     int64
	currentLSN   string
}

// NewLSNGenerator creates a new LSN generator
func NewLSNGenerator() *LSNGenerator {
	return &LSNGenerator{
		sequence:   0,
		currentLSN: "",
	}
}

// Next generates the next synthetic LSN
func (g *LSNGenerator) Next() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.sequence++
	lsn := fmt.Sprintf("0/BOOTSTRAP%016d", g.sequence)
	g.currentLSN = lsn
	return lsn
}

// Peek returns what the next LSN would be without incrementing
func (g *LSNGenerator) Peek() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	nextSeq := g.sequence + 1
	return fmt.Sprintf("0/BOOTSTRAP%016d", nextSeq)
}

// GetSequence returns the current sequence number
func (g *LSNGenerator) GetSequence() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.sequence
}

// GetCurrent returns the current LSN
func (g *LSNGenerator) GetCurrent() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.currentLSN
}

// ValidateLSNRange validates that all generated LSNs are before the snapshot LSN
func (g *LSNGenerator) ValidateLSNRange() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.sequence == 0 {
		return nil // No LSNs generated yet
	}

	// Bootstrap LSNs use negative scores in the kvbuffer, so they will always
	// sort before real PostgreSQL LSNs. However, we should still validate
	// that we haven't generated too many bootstrap LSNs.

	maxBootstrapSequence := int64(9999999999999999) // Allow up to 9,999,999,999,999,999 bootstrap changes
	if g.sequence > maxBootstrapSequence {
		return fmt.Errorf("generated too many bootstrap LSNs: %d exceeds maximum %d", g.sequence, maxBootstrapSequence)
	}

	return nil
}

// Reset resets the LSN generator to start from sequence 0
func (g *LSNGenerator) Reset() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.sequence = 0
	g.currentLSN = ""
}

// SetSequence sets the sequence to a specific value (useful for resuming)
func (g *LSNGenerator) SetSequence(seq int64) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if seq < 0 {
		return fmt.Errorf("sequence cannot be negative: %d", seq)
	}

	g.sequence = seq
	if seq > 0 {
		g.currentLSN = fmt.Sprintf("0/BOOTSTRAP%016d", seq)
	} else {
		g.currentLSN = ""
	}

	return nil
}

// ParseBootstrapLSN parses a bootstrap LSN and returns its sequence number
func ParseBootstrapLSN(lsn string) (int64, error) {
	if !strings.HasPrefix(lsn, "0/BOOTSTRAP") {
		return 0, fmt.Errorf("not a bootstrap LSN: %s", lsn)
	}

	seqStr := lsn[11:] // Remove "0/BOOTSTRAP" prefix
	seq, err := strconv.ParseInt(seqStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid bootstrap LSN sequence: %s", seqStr)
	}

	return seq, nil
}

// IsBootstrapLSN checks if an LSN is a bootstrap LSN
func IsBootstrapLSN(lsn string) bool {
	return strings.HasPrefix(lsn, "0/BOOTSTRAP")
}

// CompareLSNs compares two LSNs, handling both PostgreSQL and bootstrap LSNs
// Returns -1 if lsn1 < lsn2, 0 if equal, 1 if lsn1 > lsn2
func CompareLSNs(lsn1, lsn2 string) (int, error) {
	// Both are bootstrap LSNs
	if IsBootstrapLSN(lsn1) && IsBootstrapLSN(lsn2) {
		seq1, err := ParseBootstrapLSN(lsn1)
		if err != nil {
			return 0, err
		}
		seq2, err := ParseBootstrapLSN(lsn2)
		if err != nil {
			return 0, err
		}

		if seq1 < seq2 {
			return -1, nil
		} else if seq1 > seq2 {
			return 1, nil
		}
		return 0, nil
	}

	// Bootstrap LSNs are always less than PostgreSQL LSNs
	if IsBootstrapLSN(lsn1) && !IsBootstrapLSN(lsn2) {
		return -1, nil
	}
	if !IsBootstrapLSN(lsn1) && IsBootstrapLSN(lsn2) {
		return 1, nil
	}

	// Both are PostgreSQL LSNs
	pg1, err := pglogrepl.ParseLSN(lsn1)
	if err != nil {
		return 0, fmt.Errorf("invalid PostgreSQL LSN %s: %w", lsn1, err)
	}
	pg2, err := pglogrepl.ParseLSN(lsn2)
	if err != nil {
		return 0, fmt.Errorf("invalid PostgreSQL LSN %s: %w", lsn2, err)
	}

	if pg1 < pg2 {
		return -1, nil
	} else if pg1 > pg2 {
		return 1, nil
	}
	return 0, nil
}