package rebuild

import (
	"testing"

	"geoindex/internal/grid"
	"geoindex/internal/model"
	"geoindex/internal/query"
	"geoindex/internal/rtree"
	"geoindex/internal/wal"
	"geoindex/internal/write"
)

// TestRebuildKeepsMidRebuildWrites reproduces the reported regression: a write
// that lands between Begin and Swap must remain visible after the rebuild
// finishes. Before the delta merge was implemented, the rebuilt grid only held
// the Begin snapshot, so range queries dropped the freshly written points.
func TestRebuildKeepsMidRebuildWrites(t *testing.T) {
	g := grid.NewGrid(4, 8)
	tree := rtree.NewRTree()
	log := wal.NewLog()
	rb := NewRebuilder(g, tree, log)
	q := query.NewService(g, tree)

	// Seed the index before the rebuild starts.
	seed := model.NewPoint("seed", 1, 1)
	if _, err := g.Put(seed); err != nil {
		t.Fatal(err)
	}
	tree.Insert(seed)
	if _, err := log.Append(model.NewAddOp(0, seed)); err != nil {
		t.Fatal(err)
	}

	// Snapshot the index, then commit a new point that the snapshot missed.
	if err := rb.Begin(); err != nil {
		t.Fatal(err)
	}
	mid := model.NewPoint("mid", 5, 5)
	adder := write.NewService(g, tree, log, write.Options{})
	if err := adder.AddPoint(mid); err != nil {
		t.Fatal(err)
	}

	// Finish the rebuild: build, fold in the delta, swap.
	if err := rb.Build(); err != nil {
		t.Fatal(err)
	}
	if err := rb.MergeDelta(); err != nil {
		t.Fatal(err)
	}
	if err := rb.Swap(); err != nil {
		t.Fatal(err)
	}

	// The mid-rebuild write must be answerable by a range query against the
	// rebuilt indexes without a process restart.
	rect := model.Rect{MinX: 0, MinY: 0, MaxX: 10, MaxY: 10}
	points, err := q.Range(rect)
	if err != nil {
		t.Fatal(err)
	}
	if !containsID(points, "mid") {
		t.Fatalf("mid-rebuild write is missing after rebuild: %+v", points)
	}
	if len(points) != 2 {
		t.Fatalf("expected seed and mid after rebuild, got %d points: %+v", len(points), points)
	}
	if got := rb.Stats().AppliedOps; got != 1 {
		t.Fatalf("expected 1 applied delta op, got %d", got)
	}
}

// TestRebuildAppliesMidRebuildDelete ensures a delete committed during the
// rebuild is honored: the rebuilt index must not resurrect the deleted point.
func TestRebuildAppliesMidRebuildDelete(t *testing.T) {
	g := grid.NewGrid(4, 8)
	tree := rtree.NewRTree()
	log := wal.NewLog()
	rb := NewRebuilder(g, tree, log)
	q := query.NewService(g, tree)

	keeper := model.NewPoint("keeper", 1, 1)
	victim := model.NewPoint("victim", 2, 2)
	for _, p := range []model.Point{keeper, victim} {
		if _, err := g.Put(p); err != nil {
			t.Fatal(err)
		}
		tree.Insert(p)
		if _, err := log.Append(model.NewAddOp(0, p)); err != nil {
			t.Fatal(err)
		}
	}

	if err := rb.Begin(); err != nil {
		t.Fatal(err)
	}
	deleter := write.NewService(g, tree, log, write.Options{})
	if err := deleter.DeletePoint("victim"); err != nil {
		t.Fatal(err)
	}

	if err := rb.Build(); err != nil {
		t.Fatal(err)
	}
	if err := rb.MergeDelta(); err != nil {
		t.Fatal(err)
	}
	if err := rb.Swap(); err != nil {
		t.Fatal(err)
	}

	rect := model.Rect{MinX: 0, MinY: 0, MaxX: 10, MaxY: 10}
	points, err := q.Range(rect)
	if err != nil {
		t.Fatal(err)
	}
	if containsID(points, "victim") {
		t.Fatalf("deleted point was resurrected by rebuild: %+v", points)
	}
	if !containsID(points, "keeper") {
		t.Fatalf("keeper point was dropped by rebuild: %+v", points)
	}
}

// TestRebuildPendingDeltaAfterMerge confirms the progress indicator drops to
// zero once the delta has been folded in, so the control endpoint reports an
// accurate rebuild state.
func TestRebuildPendingDeltaAfterMerge(t *testing.T) {
	g := grid.NewGrid(4, 8)
	tree := rtree.NewRTree()
	log := wal.NewLog()
	rb := NewRebuilder(g, tree, log)

	if _, err := g.Put(model.NewPoint("a", 1, 1)); err != nil {
		t.Fatal(err)
	}
	if err := rb.Begin(); err != nil {
		t.Fatal(err)
	}
	if _, err := log.Append(model.NewAddOp(0, model.NewPoint("b", 2, 2))); err != nil {
		t.Fatal(err)
	}
	if err := rb.Build(); err != nil {
		t.Fatal(err)
	}
	if got := rb.PendingDelta(); got != 1 {
		t.Fatalf("expected 1 pending delta before merge, got %d", got)
	}
	if err := rb.MergeDelta(); err != nil {
		t.Fatal(err)
	}
	if got := rb.PendingDelta(); got != 0 {
		t.Fatalf("expected 0 pending delta after merge, got %d", got)
	}
	_ = rb.Swap()
}

func containsID(points []model.Point, id string) bool {
	for _, p := range points {
		if p.ID == id {
			return true
		}
	}
	return false
}
