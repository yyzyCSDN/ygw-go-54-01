package query

import (
	"sync"
	"time"

	"geoindex/internal/model"
)

// Stats accumulates query traffic for diagnostics and the control page.
type Stats struct {
	mu            sync.Mutex
	rangeCount    uint64
	nearestCount  uint64
	pointsServed  uint64
	lastRangeArea float64
	lastQueryAt   time.Time
}

// NewStats builds an empty collector.
func NewStats() *Stats {
	return &Stats{lastQueryAt: time.Now()}
}

// ObserveRange records one range query.
func (s *Stats) ObserveRange(rect model.Rect) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rangeCount++
	s.lastRangeArea = rect.Area()
	s.lastQueryAt = time.Now()
}

// ObserveNearest records one nearest query.
func (s *Stats) ObserveNearest() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nearestCount++
	s.lastQueryAt = time.Now()
}

// ObserveServed records how many points a query returned.
func (s *Stats) ObserveServed(count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pointsServed += uint64(count)
}

// Snapshot is a point-in-time view of the counters.
type Snapshot struct {
	RangeCount    uint64
	NearestCount  uint64
	PointsServed  uint64
	LastRangeArea float64
	LastQueryAt   time.Time
}

// Snapshot returns the current counters.
func (s *Stats) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return Snapshot{
		RangeCount:    s.rangeCount,
		NearestCount:  s.nearestCount,
		PointsServed:  s.pointsServed,
		LastRangeArea: s.lastRangeArea,
		LastQueryAt:   s.lastQueryAt,
	}
}
