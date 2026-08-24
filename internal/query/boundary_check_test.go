package query

import (
	"testing"

	"geoindex/internal/grid"
	"geoindex/internal/model"
	"geoindex/internal/rtree"
)

func TestRangeBoundaryPointsIncludedTemp(t *testing.T) {
	g := grid.NewGrid(4, 8)
	tree := rtree.NewRTree()
	service := NewService(g, tree)
	edge := model.NewPoint("edge", 20, 20)
	_, _ = g.Put(edge)
	tree.Insert(edge)
	pts, err := service.Range(model.Rect{MinX: 0, MinY: 0, MaxX: 20, MaxY: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(pts) != 1 || pts[0].ID != "edge" {
		t.Fatalf("expected boundary point 'edge' to be included, got %+v", pts)
	}
}
