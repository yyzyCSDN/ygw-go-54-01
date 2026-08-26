package grid

import (
	"fmt"
	"sort"
	"sync"

	"geoindex/internal/model"
)

// DefaultExtent is the world bounds used when no custom extent is supplied.
var DefaultExtent = model.GridBounds{MinX: -180, MinY: -90, MaxX: 180, MaxY: 90}

// Grid is the hierarchical spatial index. Points are stored in cells that
// split once they exceed the per-cell capacity. Every data mutation bumps the
// version counter, which downstream caches use to detect staleness.
type Grid struct {
	mu        sync.RWMutex
	root      *Cell
	cells     map[string]*Cell
	version   uint64
	splitCount uint64
	maxDepth  int
	maxPoints int
}

// NewGrid builds an empty grid with the given split policy.
func NewGrid(maxDepth, maxPointsPerCell int) *Grid {
	root := newCell("root", DefaultExtent, 0)
	grid := &Grid{
		root:      root,
		cells:     map[string]*Cell{root.ID: root},
		maxDepth:  maxDepth,
		maxPoints: maxPointsPerCell,
	}
	return grid
}

// MaxDepth returns the configured depth limit.
func (g *Grid) MaxDepth() int {
	return g.maxDepth
}

// MaxPoints returns the configured per-cell capacity.
func (g *Grid) MaxPoints() int {
	return g.maxPoints
}

// Put inserts a point into the grid, splitting cells when needed, and returns
// the id of the owning cell. A nil grid is not allowed.
func (g *Grid) Put(point model.Point) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if point.ID == "" {
		return "", fmt.Errorf("point id is empty")
	}
	cell := g.locateLocked(point.X, point.Y)
	cell.Add(point)
	g.version++
	g.splitIfNeededLocked(cell)
	return cell.ID, nil
}

// PutIndexed inserts a point and also marks the containing region as dirty.
// It is the write-path entry used by the write service.
func (g *Grid) PutIndexed(point model.Point) (string, error) {
	return g.Put(point)
}

// Remove deletes a point from whichever cell currently owns it, recursing
// through the cell registry so split cells are covered.
func (g *Grid) Remove(id string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, cell := range g.cells {
		if cell.Remove(id) {
			g.version++
			return true
		}
	}
	return false
}

// Get returns a point by id from the owning cell.
func (g *Grid) Get(id string) (model.Point, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, cell := range g.cells {
		if point, ok := cell.Points[id]; ok {
			return point, true
		}
	}
	return model.Point{}, false
}

// RemoveIndexed deletes a point and returns the owning cell id. The returned
// boolean reports whether the point existed.
func (g *Grid) RemoveIndexed(id string) (string, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, cell := range g.cells {
		if cell.Remove(id) {
			g.version++
			return cell.ID, true
		}
	}
	return "", false
}

// CellBounds returns the world bounds of a named cell.
func (g *Grid) CellBounds(id string) (model.Rect, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	cell, ok := g.cells[id]
	if !ok {
		return model.Rect{}, false
	}
	return cell.Bounds.Rect(), true
}

// NearestCandidates collects candidate points inside the rectangle, orders
// them by squared distance to the anchor and truncates to the top k results.
func (g *Grid) NearestCandidates(x, y float64, rect model.Rect, k int) []model.Point {
	g.mu.RLock()
	defer g.mu.RUnlock()
	cells := g.cellsForQueryLocked(rect)
	candidates := make([]model.Point, 0, 8)
	for _, cell := range cells {
		candidates = append(candidates, cell.Scan(rect)...)
	}
	sort.Slice(candidates, func(i, j int) bool {
		di := distanceSquared(candidates[i], x, y)
		dj := distanceSquared(candidates[j], x, y)
		if di != dj {
			return di < dj
		}
		return candidates[i].ID < candidates[j].ID
	})
	if len(candidates) > k {
		candidates = candidates[:k]
	}
	return candidates
}

func distanceSquared(point model.Point, x, y float64) float64 {
	return point.DistanceSquared(model.Point{X: x, Y: y})
}

// Version returns the current mutation counter.
func (g *Grid) Version() uint64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.version
}

// Bounds returns the world extent of the grid root.
func (g *Grid) Bounds() model.GridBounds {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.root.Bounds
}

// Replace swaps the entire index for a freshly built grid. The caller must
// have full ownership of the replacement.
func (g *Grid) Replace(other *Grid) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.root = other.root
	g.cells = other.cells
	g.version = other.version
	g.splitCount = other.splitCount
}

// AllPoints returns every indexed point in a deterministic depth-first order.
func (g *Grid) AllPoints() []model.Point {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var out []model.Point
	g.collectAllLocked(g.root, &out)
	return out
}

// ApplyPoint inserts a point during replay or rebuild without triggering an
// additional split policy check.
func (g *Grid) ApplyPoint(point model.Point) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	cell := g.locateLocked(point.X, point.Y)
	cell.Add(point)
	g.version++
	return nil
}

// QueryRange returns a result object for the rectangle. The result always
// carries a non-nil Points slice, even when the grid holds no data for the
// region.
func (g *Grid) QueryRange(rect model.Rect) *RangeResult {
	g.mu.RLock()
	defer g.mu.RUnlock()
	cells := g.cellsForQueryLocked(rect)
	points := make([]model.Point, 0, 8)
	for _, cell := range cells {
		points = append(points, cell.Scan(rect)...)
	}
	return &RangeResult{Points: points, Bounds: rect, CellCount: len(cells)}
}

// CellsForQuery returns the stable cells overlapping the rectangle for query
// routing. Split cells are descended so callers observe the current topology.
func (g *Grid) CellsForQuery(rect model.Rect) []*Cell {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.cellsForQueryLocked(rect)
}

// cellsForQueryLocked resolves the stable cells overlapping the rectangle,
// descending through split cells so queries observe the current topology.
func (g *Grid) cellsForQueryLocked(rect model.Rect) []*Cell {
	var out []*Cell
	g.collectLocked(g.root, rect, &out)
	return out
}

// RangeResult bundles a grid scan outcome with the query box for downstream
// filtering and statistics.
type RangeResult struct {
	Points    []model.Point
	Bounds    model.Rect
	CellCount int
}

// TileKey returns the stable cache key for a slippy tile coordinate.
func (g *Grid) TileKey(z, x, y int) string {
	return fmt.Sprintf("%d/%d/%d", z, x, y)
}

// TilesForRect maps a world rectangle onto the tile keys it intersects for
// the requested zoom levels.
func (g *Grid) TilesForRect(rect model.Rect, zooms []int) []string {
	var keys []string
	for _, zoom := range zooms {
		minX, maxX := tileRange(rect.MinX, rect.MaxX, zoom)
		minY, maxY := tileRange(rect.MinY, rect.MaxY, zoom)
		for tx := minX; tx <= maxX; tx++ {
			for ty := minY; ty <= maxY; ty++ {
				keys = append(keys, g.TileKey(zoom, tx, ty))
			}
		}
	}
	return keys
}

func tileRange(minV, maxV float64, zoom int) (int, int) {
	size := 1 << zoom
	clamp := func(value float64) float64 {
		if value < 0 {
			return 0
		}
		if value > 1 {
			return 1
		}
		return value
	}
	minF := clamp((minV + 180) / 360)
	maxF := clamp((maxV + 180) / 360)
	lo := int(minF * float64(size))
	hi := int(maxF * float64(size))
	if lo > hi {
		lo, hi = hi, lo
	}
	return lo, hi
}
