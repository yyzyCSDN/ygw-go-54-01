package rebuild

import (
	"testing"

	"geoindex/internal/grid"
	"geoindex/internal/model"
	"geoindex/internal/rtree"
	"geoindex/internal/wal"
)

func TestRebuildStateMachine(t *testing.T) {
	g := grid.NewGrid(4, 8)
	tree := rtree.NewRTree()
	log := wal.NewLog()
	rb := NewRebuilder(g, tree, log)
	if rb.State() != StateIdle {
		t.Fatalf("expected idle state, got %s", rb.State())
	}
	if err := rb.Begin(); err != nil {
		t.Fatal(err)
	}
	if rb.State() != StateBuilding {
		t.Fatalf("expected building state, got %s", rb.State())
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
	if rb.State() != StateDone {
		t.Fatalf("expected done state, got %s", rb.State())
	}
}

func TestRebuildRequiresBegin(t *testing.T) {
	g := grid.NewGrid(4, 8)
	tree := rtree.NewRTree()
	log := wal.NewLog()
	rb := NewRebuilder(g, tree, log)
	if err := rb.Build(); err == nil {
		t.Fatal("expected error when building before Begin")
	}
}

func TestRebuildPreservesPoints(t *testing.T) {
	g := grid.NewGrid(4, 8)
	tree := rtree.NewRTree()
	log := wal.NewLog()
	points := []model.Point{
		model.NewPoint("a", 1, 1),
		model.NewPoint("b", 2, 2),
		model.NewPoint("c", 3, 3),
	}
	for _, point := range points {
		_, _ = g.Put(point)
		tree.Insert(point)
	}
	rb := NewRebuilder(g, tree, log)
	if err := rb.Rebuild(); err != nil {
		t.Fatal(err)
	}
	if got := len(g.AllPoints()); got != 3 {
		t.Fatalf("expected 3 points after rebuild, got %d", got)
	}
}
