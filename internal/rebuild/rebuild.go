package rebuild

import (
	"errors"
	"sync"

	"geoindex/internal/grid"
	"geoindex/internal/model"
	"geoindex/internal/rtree"
	"geoindex/internal/wal"
)

// ErrRebuildInProgress is returned when Begin is called twice without Swap.
var ErrRebuildInProgress = errors.New("rebuild already in progress")

// ErrRebuildNotStarted is returned when a phase runs before Begin.
var ErrRebuildNotStarted = errors.New("rebuild has not started")

// State is the rebuild lifecycle state.
type State uint8

const (
	StateIdle State = iota
	StateBuilding
	StateDone
)

// String returns a stable state name.
func (s State) String() string {
	switch s {
	case StateBuilding:
		return "building"
	case StateDone:
		return "done"
	default:
		return "idle"
	}
}

// Rebuilder reconstructs the grid index from a point snapshot and then folds
// in every write that arrived after the snapshot was taken, so a rebuild
// never drops data that was committed while it was running.
type Rebuilder struct {
	mu       sync.Mutex
	grid     *grid.Grid
	tree     *rtree.RTree
	log      *wal.Log
	state    State
	baseSeq  uint64
	snapshot []model.Point
	built    *grid.Grid
	stats    Stats
}

// NewRebuilder wires the rebuilder onto the live indexes.
func NewRebuilder(g *grid.Grid, tree *rtree.RTree, log *wal.Log) *Rebuilder {
	return &Rebuilder{grid: g, tree: tree, log: log, state: StateIdle}
}

// Rebuild runs the whole pipeline: snapshot, reconstruction, delta merge and
// index swap.
func (rb *Rebuilder) Rebuild() error {
	if err := rb.Begin(); err != nil {
		return err
	}
	if err := rb.Build(); err != nil {
		return err
	}
	if err := rb.MergeDelta(); err != nil {
		return err
	}
	return rb.Swap()
}

// State returns the current lifecycle state.
func (rb *Rebuilder) State() State {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.state
}

// Stats returns the accumulated rebuild accounting.
func (rb *Rebuilder) Stats() Stats {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.stats
}

// SnapshotCount returns how many points were captured in Begin.
func (rb *Rebuilder) SnapshotCount() int {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return len(rb.snapshot)
}

// applyOp replays a delta operation onto a rebuilt index.
func (rb *Rebuilder) applyOp(index *grid.Grid, op model.Op) error {
	switch op.Kind {
	case model.OpAddPoint:
		_, err := index.Put(op.Point)
		return err
	case model.OpDeletePoint:
		index.Remove(op.Deleted)
		return nil
	case model.OpUpdatePolygon:
		return nil
	default:
		return nil
	}
}
