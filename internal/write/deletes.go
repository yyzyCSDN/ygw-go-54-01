package write

import "geoindex/internal/model"

// DeletePoint removes a point from the grid index and the range tree. The
// deletion is logged first so a rebuild does not resurrect the point.
func (s *Service) DeletePoint(id string) error {
	if id == "" {
		return ErrEmptyID
	}
	if _, err := s.log.Append(model.NewDeleteOp(0, id)); err != nil {
		return err
	}
	cellID, removed := s.grid.RemoveIndexed(id)
	if !removed {
		return ErrPointNotFound
	}
	s.tree.Remove(id)
	s.invalidateCell(cellID)
	return nil
}

// DeletePoints removes a batch of points and reports how many succeeded.
func (s *Service) DeletePoints(ids []string) (int, error) {
	deleted := 0
	for _, id := range ids {
		if err := s.DeletePoint(id); err != nil {
			return deleted, err
		}
		deleted++
	}
	return deleted, nil
}
