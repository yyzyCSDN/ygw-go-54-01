package model

// Rect is an axis aligned bounding box. The boundary is inclusive: a point
// lying exactly on MinX, MaxX, MinY or MaxY is considered inside the box.
type Rect struct {
	MinX, MinY, MaxX, MaxY float64
}

// Contains reports whether the coordinate lies on or inside the rectangle.
func (r Rect) Contains(x, y float64) bool {
	return x >= r.MinX && x <= r.MaxX && y >= r.MinY && y <= r.MaxY
}

// ContainsPoint reports whether the point lies on or inside the rectangle.
func (r Rect) ContainsPoint(p Point) bool {
	return r.Contains(p.X, p.Y)
}

// Overlaps reports whether two rectangles share at least one point. Touching
// edges count as overlapping, matching the inclusive boundary convention.
func (r Rect) Overlaps(o Rect) bool {
	return r.MinX <= o.MaxX && r.MaxX >= o.MinX && r.MinY <= o.MaxY && r.MaxY >= o.MinY
}

// Union returns the smallest rectangle that fully contains both rectangles.
func (r Rect) Union(o Rect) Rect {
	return Rect{
		MinX: min64(r.MinX, o.MinX),
		MinY: min64(r.MinY, o.MinY),
		MaxX: max64(r.MaxX, o.MaxX),
		MaxY: max64(r.MaxY, o.MaxY),
	}
}

// Area returns the rectangle area. Empty or inverted boxes yield zero.
func (r Rect) Area() float64 {
	width := r.MaxX - r.MinX
	height := r.MaxY - r.MinY
	if width < 0 {
		width = -width
	}
	if height < 0 {
		height = -height
	}
	return width * height
}

// Expand grows the rectangle by margin on every side.
func (r Rect) Expand(margin float64) Rect {
	return Rect{
		MinX: r.MinX - margin,
		MinY: r.MinY - margin,
		MaxX: r.MaxX + margin,
		MaxY: r.MaxY + margin,
	}
}

// Center returns the midpoint of the rectangle.
func (r Rect) Center() (x, y float64) {
	return (r.MinX + r.MaxX) / 2, (r.MinY + r.MaxY) / 2
}

// PointRect builds the degenerate rectangle surrounding a single point.
func PointRect(p Point) Rect {
	return Rect{MinX: p.X, MinY: p.Y, MaxX: p.X, MaxY: p.Y}
}

func min64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
