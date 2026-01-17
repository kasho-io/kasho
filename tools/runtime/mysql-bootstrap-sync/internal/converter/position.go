package converter

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// PositionGenerator generates synthetic positions for bootstrap data.
//
// Bootstrap Position Ordering Guarantee:
// Bootstrap positions are guaranteed to be processed before real binlog data through the
// scoring mechanism in pkg/kvbuffer, NOT through position string comparison.
//
// How it works:
//  1. Bootstrap positions (e.g., "0/BOOTSTRAP0000000000000001") are converted to negative scores:
//     score = -1000000 + sequence (e.g., -999999, -999998, ...)
//  2. MySQL binlog positions (e.g., "mysql-bin.000001:4") are converted to positive scores:
//     score = (filenum * 4294967296) + offset
//  3. Redis sorted sets order by score, so negative < positive always
//
// This means bootstrap positions are always processed first, regardless of their
// string representation.
type PositionGenerator struct {
	mu              sync.Mutex
	sequence        int64
	currentPosition string
}

// NewPositionGenerator creates a new position generator
func NewPositionGenerator() *PositionGenerator {
	return &PositionGenerator{
		sequence:        0,
		currentPosition: "",
	}
}

// Next generates the next synthetic position
func (g *PositionGenerator) Next() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.sequence++
	pos := fmt.Sprintf("0/BOOTSTRAP%016d", g.sequence)
	g.currentPosition = pos
	return pos
}

// Peek returns what the next position would be without incrementing
func (g *PositionGenerator) Peek() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	nextSeq := g.sequence + 1
	return fmt.Sprintf("0/BOOTSTRAP%016d", nextSeq)
}

// GetSequence returns the current sequence number
func (g *PositionGenerator) GetSequence() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.sequence
}

// GetCurrent returns the current position
func (g *PositionGenerator) GetCurrent() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.currentPosition
}

// ValidatePositionRange validates that we haven't generated too many positions
func (g *PositionGenerator) ValidatePositionRange() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.sequence == 0 {
		return nil // No positions generated yet
	}

	maxBootstrapSequence := int64(9999999999999999) // Allow up to 9,999,999,999,999,999 bootstrap changes
	if g.sequence > maxBootstrapSequence {
		return fmt.Errorf("generated too many bootstrap positions: %d exceeds maximum %d", g.sequence, maxBootstrapSequence)
	}

	return nil
}

// Reset resets the position generator to start from sequence 0
func (g *PositionGenerator) Reset() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.sequence = 0
	g.currentPosition = ""
}

// SetSequence sets the sequence to a specific value (useful for resuming)
func (g *PositionGenerator) SetSequence(seq int64) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if seq < 0 {
		return fmt.Errorf("sequence cannot be negative: %d", seq)
	}

	g.sequence = seq
	if seq > 0 {
		g.currentPosition = fmt.Sprintf("0/BOOTSTRAP%016d", seq)
	} else {
		g.currentPosition = ""
	}

	return nil
}

// ParseBootstrapPosition parses a bootstrap position and returns its sequence number
func ParseBootstrapPosition(pos string) (int64, error) {
	if !strings.HasPrefix(pos, "0/BOOTSTRAP") {
		return 0, fmt.Errorf("not a bootstrap position: %s", pos)
	}

	seqStr := pos[11:] // Remove "0/BOOTSTRAP" prefix
	seq, err := strconv.ParseInt(seqStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid bootstrap position sequence: %s", seqStr)
	}

	return seq, nil
}

// IsBootstrapPosition checks if a position is a bootstrap position
func IsBootstrapPosition(pos string) bool {
	return strings.HasPrefix(pos, "0/BOOTSTRAP")
}

// IsMySQLBinlogPosition checks if a position is a MySQL binlog position
func IsMySQLBinlogPosition(pos string) bool {
	// MySQL binlog positions look like: mysql-bin.000001:4 or binlog.000001:4
	return strings.Contains(pos, ":") && (strings.Contains(pos, "bin.") || strings.HasPrefix(pos, "binlog."))
}

// ComparePositions compares two positions, handling both MySQL binlog and bootstrap positions
// Returns -1 if pos1 < pos2, 0 if equal, 1 if pos1 > pos2
func ComparePositions(pos1, pos2 string) (int, error) {
	// Both are bootstrap positions
	if IsBootstrapPosition(pos1) && IsBootstrapPosition(pos2) {
		seq1, err := ParseBootstrapPosition(pos1)
		if err != nil {
			return 0, err
		}
		seq2, err := ParseBootstrapPosition(pos2)
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

	// Bootstrap positions are always less than MySQL binlog positions
	if IsBootstrapPosition(pos1) && !IsBootstrapPosition(pos2) {
		return -1, nil
	}
	if !IsBootstrapPosition(pos1) && IsBootstrapPosition(pos2) {
		return 1, nil
	}

	// Both are MySQL binlog positions
	score1, err := parseMySQLBinlogScore(pos1)
	if err != nil {
		return 0, fmt.Errorf("invalid MySQL position %s: %w", pos1, err)
	}
	score2, err := parseMySQLBinlogScore(pos2)
	if err != nil {
		return 0, fmt.Errorf("invalid MySQL position %s: %w", pos2, err)
	}

	if score1 < score2 {
		return -1, nil
	} else if score1 > score2 {
		return 1, nil
	}
	return 0, nil
}

// parseMySQLBinlogScore parses a MySQL binlog position and returns a score
func parseMySQLBinlogScore(position string) (float64, error) {
	parts := strings.Split(position, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid binlog position format: %s", position)
	}

	// Extract file number from "mysql-bin.000001" or "binlog.000001"
	filename := parts[0]
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		fileNumStr := filename[idx+1:]
		fileNum, err := strconv.ParseInt(fileNumStr, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid binlog file number: %s", position)
		}
		offset, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid binlog offset: %s", position)
		}
		// Combine: file number * 4GB + offset for monotonic ordering
		return float64(fileNum)*4294967296 + float64(offset), nil
	}

	return 0, fmt.Errorf("invalid binlog format: %s", position)
}
