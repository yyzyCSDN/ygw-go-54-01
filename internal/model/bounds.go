package model

// GridBounds describes the extent of one grid cell in world coordinates.
type GridBounds struct {
	MinX, MinY, MaxX, MaxY float64
}

// Center returns the midpoint of the cell.
func (b GridBounds) Center() (x, y float64) {
	return (b.MinX + b.MaxX) / 2, (b.MinY + b.MaxY) / 2
}

// Contains reports whether the coordinate lies inside the cell bounds.
func (b GridBounds) Contains(x, y float64) bool {
	return x >= b.MinX && x <= b.MaxX && y >= b.MinY && y <= b.MaxY
}

// Width returns the cell width along the X axis.
func (b GridBounds) Width() float64 {
	return b.MaxX - b.MinX
}

// Height returns the cell height along the Y axis.
func (b GridBounds) Height() float64 {
	return b.MaxY - b.MinY
}

// Split quarters the bounds. The returned order is fixed: north-east,
// north-west, south-west and south-east, which keeps scan order stable.
func (b GridBounds) Split() [4]GridBounds {
	midX := (b.MinX + b.MaxX) / 2
	midY := (b.MinY + b.MaxY) / 2
	return [4]GridBounds{
		{MinX: midX, MinY: midY, MaxX: b.MaxX, MaxY: b.MaxY},
		{MinX: b.MinX, MinY: midY, MaxX: midX, MaxY: b.MaxY},
		{MinX: b.MinX, MinY: b.MinY, MaxX: midX, MaxY: midY},
		{MinX: midX, MinY: b.MinY, MaxX: b.MaxX, MaxY: midY},
	}
}

// Rect converts the cell bounds into the shared rectangle type.
func (b GridBounds) Rect() Rect {
	return Rect{MinX: b.MinX, MinY: b.MinY, MaxX: b.MaxX, MaxY: b.MaxY}
}
