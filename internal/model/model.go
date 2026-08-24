package model

import "errors"

// ErrPolygonOutOfExtent reports geometry that falls outside the world extent.
var ErrPolygonOutOfExtent = errors.New("polygon extends outside the world extent")

// WorldExtent returns the coordinate bounds the service indexes.
func WorldExtent() Rect {
	return Rect{MinX: -180, MinY: -90, MaxX: 180, MaxY: 90}
}

// ValidateGeometryUpdate checks that a polygon update can be committed: the
// vertex ring must be valid and its bounding box must stay inside the world
// extent. Write failures are propagated instead of being swallowed.
func ValidateGeometryUpdate(polygon Polygon) error {
	if err := polygon.Validate(); err != nil {
		return err
	}
	box := polygon.MBR()
	extent := WorldExtent()
	if box.MinX < extent.MinX || box.MaxX > extent.MaxX || box.MinY < extent.MinY || box.MaxY > extent.MaxY {
		return ErrPolygonOutOfExtent
	}
	return nil
}
