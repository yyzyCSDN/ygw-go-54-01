package write

import (
	"testing"

	"geoindex/internal/grid"
	"geoindex/internal/model"
	"geoindex/internal/rtree"
	"geoindex/internal/wal"
)

func TestAddPointRequiresID(t *testing.T) {
	service := newTestService()
	if err := service.AddPoint(model.Point{}); err != ErrEmptyID {
		t.Fatalf("expected ErrEmptyID, got %v", err)
	}
}

func TestAddPointAppendsLog(t *testing.T) {
	service := newTestService()
	if err := service.AddPoint(model.NewPoint("a", 1, 2)); err != nil {
		t.Fatal(err)
	}
	if service.Log().Size() != 1 {
		t.Fatalf("expected 1 log entry, got %d", service.Log().Size())
	}
}

func TestPolygonStoreValidation(t *testing.T) {
	store := NewPolygonStore()
	bad := model.Polygon{ID: "district", Vertices: []model.Point{{ID: "v1"}}}
	if err := store.Put(bad); err == nil {
		t.Fatal("expected validation error for a degenerate ring")
	}
	good := model.Polygon{
		ID:       "district",
		Vertices: []model.Point{model.NewPoint("v1", 0, 0), model.NewPoint("v2", 10, 0), model.NewPoint("v3", 5, 10)},
	}
	if err := store.Put(good); err != nil {
		t.Fatal(err)
	}
	if !store.Belongs("district", 5, 5) {
		t.Fatal("expected point inside district to belong")
	}
	if store.Belongs("district", 50, 50) {
		t.Fatal("expected point outside district to not belong")
	}
}

func newTestService() *Service {
	g := grid.NewGrid(4, 8)
	tree := rtree.NewRTree()
	log := wal.NewLog()
	return NewService(g, tree, log, Options{})
}
