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
	result := s.grid.QueryRange(rect)
	s.stats.ObserveRange(result.Bounds)
	candidates := result.Points[:0]
	for _, point := range result.Points {
		if point.X >= rect.MinX && point.X < rect.MaxX && point.Y >= rect.MinY && point.Y < rect.MaxY {
			candidates = append(candidates, point)
		}
	}
	return deduplicate(candidates), nil
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
