package query

import (
	"geoindex/internal/grid"
	"geoindex/internal/model"
	"geoindex/internal/rtree"
)

// Service answers spatial queries against the grid index and the range tree.
type Service struct {
	grid  *grid.Grid
	tree  *rtree.RTree
	stats *Stats
}

// NewService wires the query service onto the shared indexes.
func NewService(g *grid.Grid, tree *rtree.RTree) *Service {
	return &Service{
		grid:  g,
		tree:  tree,
		stats: NewStats(),
	}
}

// Stats returns the query statistics collector.
func (s *Service) Stats() *Stats {
	return s.stats
}

// Grid exposes the underlying grid for diagnostics.
func (s *Service) Grid() *grid.Grid {
	return s.grid
}

// Range returns every indexed point inside the rectangle. The boundary is
// inclusive: points lying exactly on the rectangle edge are returned.
func (s *Service) Range(rect model.Rect) ([]model.Point, error) {
	cells := s.grid.CellsForQuery(rect)
	points := make([]model.Point, 0, 8)
	for _, cell := range cells {
		for _, id := range cell.PointIDs() {
			point := cell.Points[id]
			if rect.ContainsPoint(point) {
				points = append(points, point)
			}
		}
	}
	s.stats.ObserveRange(rect)
	return deduplicate(points), nil
}

// Count returns the number of points inside the rectangle.
func (s *Service) Count(rect model.Rect) (int, error) {
	points, err := s.Range(rect)
	if err != nil {
		return 0, err
	}
	return len(points), nil
}

// Exists reports whether at least one point is inside the rectangle.
func (s *Service) Exists(rect model.Rect) (bool, error) {
	count, err := s.Count(rect)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
