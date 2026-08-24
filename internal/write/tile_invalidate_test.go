package write

import (
	"bytes"
	"testing"

	"geoindex/internal/grid"
	"geoindex/internal/model"
	"geoindex/internal/rtree"
	"geoindex/internal/tile"
	"geoindex/internal/wal"
)

// newTileService wires a write service whose mutations drive the shared tile
// renderer, mirroring how cmd/geoindex assembles the live server.
func newTileService() (*Service, *tile.Renderer) {
	g := grid.NewGrid(4, 8)
	tree := rtree.NewRTree()
	log := wal.NewLog()
	renderer := tile.NewRenderer(g, []int{0})
	return NewService(g, tree, log, Options{Tiles: renderer}), renderer
}

// TestTileInvalidatedOnAddPoint verifies that inserting a point flushes the
// cached tile covering it, so the next render reflects the new data rather
// than serving a stale payload. This is the regression guard for the
// "updated spatial data still shows old tiles" report.
func TestTileInvalidatedOnAddPoint(t *testing.T) {
	service, renderer := newTileService()
	if err := service.AddPoint(model.NewPoint("a", 10, 20)); err != nil {
		t.Fatal(err)
	}

	first, err := renderer.Render(0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(first, []byte("a:")) {
		t.Fatalf("expected first tile to contain point a: %s", first)
	}
	if renderer.CacheSize() != 1 {
		t.Fatalf("expected 1 cached tile, got %d", renderer.CacheSize())
	}

	// Add a second point inside the same tile. The write path must invalidate
	// the cached entry so the subsequent render regenerates the payload.
	if err := service.AddPoint(model.NewPoint("b", 12, 18)); err != nil {
		t.Fatal(err)
	}

	second, err := renderer.Render(0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(second, []byte("b:")) {
		t.Fatalf("expected refreshed tile to contain point b: %s", second)
	}
	if bytes.Equal(first, second) {
		t.Fatalf("tile payload unchanged after data update: %s", second)
	}
}

// TestTileInvalidatedOnDeletePoint verifies that deleting a point flushes the
// covering tile so the next render no longer advertises the removed point.
func TestTileInvalidatedOnDeletePoint(t *testing.T) {
	service, renderer := newTileService()
	if err := service.AddPoint(model.NewPoint("a", 10, 20)); err != nil {
		t.Fatal(err)
	}
	if _, err := renderer.Render(0, 0, 0); err != nil {
		t.Fatal(err)
	}

	if err := service.DeletePoint("a"); err != nil {
		t.Fatal(err)
	}

	second, err := renderer.Render(0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(second, []byte("a:")) {
		t.Fatalf("expected stale point a to be gone after delete: %s", second)
	}
}
