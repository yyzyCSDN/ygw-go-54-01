package model

// WorldExtent returns the coordinate bounds the service indexes.
func WorldExtent() Rect {
	return Rect{MinX: -180, MinY: -90, MaxX: 180, MaxY: 90}
}
