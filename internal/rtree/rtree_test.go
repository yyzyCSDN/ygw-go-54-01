package rtree

import (
	"testing"

	"geoindex/internal/model"
)

func TestRTreeInsertAndRangeSearch(t *testing.T) {
	tree := NewRTree()
	tree.Insert(model.NewPoint("a", 10, 10))
	tree.Insert(model.NewPoint("b", 20, 20))
	tree.Insert(model.NewPoint("c", 30, 30))
	box := model.Rect{MinX: 0, MinY: 0, MaxX: 25, MaxY: 25}
	got := tree.RangeSearch(box)
	if len(got) != 2 {
		t.Fatalf("expected 2 results, got %d", len(got))
	}
	ids := map[string]bool{}
	for _, point := range got {
		ids[point.ID] = true
	}
	if !ids["a"] || !ids["b"] {
		t.Fatalf("unexpected results: %+v", got)
	}
}

func TestRTreeRemove(t *testing.T) {
	tree := NewRTree()
	tree.Insert(model.NewPoint("a", 1, 1))
	tree.Insert(model.NewPoint("b", 2, 2))
	if !tree.Remove("a") {
		t.Fatal("expected removal to succeed")
	}
	box := model.Rect{MinX: 0, MinY: 0, MaxX: 10, MaxY: 10}
	if len(tree.RangeSearch(box)) != 1 {
		t.Fatal("removed point is still searchable")
	}
	if tree.Size() != 1 {
		t.Fatalf("expected size 1, got %d", tree.Size())
	}
}

func TestRTreeRebuild(t *testing.T) {
	tree := NewRTree()
	points := []model.Point{
		model.NewPoint("a", 1, 1),
		model.NewPoint("b", 2, 2),
		model.NewPoint("c", 3, 3),
	}
	tree.Rebuild(points)
	if tree.Size() != 3 {
		t.Fatalf("expected size 3, got %d", tree.Size())
	}
	box := model.Rect{MinX: 0, MinY: 0, MaxX: 10, MaxY: 10}
	if len(tree.RangeSearch(box)) != 3 {
		t.Fatal("rebuild did not make all points searchable")
	}
}
