package grid

import (
	"encoding/binary"
	"math"
	"strconv"

	"github.com/cespare/xxhash/v2"

	"geoindex/internal/model"
)

// Metrics is a point-in-time snapshot of grid structure used by diagnostics.
type Metrics struct {
	CellCount  int
	PointCount int
	MaxDepth   int
	SplitCount uint64
	Version    uint64
}

// Metrics returns the current structural snapshot.
func (g *Grid) Metrics() Metrics {
	g.mu.RLock()
	defer g.mu.RUnlock()
	metrics := Metrics{
		CellCount:  len(g.cells),
		MaxDepth:   g.maxDepth,
		SplitCount: g.splitCount,
		Version:    g.version,
	}
	var countPoints func(*Cell)
	countPoints = func(cell *Cell) {
		if cell.State == CellSplit {
			for _, child := range cell.Children {
				countPoints(child)
			}
			return
		}
		metrics.PointCount += cell.Count()
	}
	countPoints(g.root)
	return metrics
}

// SpatialKey hashes a coordinate pair into a stable hexadecimal bucket key.
// The key is used to group cache invalidations and tile fingerprints.
func SpatialKey(x, y float64) string {
	var buffer [16]byte
	binary.LittleEndian.PutUint64(buffer[0:8], math.Float64bits(x))
	binary.LittleEndian.PutUint64(buffer[8:16], math.Float64bits(y))
	return strconv.FormatUint(xxhash.Sum64(buffer[:]), 16)
}

// RegionKey hashes a rectangle into a stable key identifying the region.
func RegionKey(rect model.Rect) string {
	var buffer [32]byte
	binary.LittleEndian.PutUint64(buffer[0:8], math.Float64bits(rect.MinX))
	binary.LittleEndian.PutUint64(buffer[8:16], math.Float64bits(rect.MinY))
	binary.LittleEndian.PutUint64(buffer[16:24], math.Float64bits(rect.MaxX))
	binary.LittleEndian.PutUint64(buffer[24:32], math.Float64bits(rect.MaxY))
	return strconv.FormatUint(xxhash.Sum64(buffer[:]), 16)
}
