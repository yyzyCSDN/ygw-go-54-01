package model

import (
	"sort"
	"strings"
)

// Point is a spatial point with a stable identifier, coordinates and a small
// set of caller supplied attributes. Points are the unit of data indexed by
// the grid and the range tree.
type Point struct {
	ID       string
	X        float64
	Y        float64
	Category string
	Weight   float64
	Meta     map[string]string
}

// NewPoint builds a point with default weight and no attributes.
func NewPoint(id string, x, y float64) Point {
	return Point{ID: id, X: x, Y: y, Weight: 1}
}

// Clone returns an independent copy of the point. The attribute map is copied
// so that later mutations made by callers cannot leak into the index.
func (p Point) Clone() Point {
	meta := make(map[string]string, len(p.Meta))
	for key, value := range p.Meta {
		meta[key] = value
	}
	return Point{
		ID:       p.ID,
		X:        p.X,
		Y:        p.Y,
		Category: p.Category,
		Weight:   p.Weight,
		Meta:     meta,
	}
}

// Tags returns the attribute keys in stable order for deterministic output.
func (p Point) Tags() []string {
	keys := make([]string, 0, len(p.Meta))
	for key := range p.Meta {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Summary renders a one-line description used by tiles and diagnostics. The
// output only contains the identity, category and sorted attribute pairs.
func (p Point) Summary() string {
	var builder strings.Builder
	builder.WriteString(p.ID)
	builder.WriteByte(':')
	builder.WriteString(p.Category)
	for _, key := range p.Tags() {
		builder.WriteByte('|')
		builder.WriteString(key)
		builder.WriteByte('=')
		builder.WriteString(p.Meta[key])
	}
	return builder.String()
}

// DistanceSquared returns the squared euclidean distance to another point.
func (p Point) DistanceSquared(o Point) float64 {
	dx := p.X - o.X
	dy := p.Y - o.Y
	return dx*dx + dy*dy
}
