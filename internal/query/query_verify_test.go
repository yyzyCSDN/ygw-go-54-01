package query

import (
	"testing"

	"geoindex/internal/grid"
	"geoindex/internal/model"
	"geoindex/internal/rtree"
)

func TestNearestSortedByDistance(t *testing.T) {
	g := grid.NewGrid(4, 64)
	tree := rtree.NewRTree()
	service := NewService(g, tree)
	points := []model.Point{
		model.NewPoint("b", 65, 65),
		model.NewPoint("c", 70, 60),
		model.NewPoint("a", 58, 58),
	}
	for _, point := range points {
		if _, err := g.Put(point); err != nil {
			t.Fatal(err)
		}
		tree.Insert(point)
	}
	got, err := service.Nearest(60, 60, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 results, got %d: %+v", len(got), got)
	}
	for i := 1; i < len(got); i++ {
		prev := DistanceTo(got[i-1].X, got[i-1].Y, 60, 60)
		cur := DistanceTo(got[i].X, got[i].Y, 60, 60)
		if cur < prev {
			t.Fatalf("nearest results are not sorted by distance: %+v", got)
		}
	}
}
