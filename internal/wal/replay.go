package wal

import "geoindex/internal/model"

// ApplyFunc applies a single replayed operation.
type ApplyFunc func(model.Op) error

// Replay replays every retained operation through the apply function in
// sequence order and returns the highest sequence number seen.
func (l *Log) Replay(apply ApplyFunc) (uint64, error) {
	l.mu.RLock()
	entries := make([]model.Op, len(l.entries))
	copy(entries, l.entries)
	l.mu.RUnlock()
	var last uint64
	for _, op := range entries {
		if err := apply(op); err != nil {
			return last, err
		}
		last = op.Seq
	}
	return last, nil
}

// SnapshotState describes the log position and operation count at one point
// in time.
type SnapshotState struct {
	NextSeq  uint64
	Entries  int
	Compacted uint64
}

// State returns the current accounting snapshot.
func (l *Log) State() SnapshotState {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return SnapshotState{
		NextSeq:   l.nextSeq,
		Entries:   len(l.entries),
		Compacted: l.compacted,
	}
}

// LastSeq returns the highest assigned sequence or zero when empty.
func (l *Log) LastSeq() uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if len(l.entries) == 0 {
		return 0
	}
	return l.entries[len(l.entries)-1].Seq
}
