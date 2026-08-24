package grid

import (
	"testing"

	"geoindex/internal/model"
)

// TestQueryRangeAfterSplit verifies that range queries observe the post-split
// topology. Before the fix CellsForQuery walked only the root, so once the
// root split and its points migrated into children, every point within the
// old (pre-split) cell disappeared from range queries.
func TestQueryRangeAfterSplit(t *testing.T) {
	const maxPoints = 4
	g := NewGrid(4, maxPoints)

	// Pack the root past capacity to force a split. Points straddle the
	// split boundary (midX=0, midY=0) so the bug surfaces at the border.
	points := []model.Point{
		model.NewPoint("ne", 1, 1),
		model.NewPoint("nw", -1, 1),
		model.NewPoint("sw", -1, -1),
		model.NewPoint("se", 1, -1),
		model.NewPoint("ctr", 0, 0), // lies exactly on the boundary
	}
	for _, p := range points {
		if _, err := g.Put(p); err != nil {
			t.Fatal(err)
		}
	}

	if g.SplitCount() == 0 {
		t.Fatalf("expected at least one split, got %d", g.SplitCount())
	}

	rect := model.Rect{MinX: -2, MinY: -2, MaxX: 2, MaxY: 2}
	result := g.QueryRange(rect)

	got := make(map[string]bool, len(result.Points))
	for _, p := range result.Points {
		got[p.ID] = true
	}
	for _, p := range points {
		if !got[p.ID] {
			t.Errorf("point %s missing from range query after split: %+v", p.ID, result.Points)
		}
	}
}

// TestCellsForQueryDescendsSplitChildren checks the public CellsForQuery path
// used by the query service returns the leaf cells, not the emptied parent.
func TestCellsForQueryDescendsSplitChildren(t *testing.T) {
	const maxPoints = 2
	g := NewGrid(2, maxPoints)

	for i := 0; i < maxPoints+1; i++ {
		x := float64(i)
		if _, err := g.Put(model.NewPoint(string(rune('a'+i)), x, x)); err != nil {
			t.Fatal(err)
		}
	}

	if g.SplitCount() == 0 {
		t.Fatalf("expected a split to have occurred")
	}

	cells := g.CellsForQuery(model.WorldExtent())
	for _, c := range cells {
		if c.State == CellSplit {
			t.Fatalf("CellsForQuery returned a split cell %q instead of its children", c.ID)
		}
	}
	if len(cells) < 2 {
		t.Fatalf("expected multiple leaf cells after split, got %d", len(cells))
	}
}
