package rebuild

import "time"

// MergeDelta folds every operation committed after Begin into the rebuilt
// index. Skipping this phase would drop writes that arrived mid-rebuild.
func (rb *Rebuilder) MergeDelta() error {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	if rb.built == nil {
		return ErrRebuildNotStarted
	}
	started := time.Now()
	ops := rb.log.Since(rb.baseSeq)
	for _, op := range ops {
		if err := rb.applyOp(rb.built, op); err != nil {
			return err
		}
		rb.stats.AppliedOps++
	}
	rb.stats.DeltaMs = time.Since(started).Milliseconds()
	return nil
}

// PendingDelta returns the operations that still need folding, used by the
// control endpoint to report rebuild progress.
func (rb *Rebuilder) PendingDelta() int {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	if rb.built == nil {
		return 0
	}
	return len(rb.log.Since(rb.baseSeq))
}

// Stats accumulates phase durations and point accounting for diagnostics.
type Stats struct {
	SnapshotPoints int
	BuiltPoints    int
	FinalPoints    int
	AppliedOps     int
	BeginMs        int64
	BuildMs        int64
	DeltaMs        int64
	SwapMs         int64
}
