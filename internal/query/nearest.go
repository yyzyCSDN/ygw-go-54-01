package query

import (
	"sort"

	"geoindex/internal/model"
)

// DefaultSearchMargin is the initial expansion applied to the nearest
// candidate window.
const DefaultSearchMargin = 10.0

// nearestAnchor is the query origin for a nearest-neighbour lookup.
type nearestAnchor struct {
	x, y float64
}

// distanceSquared returns the squared euclidean distance from the anchor to a
// candidate point. Squared distance preserves the ordering of the true
// distance while avoiding the cost of the square root.
func (a nearestAnchor) distanceSquared(p model.Point) float64 {
	dx := a.x - p.X
	dy := a.y - p.Y
	return dx*dx + dy*dy
}

// Nearest returns up to k points closest to the anchor coordinate. The
// candidate set is merged across every overlapping grid cell and reordered by
// distance before truncation.
func (s *Service) Nearest(x, y float64, k int) ([]model.Point, error) {
	if k <= 0 {
		k = 1
	}
	anchor := nearestAnchor{x: x, y: y}
	rect := s.searchWindow(x, y, k)
	cells := s.grid.CellsForQuery(rect)
	candidates := make([]model.Point, 0, 8)
	for _, cell := range cells {
		candidates = append(candidates, cell.Scan(rect)...)
	}
	if len(candidates) < k {
		expanded := rect.Expand(DefaultSearchMargin)
		cells = s.grid.CellsForQuery(expanded)
		candidates = candidates[:0]
		for _, cell := range cells {
			candidates = append(candidates, cell.Scan(expanded)...)
		}
	}
	// Candidates are merged across every overlapping cell, so their order
	// reflects cell and insertion order rather than distance to the anchor.
	// Re-sort by distance before truncation so a farther point from an
	// earlier cell cannot push a nearer point beyond the k-th slot. The
	// stable sort keeps the existing cell/insertion order for equidistant
	// points so the result stays deterministic.
	sort.SliceStable(candidates, func(i, j int) bool {
		return anchor.distanceSquared(candidates[i]) < anchor.distanceSquared(candidates[j])
	})
	if len(candidates) > k {
		candidates = candidates[:k]
	}
	s.stats.ObserveNearest()
	s.stats.ObserveServed(len(candidates))
	return deduplicate(candidates), nil
}

// searchWindow builds the initial candidate rectangle around the anchor. The
// size grows with the requested result count so dense grids can satisfy
// larger k values without an immediate second pass.
func (s *Service) searchWindow(x, y float64, k int) model.Rect {
	margin := DefaultSearchMargin
	if k > 4 {
		margin *= float64(k) / 4
	}
	return model.Rect{
		MinX: x - margin,
		MinY: y - margin,
		MaxX: x + margin,
		MaxY: y + margin,
	}
}

// DistanceTo returns the euclidean distance between two coordinates.
func DistanceTo(px, py, qx, qy float64) float64 {
	dx := px - qx
	dy := py - qy
	return sqrt(dx*dx + dy*dy)
}

func sqrt(value float64) float64 {
	if value <= 0 {
		return 0
	}
	estimate := value
	for i := 0; i < 40; i++ {
		next := (estimate + value/estimate) / 2
		if next == estimate {
			break
		}
		estimate = next
	}
	return estimate
}
