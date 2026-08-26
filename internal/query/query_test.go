package query

import (
	"testing"

	"geoindex/internal/grid"
	"geoindex/internal/model"
	"geoindex/internal/rtree"
)

func TestQueryRangeBasics(t *testing.T) {
	g := grid.NewGrid(4, 8)
	tree := rtree.NewRTree()
	service := NewService(g, tree)
	_, _ = g.Put(model.NewPoint("a", 10, 10))
	tree.Insert(model.NewPoint("a", 10, 10))
	points, err := service.Range(model.Rect{MinX: 0, MinY: 0, MaxX: 20, MaxY: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 1 || points[0].ID != "a" {
		t.Fatalf("unexpected range results: %+v", points)
	}
}

func TestQueryCount(t *testing.T) {
	g := grid.NewGrid(4, 8)
	tree := rtree.NewRTree()
	service := NewService(g, tree)
	for i := 0; i < 5; i++ {
		point := model.NewPoint(string(rune('a'+i)), float64(i), float64(i))
		_, _ = g.Put(point)
		tree.Insert(point)
	}
	count, err := service.Count(model.Rect{MinX: 0, MinY: 0, MaxX: 10, MaxY: 10})
	if err != nil {
		t.Fatal(err)
	}
	if count != 5 {
		t.Fatalf("expected 5 points, got %d", count)
	}
}

func TestDistanceTo(t *testing.T) {
	got := DistanceTo(0, 0, 3, 4)
	if got < 4.999 || got > 5.001 {
		t.Fatalf("expected distance ~5, got %f", got)
	}
}
