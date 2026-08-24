package write

import (
	"errors"

	"geoindex/internal/grid"
	"geoindex/internal/model"
	"geoindex/internal/rtree"
	"geoindex/internal/tile"
	"geoindex/internal/wal"
)

// ErrEmptyID is returned when an operation lacks an identifier.
var ErrEmptyID = errors.New("spatial object id is empty")

// ErrPointNotFound is returned when deleting a point that is not indexed.
var ErrPointNotFound = errors.New("point not found")

// Service applies spatial changes to the grid index, the range tree, the
// write-ahead log and the tile cache. Every public mutation is recorded in
// the log first so rebuilds can fold in writes that arrive mid-rebuild.
type Service struct {
	grid  *grid.Grid
	tree  *rtree.RTree
	log   *wal.Log
	tiles *tile.Renderer
	polys *PolygonStore
}

// Options controls optional service wiring.
type Options struct {
	Tiles *tile.Renderer
}

// NewService wires the write service onto the shared indexes.
func NewService(g *grid.Grid, tree *rtree.RTree, log *wal.Log, opts Options) *Service {
	return &Service{
		grid:  g,
		tree:  tree,
		log:   log,
		tiles: opts.Tiles,
		polys: NewPolygonStore(),
	}
}

// Grid exposes the underlying index for read-only consumers.
func (s *Service) Grid() *grid.Grid {
	return s.grid
}

// Tree exposes the underlying range tree.
func (s *Service) Tree() *rtree.RTree {
	return s.tree
}

// Log exposes the write-ahead log for rebuilds and diagnostics.
func (s *Service) Log() *wal.Log {
	return s.log
}

// Polygons exposes the polygon store used by belonging checks.
func (s *Service) Polygons() *PolygonStore {
	return s.polys
}

// invalidateRegion clears tile cache entries touched by a world rectangle.
func (s *Service) invalidateRegion(box model.Rect) {
	if s.tiles == nil {
		return
	}
	s.tiles.InvalidateRegion(box)
}

// invalidateCell clears tiles touched by a named cell.
func (s *Service) invalidateCell(cellID string) {
	bounds, ok := s.grid.CellBounds(cellID)
	if !ok || s.tiles == nil {
		return
	}
	s.tiles.InvalidateRegion(bounds)
}
