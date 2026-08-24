package tile

import (
	"fmt"
	"sort"

	"geoindex/internal/grid"
	"geoindex/internal/model"
)

// Renderer generates text based map tiles from the grid index and keeps a
// bounded cache. Data updates invalidate the affected tile keys through the
// write service.
type Renderer struct {
	grid  *grid.Grid
	cache *TileCache
	jobs  *JobTracker
	zooms []int
}

// NewRenderer builds a renderer for the given zoom levels.
func NewRenderer(g *grid.Grid, zooms []int) *Renderer {
	levels := make([]int, len(zooms))
	copy(levels, zooms)
	sort.Ints(levels)
	return &Renderer{
		grid:  g,
		cache: NewTileCache(256),
		jobs:  NewJobTracker(),
		zooms: levels,
	}
}

// Zooms returns a copy of the configured zoom levels.
func (r *Renderer) Zooms() []int {
	out := make([]int, len(r.zooms))
	copy(out, r.zooms)
	return out
}

// Render produces the tile payload for a slippy coordinate. Cached tiles are
// served directly; otherwise a render job moves through pending, rendering
// and published states before the payload is cached.
func (r *Renderer) Render(z, x, y int) ([]byte, error) {
	key := r.grid.TileKey(z, x, y)
	if data, ok := r.cache.Get(key); ok {
		return data, nil
	}
	job := r.jobs.Begin(key)
	if job == nil {
		return nil, fmt.Errorf("tile %s already rendering", key)
	}
	payload := r.renderTile(z, x, y)
	r.cache.Put(key, payload)
	r.jobs.Publish(key)
	return payload, nil
}

// renderTile computes the tile bounds and serializes every indexed point that
// falls inside them.
func (r *Renderer) renderTile(z, x, y int) []byte {
	bounds := tileBounds(z, x, y)
	result := r.grid.QueryRange(bounds)
	points := make([]model.Point, len(result.Points))
	copy(points, result.Points)
	sort.Slice(points, func(i, j int) bool { return points[i].ID < points[j].ID })
	centerX, centerY := bounds.Center()
	header := fmt.Sprintf("#region %s\n#spatial %s\n", grid.RegionKey(bounds), grid.SpatialKey(centerX, centerY))
	return append([]byte(header), r.payload(points)...)
}

// payload renders the point set into deterministic text lines.
func (r *Renderer) payload(points []model.Point) []byte {
	if len(points) == 0 {
		return []byte("empty\n")
	}
	output := make([]byte, 0, len(points)*32)
	for _, point := range points {
		output = append(output, fmt.Sprintf("%.4f,%.4f %s\n", point.X, point.Y, point.Summary())...)
	}
	return output
}

// InvalidateRegion clears every cached tile touched by the rectangle.
func (r *Renderer) InvalidateRegion(box model.Rect) int {
	keys := r.grid.TilesForRect(box, r.zooms)
	cleared := r.cache.Invalidate(keys)
	for _, key := range keys {
		r.jobs.Reset(key)
	}
	return cleared
}

// InvalidateAll clears the whole tile cache.
func (r *Renderer) InvalidateAll() int {
	cleared := r.cache.Clear()
	r.jobs.ResetAll()
	return cleared
}

// CacheSize returns the number of cached tiles.
func (r *Renderer) CacheSize() int {
	return r.cache.Size()
}

// HitRate returns the tile cache hit ratio.
func (r *Renderer) HitRate() float64 {
	return r.cache.HitRate()
}

// JobStats returns the render state distribution.
func (r *Renderer) JobStats() map[JobState]int {
	return r.jobs.Stats()
}

// tileBounds maps slippy coordinates back onto world coordinates assuming a
// linear lon/lat projection over the default world extent.
func tileBounds(z, x, y int) model.Rect {
	size := 1 << z
	minX := float64(x)/float64(size)*360 - 180
	maxX := float64(x+1)/float64(size)*360 - 180
	minY := float64(y)/float64(size)*180 - 90
	maxY := float64(y+1)/float64(size)*180 - 90
	return model.Rect{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY}
}
