package rtree

import (
	"testing"
	"time"

	"geoindex/internal/model"
)

func TestRTreeLockReleasedAfterQuery(t *testing.T) {
	tree := NewRTree()
	tree.Insert(model.NewPoint("a", 1, 1))
	tree.Insert(model.NewPoint("b", 2, 2))
	box := model.Rect{MinX: 0, MinY: 0, MaxX: 10, MaxY: 10}
	results := tree.RangeSearch(box)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	done := make(chan struct{})
	go func() {
		tree.Insert(model.NewPoint("c", 3, 3))
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("range tree write lock is still held after the query returned")
	}
}
