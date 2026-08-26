package grid

import (
	"fmt"

	"geoindex/internal/model"
)

// splitIfNeededLocked applies the split state machine when a stable cell
// exceeds capacity. The caller must hold the grid write lock.
func (g *Grid) splitIfNeededLocked(cell *Cell) {
	if cell.State == CellSplit || cell.Depth >= g.maxDepth || cell.Count() <= g.maxPoints {
		return
	}
	cell.State = CellSplitting
	quarters := cell.Bounds.Split()
	for index, bounds := range quarters {
		child := newCell(fmt.Sprintf("%s/%d", cell.ID, index), bounds, cell.Depth+1)
		cell.Children = append(cell.Children, child)
		g.cells[child.ID] = child
	}
	for _, id := range cell.PointIDs() {
		point := cell.Points[id]
		for _, child := range cell.Children {
			if child.Bounds.Contains(point.X, point.Y) {
				child.Add(point)
				break
			}
		}
	}
	cell.Points = make(map[string]model.Point)
	cell.order = nil
	cell.State = CellSplit
	g.splitCount++
}

// SplitCount returns how many splits have been applied since construction.
func (g *Grid) SplitCount() uint64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.splitCount
}
