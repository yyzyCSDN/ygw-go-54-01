package query

import (
	"testing"

	"geoindex/internal/grid"
	"geoindex/internal/model"
	"geoindex/internal/rtree"
)

func TestRangeIncludesBoundaryPoints(t *testing.T) {
	g := grid.NewGrid(4, 8)
	tree := rtree.NewRTree()
	service := NewService(g, tree)
	point := model.NewPoint("edge", 10, 20)
	if _, err := g.Put(point); err != nil {
		t.Fatal(err)
	}
	tree.Insert(point)
	points, err := service.Range(model.Rect{MinX: 0, MinY: 0, MaxX: 10, MaxY: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 1 || points[0].ID != "edge" {
		t.Fatalf("boundary point missed by range query: %+v", points)
	}
}
