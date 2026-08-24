package wal

import (
	"fmt"
	"sync"

	"geoindex/internal/model"
)

// ErrEmptyOp is returned when an operation without a kind is appended.
var ErrEmptyOp = fmt.Errorf("wal op kind is empty")

// Log is an in-process write-ahead log. Every spatial change is appended
// before it is applied to the indexes, which gives replay and rebuild a
// consistent, ordered view of all mutations.
type Log struct {
	mu       sync.RWMutex
	nextSeq  uint64
	entries  []model.Op
	compacted uint64
}

// NewLog builds an empty log.
func NewLog() *Log {
	return &Log{nextSeq: 1}
}

// Append records the operation and returns its assigned sequence number.
func (l *Log) Append(op model.Op) (uint64, error) {
	if op.Kind == 0 {
		return 0, ErrEmptyOp
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	op.Seq = l.nextSeq
	l.nextSeq++
	l.entries = append(l.entries, op)
	return op.Seq, nil
}

// NextSeq returns the sequence number the next append will receive.
func (l *Log) NextSeq() uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.nextSeq
}

// Since returns operations with sequence numbers greater than or equal to
// the given boundary, in append order.
func (l *Log) Since(seq uint64) []model.Op {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]model.Op, 0, len(l.entries))
	for _, op := range l.entries {
		if op.Seq >= seq {
			out = append(out, op)
		}
	}
	return out
}

// Entries returns a copy of all retained operations.
func (l *Log) Entries() []model.Op {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]model.Op, len(l.entries))
	copy(out, l.entries)
	return out
}

// Truncate removes operations with sequence numbers below the boundary. The
// compacted counter keeps accounting visible in metrics.
func (l *Log) Truncate(below uint64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	kept := l.entries[:0]
	for _, op := range l.entries {
		if op.Seq < below {
			l.compacted++
			continue
		}
		kept = append(kept, op)
	}
	l.entries = kept
}

// Size returns the number of retained operations.
func (l *Log) Size() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.entries)
}

// Compacted returns how many operations were dropped by truncation.
func (l *Log) Compacted() uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.compacted
}
