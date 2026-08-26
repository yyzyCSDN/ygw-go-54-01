package write

import (
	"testing"

	"geoindex/internal/grid"
	"geoindex/internal/model"
	"geoindex/internal/rtree"
	"geoindex/internal/wal"
)

func TestPolygonUpdateErrorNotSwallowed(t *testing.T) {
	g := grid.NewGrid(4, 8)
	tree := rtree.NewRTree()
	log := wal.NewLog()
	service := NewService(g, tree, log, Options{})
	bad := model.Polygon{
		ID:       "district",
		Vertices: []model.Point{model.NewPoint("v1", 0, 0), model.NewPoint("v2", 10, 0)},
	}
	if err := service.UpdatePolygon(bad); err == nil {
		t.Fatal("polygon update failure must be propagated to the caller")
	}
}
