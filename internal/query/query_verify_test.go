package query

import (
	"testing"

	"geoindex/internal/grid"
	"geoindex/internal/model"
	"geoindex/internal/rtree"
)

func TestQueryUsesSplitGridCells(t *testing.T) {
	g := grid.NewGrid(8, 2)
	tree := rtree.NewRTree()
	service := NewService(g, tree)
	points := []model.Point{
		model.NewPoint("s1", 10, 10),
		model.NewPoint("s2", 10.1, 10),
		model.NewPoint("s3", 10, 10.1),
		model.NewPoint("s4", 10.1, 10.1),
		model.NewPoint("s5", 10.2, 10.2),
	}
	for _, point := range points {
		if _, err := g.Put(point); err != nil {
			t.Fatal(err)
		}
		tree.Insert(point)
	}
	points, err := service.Range(model.Rect{MinX: 9, MinY: 9, MaxX: 11, MaxY: 11})
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 5 {
		t.Fatalf("query after grid split lost points: got %d, want 5", len(points))
	}
}
