package query

import (
	"testing"

	"geoindex/internal/grid"
	"geoindex/internal/model"
	"geoindex/internal/rtree"
)

func TestEmptyGridQueryNoNilPanic(t *testing.T) {
	g := grid.NewGrid(4, 8)
	tree := rtree.NewRTree()
	service := NewService(g, tree)
	points, err := service.Range(model.Rect{MinX: -10, MinY: -10, MaxX: 10, MaxY: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 0 {
		t.Fatalf("expected empty result for empty grid, got %+v", points)
	}
}
