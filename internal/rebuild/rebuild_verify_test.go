package rebuild

import (
	"testing"

	"geoindex/internal/grid"
	"geoindex/internal/model"
	"geoindex/internal/rtree"
	"geoindex/internal/wal"
	"geoindex/internal/write"
)

func TestRebuildIncludesConcurrentWrites(t *testing.T) {
	g := grid.NewGrid(4, 8)
	tree := rtree.NewRTree()
	log := wal.NewLog()
	writeService := write.NewService(g, tree, log, write.Options{})
	if err := writeService.AddPoint(model.NewPoint("p1", 1, 1)); err != nil {
		t.Fatal(err)
	}
	rb := NewRebuilder(g, tree, log)
	if err := rb.Begin(); err != nil {
		t.Fatal(err)
	}
	if err := rb.Build(); err != nil {
		t.Fatal(err)
	}
	if err := writeService.AddPoint(model.NewPoint("p2", 2, 2)); err != nil {
		t.Fatal(err)
	}
	if err := rb.MergeDelta(); err != nil {
		t.Fatal(err)
	}
	if err := rb.Swap(); err != nil {
		t.Fatal(err)
	}
	result := g.QueryRange(model.Rect{MinX: 0, MinY: 0, MaxX: 10, MaxY: 10})
	if len(result.Points) != 2 {
		t.Fatalf("rebuild dropped writes that arrived mid-rebuild: %+v", result.Points)
	}
}
