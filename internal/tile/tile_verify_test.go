package tile

import (
	"bytes"
	"testing"

	"geoindex/internal/grid"
	"geoindex/internal/model"
)

func TestTileCacheInvalidatedOnDataChange(t *testing.T) {
	g := grid.NewGrid(4, 8)
	renderer := NewRenderer(g, []int{0, 1})
	if _, err := g.Put(model.NewPoint("first", 10, 20)); err != nil {
		t.Fatal(err)
	}
	first, err := renderer.Render(0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(first, []byte("first")) {
		t.Fatalf("first tile misses the point: %s", first)
	}
	if _, err := g.Put(model.NewPoint("second", 11, 21)); err != nil {
		t.Fatal(err)
	}
	renderer.InvalidateRegion(model.Rect{MinX: 10, MinY: 20, MaxX: 12, MaxY: 22})
	second, err := renderer.Render(0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(second, []byte("second")) {
		t.Fatalf("tile cache was not invalidated after a data change: %s", second)
	}
}
