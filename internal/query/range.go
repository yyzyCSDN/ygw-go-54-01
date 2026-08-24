package query

import "geoindex/internal/model"

// filterInclusive keeps points that lie on or inside the rectangle.
func (s *Service) filterInclusive(points []model.Point, rect model.Rect) []model.Point {
	out := points[:0]
	for _, point := range points {
		if rect.ContainsPoint(point) {
			out = append(out, point)
		}
	}
	return out
}

// deduplicate removes repeated ids keeping the first occurrence.
func deduplicate(points []model.Point) []model.Point {
	seen := make(map[string]struct{}, len(points))
	out := points[:0]
	for _, point := range points {
		if _, ok := seen[point.ID]; ok {
			continue
		}
		seen[point.ID] = struct{}{}
		out = append(out, point)
	}
	return out
}

// intersect keeps the grid candidates that the range tree still indexes.
func intersect(gridPoints, treePoints []model.Point) []model.Point {
	inTree := make(map[string]struct{}, len(treePoints))
	for _, point := range treePoints {
		inTree[point.ID] = struct{}{}
	}
	out := gridPoints[:0]
	for _, point := range gridPoints {
		if _, ok := inTree[point.ID]; ok {
			out = append(out, point)
		}
	}
	return out
}
