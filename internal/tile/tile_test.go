package tile

import (
	"bytes"
	"testing"

	"geoindex/internal/grid"
	"geoindex/internal/model"
)

func TestTileRenderAndCache(t *testing.T) {
	g := grid.NewGrid(4, 8)
	_, _ = g.Put(model.NewPoint("a", 10, 20))
	renderer := NewRenderer(g, []int{0, 1})
	first, err := renderer.Render(0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(first, []byte("a")) {
		t.Fatalf("tile payload misses point a: %s", first)
	}
	if renderer.CacheSize() != 1 {
		t.Fatalf("expected 1 cached tile, got %d", renderer.CacheSize())
	}
	if renderer.JobStats()[JobPublished] != 1 {
		t.Fatalf("expected one published job, got %v", renderer.JobStats())
	}
	second, err := renderer.Render(0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("cached tile content changed")
	}
}

func TestTileInvalidateAll(t *testing.T) {
	g := grid.NewGrid(4, 8)
	_, _ = g.Put(model.NewPoint("a", 10, 20))
	renderer := NewRenderer(g, []int{0, 1})
	if _, err := renderer.Render(0, 0, 0); err != nil {
		t.Fatal(err)
	}
	cleared := renderer.InvalidateAll()
	if cleared != 1 {
		t.Fatalf("expected 1 cleared tile, got %d", cleared)
	}
	if renderer.CacheSize() != 0 {
		t.Fatalf("expected empty cache after invalidation, got %d", renderer.CacheSize())
	}
}
