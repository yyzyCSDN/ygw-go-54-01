package write

import "geoindex/internal/model"

// AddPoint validates and indexes a new point. The mutation is appended to the
// write-ahead log, applied to the grid index and the range tree, and the
// affected tiles are invalidated before the call returns.
func (s *Service) AddPoint(point model.Point) error {
	if point.ID == "" {
		return ErrEmptyID
	}
	stored := point.Clone()
	op := model.NewAddOp(0, stored)
	if _, err := s.log.Append(op); err != nil {
		return err
	}
	s.tree.Insert(stored)
	s.invalidateRegion(model.PointRect(stored))
	return nil
}

// AddPoints indexes a batch of points and reports how many succeeded.
func (s *Service) AddPoints(points []model.Point) (int, error) {
	added := 0
	for _, point := range points {
		if err := s.AddPoint(point); err != nil {
			return added, err
		}
		added++
	}
	return added, nil
}

// ReindexPoint re-applies an existing point to the index. It is used during
// recovery after a WAL replay.
func (s *Service) ReindexPoint(point model.Point) error {
	if point.ID == "" {
		return ErrEmptyID
	}
	_, err := s.grid.Put(point.Clone())
	return err
}
