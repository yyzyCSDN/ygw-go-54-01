package model

import (
	"errors"
	"fmt"
)

// ErrInvalidPolygon reports a vertex ring that cannot be used for containment
// tests, such as rings with fewer than three distinct vertices.
var ErrInvalidPolygon = errors.New("polygon needs at least three distinct vertices")

// Polygon is an ordered vertex ring with a stable identifier and display name.
// Polygons are used as administrative or district boundaries for point
// belonging checks.
type Polygon struct {
	ID       string
	Name     string
	Vertices []Point
	Tags     []string
}

// Validate checks that the ring can be used for containment tests.
func (p Polygon) Validate() error {
	if len(p.Vertices) < 3 {
		return ErrInvalidPolygon
	}
	seen := make(map[string]struct{}, len(p.Vertices))
	for _, vertex := range p.Vertices {
		key := fmt.Sprintf("%.6f,%.6f", vertex.X, vertex.Y)
		seen[key] = struct{}{}
	}
	if len(seen) < 3 {
		return ErrInvalidPolygon
	}
	return nil
}

// MBR returns the bounding box of the vertex ring.
func (p Polygon) MBR() Rect {
	if len(p.Vertices) == 0 {
		return Rect{}
	}
	box := PointRect(p.Vertices[0])
	for _, vertex := range p.Vertices[1:] {
		box = box.Union(PointRect(vertex))
	}
	return box
}

// Contains implements ray casting against the ring. Boundary points are
// considered inside, matching the rectangle convention.
func (p Polygon) Contains(x, y float64) bool {
	vertices := p.Vertices
	if len(vertices) < 3 {
		return false
	}
	inside := false
	j := len(vertices) - 1
	for i := 0; i < len(vertices); i++ {
		a := vertices[i]
		b := vertices[j]
		if (a.Y > y) != (b.Y > y) {
			intersectionX := (b.X-a.X)*(y-a.Y)/(b.Y-a.Y) + a.X
			if x <= intersectionX {
				inside = !inside
			}
		}
		j = i
	}
	return inside
}

// VertexCount returns the number of ring vertices.
func (p Polygon) VertexCount() int {
	return len(p.Vertices)
}
