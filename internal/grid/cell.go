package grid

import "geoindex/internal/model"

// CellState models the lifecycle of a grid cell. A cell starts stable, moves
// to splitting while its points are migrated, and finally becomes split with
// four children. The transition is driven by the grid under its write lock.
type CellState uint8

const (
	CellStable CellState = iota
	CellSplitting
	CellSplit
)

// String returns a stable name for the state.
func (s CellState) String() string {
	switch s {
	case CellStable:
		return "stable"
	case CellSplitting:
		return "splitting"
	case CellSplit:
		return "split"
	default:
		return "unknown"
	}
}

// Cell is one node of the grid tree. Stable cells own a point map; split
// cells own four children and keep an empty point map.
type Cell struct {
	ID       string
	Bounds   model.GridBounds
	Points   map[string]model.Point
	order    []string
	State    CellState
	Children []*Cell
	Depth    int
}

func newCell(id string, bounds model.GridBounds, depth int) *Cell {
	return &Cell{
		ID:     id,
		Bounds: bounds,
		Points: make(map[string]model.Point),
		State:  CellStable,
		Depth:  depth,
	}
}

// Add inserts a point into the cell and records its insertion order.
func (c *Cell) Add(point model.Point) {
	if _, exists := c.Points[point.ID]; !exists {
		c.order = append(c.order, point.ID)
	}
	c.Points[point.ID] = point
}

// Remove deletes a point and returns whether it existed.
func (c *Cell) Remove(id string) bool {
	if _, exists := c.Points[id]; !exists {
		return false
	}
	delete(c.Points, id)
	for index, current := range c.order {
		if current == id {
			c.order = append(c.order[:index], c.order[index+1:]...)
			break
		}
	}
	return true
}

// Count returns the number of points owned directly by this cell.
func (c *Cell) Count() int {
	return len(c.Points)
}

// PointIDs returns the direct point identifiers in insertion order.
func (c *Cell) PointIDs() []string {
	out := make([]string, 0, len(c.order))
	out = append(out, c.order...)
	return out
}

// Scan returns points inside the rectangle. Split cells recurse into their
// children so that queries always observe the current cell topology. The
// rectangle boundary is inclusive.
func (c *Cell) Scan(rect model.Rect) []model.Point {
	if c.State == CellSplit {
		var out []model.Point
		for _, child := range c.Children {
			if child.Bounds.Rect().Overlaps(rect) {
				out = append(out, child.Scan(rect)...)
			}
		}
		return out
	}
	out := make([]model.Point, 0, len(c.order))
	for _, id := range c.order {
		point := c.Points[id]
		if rect.ContainsPoint(point) {
			out = append(out, point)
		}
	}
	return out
}

// ForEachPoint visits every direct point in stable order.
func (c *Cell) ForEachPoint(fn func(model.Point) bool) {
	for _, id := range c.order {
		if !fn(c.Points[id]) {
			return
		}
	}
}
