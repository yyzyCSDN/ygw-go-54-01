package write

import (
	"testing"

	"geoindex/internal/grid"
	"geoindex/internal/model"
	"geoindex/internal/rtree"
	"geoindex/internal/wal"
)

func TestWriteUpdatesGridIndex(t *testing.T) {
	g := grid.NewGrid(4, 8)
	tree := rtree.NewRTree()
	log := wal.NewLog()
	service := NewService(g, tree, log, Options{})
	if err := service.AddPoint(model.NewPoint("poi-1", 10, 20)); err != nil {
		t.Fatal(err)
	}
	result := g.QueryRange(model.Rect{MinX: 0, MinY: 0, MaxX: 30, MaxY: 30})
	if len(result.Points) != 1 || result.Points[0].ID != "poi-1" {
		t.Fatalf("written point is not visible in the grid index: %+v", result.Points)
	}
}
