package write

import (
	"sync"

	"geoindex/internal/model"
)

// PolygonStore keeps the committed polygon geometries used for belonging
// checks. Geometry is only replaced after a successful validation.
type PolygonStore struct {
	mu   sync.RWMutex
	byID map[string]model.Polygon
}

// NewPolygonStore builds an empty store.
func NewPolygonStore() *PolygonStore {
	return &PolygonStore{byID: make(map[string]model.Polygon)}
}

// Put validates and commits a polygon.
func (s *PolygonStore) Put(polygon model.Polygon) error {
	if err := polygon.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byID[polygon.ID] = polygon
	return nil
}

// Get returns the committed polygon for an id.
func (s *PolygonStore) Get(id string) (model.Polygon, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	polygon, ok := s.byID[id]
	return polygon, ok
}

// All returns every committed polygon in stable order.
func (s *PolygonStore) All() []model.Polygon {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]string, 0, len(s.byID))
	for id := range s.byID {
		ids = append(ids, id)
	}
	sortStrings(ids)
	out := make([]model.Polygon, 0, len(ids))
	for _, id := range ids {
		out = append(out, s.byID[id])
	}
	return out
}

// Belongs reports whether a coordinate falls inside the committed geometry of
// the named polygon.
func (s *PolygonStore) Belongs(id string, x, y float64) bool {
	polygon, ok := s.Get(id)
	if !ok {
		return false
	}
	return polygon.Contains(x, y)
}

// Size returns the number of committed polygons.
func (s *PolygonStore) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.byID)
}

// UpdatePolygon validates and commits a new polygon geometry. Any validation
// failure is propagated to the caller and the previous geometry stays active.
func (s *Service) UpdatePolygon(polygon model.Polygon) error {
	if polygon.ID == "" {
		return ErrEmptyID
	}
	if err := model.ValidateGeometryUpdate(polygon); err != nil {
		return err
	}
	if err := s.polys.Put(polygon); err != nil {
		return err
	}
	if _, err := s.log.Append(model.NewPolygonOp(0, polygon)); err != nil {
		return err
	}
	s.invalidateRegion(polygon.MBR())
	return nil
}

// PolygonSummary lists the committed polygon ids joined by commas.
func (s *Service) PolygonSummary() string {
	polygons := s.polys.All()
	ids := make([]string, 0, len(polygons))
	for _, polygon := range polygons {
		ids = append(ids, polygon.ID)
	}
	return joinStrings(ids, ",")
}
