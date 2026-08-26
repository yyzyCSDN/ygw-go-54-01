package write

import (
	"testing"

	"geoindex/internal/grid"
	"geoindex/internal/model"
	"geoindex/internal/rtree"
	"geoindex/internal/wal"
)

func TestDeletePointClearsIndexEntry(t *testing.T) {
	g := grid.NewGrid(4, 8)
	tree := rtree.NewRTree()
	log := wal.NewLog()
	service := NewService(g, tree, log, Options{})
	if err := service.AddPoint(model.NewPoint("gone", 5, 5)); err != nil {
		t.Fatal(err)
	}
	if err := service.DeletePoint("gone"); err != nil {
		t.Fatal(err)
	}
	result := g.QueryRange(model.Rect{MinX: 0, MinY: 0, MaxX: 10, MaxY: 10})
	for _, point := range result.Points {
		if point.ID == "gone" {
			t.Fatalf("deleted point is still returned by the grid index: %+v", result.Points)
		}
	}
}
