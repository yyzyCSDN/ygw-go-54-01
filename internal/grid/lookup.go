package grid

import (
	"geoindex/internal/model"
)

// Locate returns the deepest stable cell containing the coordinate. Split
// cells are descended so that callers always observe the current topology.
func (g *Grid) Locate(x, y float64) *Cell {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.locateLocked(x, y)
}

func (g *Grid) locateLocked(x, y float64) *Cell {
	cell := g.root
	for cell.State == CellSplit && len(cell.Children) > 0 {
		next := (*Cell)(nil)
		for _, child := range cell.Children {
			if child.Bounds.Contains(x, y) {
				next = child
				break
			}
		}
		if next == nil {
			break
		}
		cell = next
	}
	return cell
}

// CellsForRect returns the stable cells overlapping the rectangle in a
// deterministic depth-first order.
func (g *Grid) CellsForRect(rect model.Rect) []*Cell {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.cellsForRectLocked(rect)
}

func (g *Grid) cellsForRectLocked(rect model.Rect) []*Cell {
	var out []*Cell
	g.collectLocked(g.root, rect, &out)
	return out
}

func (g *Grid) collectLocked(cell *Cell, rect model.Rect, out *[]*Cell) {
	if !cell.Bounds.Rect().Overlaps(rect) {
		return
	}
	if cell.State == CellSplit {
		for _, child := range cell.Children {
			g.collectLocked(child, rect, out)
		}
		return
	}
	*out = append(*out, cell)
}

func (g *Grid) collectAllLocked(cell *Cell, out *[]model.Point) {
	if cell.State == CellSplit {
		for _, child := range cell.Children {
			g.collectAllLocked(child, out)
		}
		return
	}
	cell.ForEachPoint(func(point model.Point) bool {
		*out = append(*out, point)
		return true
	})
}
