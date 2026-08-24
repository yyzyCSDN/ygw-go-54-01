package rebuild

import (
	"time"

	"geoindex/internal/grid"
)

// Begin captures the current index contents and the WAL position. Writes that
// commit after this point are folded in by MergeDelta.
func (rb *Rebuilder) Begin() error {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	if rb.state != StateIdle {
		return ErrRebuildInProgress
	}
	started := time.Now()
	rb.snapshot = rb.grid.AllPoints()
	rb.baseSeq = rb.log.NextSeq()
	rb.state = StateBuilding
	rb.stats.SnapshotPoints = len(rb.snapshot)
	rb.stats.BeginMs = time.Since(started).Milliseconds()
	return nil
}

// Build reconstructs a fresh grid from the snapshot points.
func (rb *Rebuilder) Build() error {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	if rb.state != StateBuilding {
		return ErrRebuildNotStarted
	}
	started := time.Now()
	next := grid.NewGrid(rb.grid.MaxDepth(), rb.grid.MaxPoints())
	for _, point := range rb.snapshot {
		if _, err := next.Put(point.Clone()); err != nil {
			return err
		}
	}
	rb.built = next
	rb.stats.BuiltPoints = len(rb.snapshot)
	rb.stats.BuildMs = time.Since(started).Milliseconds()
	return nil
}

// Swap publishes the rebuilt index and reconstructs the range tree from the
// final point set.
func (rb *Rebuilder) Swap() error {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	if rb.built == nil {
		return ErrRebuildNotStarted
	}
	started := time.Now()
	rb.grid.Replace(rb.built)
	rb.tree.Rebuild(rb.grid.AllPoints())
	rb.state = StateDone
	rb.stats.FinalPoints = len(rb.grid.AllPoints())
	rb.stats.SwapMs = time.Since(started).Milliseconds()
	return nil
}
